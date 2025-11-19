package servers

import (
	"fmt"
	"net"
)

func UdpServer() {
	fmt.Println("Starting server on UDP port 53...")

	addr, err := net.ResolveUDPAddr("udp", ":53")
	if err != nil {
		fmt.Println("Error resolving address:", err)
		return
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}
	defer conn.Close()

	fmt.Println("UDP server is listening on port 53")

	buffer := make([]byte, 4096)

	for {
		n, client, err := conn.ReadFrom(buffer)
		if err != nil {
			fmt.Println("Error reading from connection:", err)
			continue
		}

		addr := client.String()
		fmt.Printf("Received %d bytes from %s\n", n, addr)

		// will pass later to resolver
		//message := buffer[:n]
		fmt.Printf("Received from %s\n", addr)

		res := []byte("Hello Bitch")

		_, err = conn.WriteTo(res, client)
		if err != nil {
			fmt.Println("Error writing to connection:", err)
			continue
		}
	}
}
