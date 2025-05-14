package main

import (
	"fmt"
	"net"

	"github.com/codecrafters-io/dns-server-starter-go/app/dns"
)

func StartDNSServer() {
	udpAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:2053")
	if err != nil {
		fmt.Println("Failed to resolve UDP address:", err)
		return
	}
	fmt.Print("<DNS server listening on ", udpAddr, ">\n")
	listenAndRespond(udpAddr)
	fmt.Print("<DNS server stopped>\n")
}

func listenAndRespond(udpAddr *net.UDPAddr) {
	udpConn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		fmt.Println("Failed to bind to address:", err)
		return
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
			recievedMessage, err := dns.ParseDNSMessage(packet)
			if err != nil {
				fmt.Printf("Failed to parse DNS query from %s\n", source)
				fmt.Println(err)
				return
			}
			for _, question := range recievedMessage.Questions {
				fmt.Printf("Parsed DNS request from %s for %s\n", source, question.QNAME)
			}

			response := dns.BuildDNSResponse(recievedMessage)
			_, err = udpConn.WriteToUDP(response, source)
			if err != nil {
				fmt.Println("Failed to send response:", err)
			}
			fmt.Printf("Sent response to %s\n", source)
		}(packet, source)
	}
}

func main() {
	StartDNSServer()
}
