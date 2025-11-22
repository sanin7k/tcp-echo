package netutil

import "net"

func WriteAll(conn net.Conn, b []byte) error {
	written := 0
	for written < len(b) {
		n, err := conn.Write(b[written:])
		if err != nil {
			return err
		}
		written += n
	}
	return nil
}
