package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"sanin7k/tcp-echo/internal/netutil"
)

func handleConn(conn net.Conn) {
	defer conn.Close()
	fmt.Println("Connected: ", conn.RemoteAddr())
	headerBuf := make([]byte, 4)

	const maxFrame = 20 * 1024 * 1024

	for {
		if _, err := io.ReadFull(conn, headerBuf); err != nil {
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				fmt.Println("Client closed connection")
				return
			}
			fmt.Println("header read err: ", err)
			return
		}

		size := binary.BigEndian.Uint32(headerBuf)
		if size > uint32(maxFrame) {
			fmt.Println("frame too large:", size)
			return
		}

		buf := make([]byte, size)
		if _, err := io.ReadFull(conn, buf); err != nil {
			fmt.Println("payload read err: ", err)
			return
		}
		fmt.Println("Received: ", string(buf))

		payload := append([]byte("ECHO:"), buf...)
		resHeader := make([]byte, 4)
		resSize := uint32(len(payload))
		if resSize > uint32(maxFrame) {
			fmt.Println("response frame too large:", resSize)
			return
		}
		binary.BigEndian.PutUint32(resHeader, resSize)

		if err := netutil.WriteAll(conn, resHeader); err != nil {
			fmt.Println("Error writing echo header: ", err)
			return
		}
		if err := netutil.WriteAll(conn, payload); err != nil {
			fmt.Println("Error writing echo payload: ", err)
			return
		}
	}
}

func main() {
	ln, err := net.Listen("tcp", ":9090")
	if err != nil {
		panic(err)
	}
	fmt.Println("Listening on :9090")

	const maxConcurrency = 100
	limiter := make(chan struct{}, maxConcurrency)
	var wg sync.WaitGroup

	sigch := make(chan os.Signal, 1)
	signal.Notify(sigch, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigch
		fmt.Println("\nSignal received - shutting down listener")
		ln.Close()
	}()

	for {
		conn, err := ln.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				break
			}
			fmt.Println("Accept err: ", err)
			continue
		}
		limiter <- struct{}{}
		wg.Add(1)
		go func(c net.Conn) {
			defer func() {
				<-limiter
				wg.Done()
			}()
			handleConn(c)
		}(conn)
	}
	fmt.Println("Waiting for active connections to finish...")
	wg.Wait()
	fmt.Println("Shutdown complete.")
}
