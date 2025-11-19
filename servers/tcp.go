package servers

import (
	"fmt"
	"io"
	"net"
	"time"
)

func TcpServer() {
	fmt.Println("Starting server on TCP port 53...")
	ln, err := net.Listen("tcp", ":53")
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}
	defer ln.Close()

	fmt.Println("TCP server is listening on port 53")

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	fmt.Printf("Connection established from %s\n", conn.RemoteAddr().String())
	conn.SetDeadline(time.Now().Add(10 * time.Minute))

	msgLen := make([]byte, 2)
	totalRead := 0
	for totalRead < 2 {
		n, e := conn.Read(msgLen[totalRead:])
		if e != nil {
			if e == io.EOF {
				fmt.Printf("Client %s disconnected while sending length prefix\n", conn.RemoteAddr())
			} else {
				fmt.Println("Error reading length prefix:", e)
			}
			return
		}
		totalRead += n
	}

	length := (uint16(msgLen[0]) << 8) | uint16(msgLen[1])

	if length == 0 {
		fmt.Printf("Client %s sent zero-length message, closing connection\n", conn.RemoteAddr())
		return
	}

	buffer := make([]byte, length)
	totalRead = 0
	for totalRead < int(length) {
		n, e := conn.Read(buffer[totalRead:])
		if e != nil {
			if e == io.EOF {
				fmt.Printf("Client %s disconnected mid-message after %d bytes\n", conn.RemoteAddr(), totalRead)
			} else {
				fmt.Println("Error reading DNS payload:", e)
			}
			return
		}
		totalRead += n
	}

	fmt.Printf("Received complete DNS message of %d bytes from %s\n", length, conn.RemoteAddr())

	res := []byte("Hello Bitch")
	resLen := uint16(len(res))

	resLenBytes := []byte{byte(resLen >> 8), byte(resLen & 0xff)}

	_, err := conn.Write(resLenBytes)
	if err != nil {
		fmt.Println("Error writing to connection:", err)
		return
	}

	_, err = conn.Write(res)
	if err != nil {
		fmt.Println("Error writing to connection:", err)
		return
	}

	fmt.Printf("Response sent to %s (%d bytes)\n", conn.RemoteAddr(), resLen)

}
