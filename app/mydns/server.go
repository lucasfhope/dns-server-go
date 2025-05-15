package mydns

import (
	"fmt"
	"net"
	"time"
)

func StartDNSServer(resolver string) {

	udpAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:2053")
	if err != nil {
		fmt.Println("Failed to resolve UDP address:", err)
		return
	}

	err = testResolver(resolver)
	if err != nil {
		fmt.Print("[Failed to connect to resolver at ", resolver, "]\n")
		fmt.Println(err)
		return
	}

	err = listenAndRespond(udpAddr, resolver)
	if err != nil {
		fmt.Println("[Failed to start DNS server]")
		fmt.Println(err)
		return
	}
}

func listenAndRespond(udpAddr *net.UDPAddr, resolver string) error {
	udpConn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return err
	}
	if resolver != "" {
		fmt.Print("[DNS server that will forward to ", resolver, " listening on ", udpAddr, "]\n")
	} else {
		fmt.Print("[DNS server listening on ", udpAddr, "]\n")
	}
	defer udpConn.Close()

	buf := make([]byte, 512)
	for {

		size, source, err := udpConn.ReadFromUDP(buf)
		if err != nil {
			fmt.Println("Error receiving data:", err)
			break
		}
		packet := make([]byte, size)
		copy(packet, buf[:size])

		go func(packet []byte, source *net.UDPAddr) {
			fmt.Printf("Received %d bytes from %s\n", len(packet), source)

			var response []byte
			if resolver != "" {
				fmt.Printf("Forwarding query to resolver: %s\n", resolver)
				response, err = forwardQueryToResolver(packet, resolver)
				if err != nil {
					fmt.Printf("Failed to forward query to resolver: %v\n", err)
					return
				}
			} else {

				recievedMessage, err := ParseDNSMessage(packet)
				if err != nil {
					fmt.Printf("Failed to parse DNS query from %s\n", source)
					fmt.Println(err)
					return
				}
				for _, question := range recievedMessage.Questions {
					fmt.Printf("Parsed DNS request from %s for %s\n", source, question.QNAME)
				}

				response = BuildDNSResponse(recievedMessage)
			}
			_, err = udpConn.WriteToUDP(response, source)
			if err != nil {
				fmt.Println("Failed to send response:", err)
			}
			fmt.Printf("Sent response to %s\n", source)
		}(packet, source)
	}
	return nil
}

func forwardQueryToResolver(query []byte, resolver string) ([]byte, error) {
	conn, err := net.Dial("udp", resolver)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	timeout := 5 * time.Second
    err = conn.SetDeadline(time.Now().Add(timeout))
    if err != nil {
        return nil, fmt.Errorf("failed to set deadline: %v", err)
    }

	_, err = conn.Write(query)
	if err != nil {
		return nil, err
	}

	response := make([]byte, 512)
	n, err := conn.Read(response)
	if err != nil {
		return nil, err
	}
	return response[:n], nil
}

func testResolver(resolver string) error {
	if resolver == "" {
		return nil
	}

	query := []byte{
		0x12, 0x34, // Transaction ID
		0x01, 0x00, // Flags: Standard query
		0x00, 0x01, // Questions: 1
		0x00, 0x00, // Answer RRs: 0
		0x00, 0x00, // Authority RRs: 0
		0x00, 0x00, // Additional RRs: 0
		0x07, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, // "example"
		0x03, 0x63, 0x6f, 0x6d, 0x00, // "com"
		0x00, 0x01, // QTYPE: A (IPv4 address)
		0x00, 0x01, // QCLASS: IN (Internet)
	}

	response, err := forwardQueryToResolver(query, resolver)
	if err != nil {
		return fmt.Errorf("failed to query resolver %s: %v", resolver, err)
	}
	if len(response) < 12 {
		return fmt.Errorf("invalid response from resolver %s", resolver)
	}
	return nil
}
