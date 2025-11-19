package servers

import (
	"context"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"aifia.com/dns-server/resolver"
)

func TcpServer(ctx context.Context, wg *sync.WaitGroup) error {
	defer wg.Done()

	ln, err := net.Listen("tcp", ":8081")
	if err != nil {
		fmt.Println("Error starting server:", err)
		return err
	}
	defer ln.Close()

	fmt.Println("TCP server is listening on port 8081")

	go func() {
		<-ctx.Done()
		fmt.Printf("TCP server: received shutdown signal, closing listener\n")
		ln.Close()
	}()

	for {
		if tcpListener, ok := ln.(*net.TCPListener); ok {
			tcpListener.SetDeadline(time.Now().Add(1 * time.Second))
		}

		conn, err := ln.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				fmt.Printf("TCP server: shutting down\n")
				return nil
			default:
				if ne, ok := err.(net.Error); ok && ne.Timeout() {
					continue
				}
				fmt.Println("Error accepting connection:", err)
				continue
			}
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

	dnsmessage, err := resolver.Parser(buffer)
	if err != nil {
		fmt.Printf("Error parsing DNS message from %s: %v\n", conn.RemoteAddr(), err)
		return
	}

	fmt.Printf("Parsed DNS message from %s: %+v\n", conn.RemoteAddr(), dnsmessage)

	res := []byte("Hello Bitch")
	resLen := uint16(len(res))

	resLenBytes := []byte{byte(resLen >> 8), byte(resLen & 0xff)}

	_, err = conn.Write(resLenBytes)
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
