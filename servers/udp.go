package servers

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"aifia.com/dns-server/resolver"
)

func UdpServer(ctx context.Context, wg *sync.WaitGroup) error {
	defer wg.Done()

	addr, err := net.ResolveUDPAddr("udp", ":8080")
	if err != nil {
		fmt.Println("Error resolving address:", err)
		return err
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		fmt.Println("Error starting server:", err)
		return err
	}
	defer conn.Close()

	fmt.Println("UDP server is listening on port 8080")

	go func() {
		<-ctx.Done()
		fmt.Printf("UDP server: received shutdown signal, closing listener\n")
		conn.Close()
	}()

	buffer := make([]byte, 4096)

	for {
		conn.SetReadDeadline(time.Now().Add(1 * time.Second))

		n, client, err := conn.ReadFrom(buffer)
		if err != nil {
			select {
			case <-ctx.Done():
				fmt.Println("UDP server: shutting down")
				return nil
			default:
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					continue
				}
				fmt.Println("Error reading from connection:", err)
				continue
			}
		}

		addr := client.String()
		fmt.Printf("Received %d bytes from %s\n", n, addr)

		dnsmessage, err := resolver.Parser(buffer[:n])
		if err != nil {
			fmt.Printf("Error parsing DNS message from %s: %v\n", addr, err)
			continue
		}
		fmt.Printf("Parsed DNS message from %s: %+v\n", addr, dnsmessage)

		res := []byte("Hello Bitch")

		_, err = conn.WriteTo(res, client)
		if err != nil {
			fmt.Println("Error writing to connection:", err)
			continue
		}
	}
}
