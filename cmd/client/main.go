package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"os"
	"strings"

	"sanin7k/tcp-echo/internal/netutil"
)

func main() {
	conn, err := net.Dial("tcp", "127.0.0.1:9090")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	headerBuf := make([]byte, 4)
	r := bufio.NewReader(os.Stdin)

	const maxEcho = 20 * 1024 * 1024
	for {
		fmt.Print("Enter message: ")
		line, err := r.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			panic(err)
		}
		line = strings.TrimRight(line, "\r\n")

		binary.BigEndian.PutUint32(headerBuf, uint32(len(line)))
		if err := netutil.WriteAll(conn, headerBuf); err != nil {
			fmt.Println("Error writing header")
			return
		}
		if err := netutil.WriteAll(conn, []byte(line)); err != nil {
			fmt.Println("Error writing payload")
			return
		}

		if _, err = io.ReadFull(conn, headerBuf); err != nil {
			fmt.Println("Error receiving echo header")
			return
		}
		echoSize := binary.BigEndian.Uint32(headerBuf)
		if echoSize > uint32(maxEcho) {
			fmt.Println("echo frame is too large", echoSize)
			return
		}
		buf := make([]byte, echoSize)
		if _, err = io.ReadFull(conn, buf); err != nil {
			fmt.Println("Error receiving echo")
			return
		}
		fmt.Println(string(buf))
	}
}
