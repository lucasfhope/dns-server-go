package server_response_test

import (
	"encoding/binary"
	"net"
	"testing"
	"time"

	"github.com/codecrafters-io/dns-server-starter-go/app/dns"
)

func TestDNSServerResponse(t *testing.T) {
	go func() {
		dns.StartDNSServer()
	}()
	time.Sleep(1 * time.Second)

	clientAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:2053")
	if err != nil {
		t.Fatalf("Failed to resolve server address: %v", err)
	}

	conn, err := net.DialUDP("udp", nil, clientAddr)
	if err != nil {
		t.Fatalf("Failed to connect to server: %v", err)
	}
	defer conn.Close()

	testBasicQuery(t, conn)
	testQueryWithUnimplementedOpcode(t, conn)

}

func testBasicQuery(t *testing.T, conn *net.UDPConn) {
	query := []byte{
		0x12, 0x34, // Transaction ID
		0x01, 0x00, // Flags: [QR=0, OPCODE=0000, AA=0, TC=0, RD=1], [RA=0, Z=000, RCODE=0000]
		0x00, 0x01, // Questions: 1
		0x00, 0x00, // Answer RRs: 0
		0x00, 0x00, // Authority RRs: 0
		0x00, 0x00, // Additional RRs: 0
		0x07, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, // "example"
		0x03, 0x63, 0x6f, 0x6d, 0x00, // "com"
		0x00, 0x01, // Type: A
		0x00, 0x01, // Class: IN
	}
	_, err := conn.Write(query)
	if err != nil {
		t.Fatalf("Failed to send query: %v", err)
	}

	// Read and parse the response
	packet := make([]byte, 512)
	n, _, err := conn.ReadFromUDP(packet)
	if err != nil {
		t.Fatalf("Failed to read response: %v", err)
	}
	if n == 0 {
		t.Fatalf("No response received from server")
	}
	response, err := dns.ParseDNSMessage(packet[:n])
	if err != nil {
		t.Fatalf("Failed to parse DNS response: %v", err)
	}

	// Check the response header
	if response.Header.ID != 0x1234 { // Mimic Transaction ID
		t.Errorf("Transaction ID mismatch: got %x, expected %x", response.Header.ID, 0x1234)
	}
	if response.Header.Flags != 0x8100 { // Expected flags (QR=1, mimic OPCODE and RD)
		t.Errorf("Flags mismatch: got %x, expected %x", response.Header.Flags, 0x8180)
	}
	if response.Header.QDCount != 1 {
		t.Errorf("Question count mismatch: got %d, expected 1", response.Header.QDCount)
	}
	if response.Header.ANCount != 1 {
		t.Errorf("Answer count mismatch: got %d, expected 1", response.Header.ANCount)
	}
	if response.Header.NSCount != 0 {
		t.Errorf("Authority count mismatch: got %d, expected 0", response.Header.NSCount)
	}
	if response.Header.ARCount != 0 {
		t.Errorf("Additional count mismatch: got %d, expected 0", response.Header.ARCount)
	}

	// Check the question section
	if len(response.Questions) != 1 {
		t.Errorf("Expected 1 question, got %d", len(response.Questions))
	}
	if response.Questions[0].QNAME != "example.com" {
		t.Errorf("QNAME mismatch: got %s, expected example.com", response.Questions[0].QNAME)
	}
	if response.Questions[0].QTYPE != 1 { // 1 = A (IPv4 address)
		t.Errorf("QTYPE mismatch: got %d, expected 1 (A)", response.Questions[0].QTYPE)
	}
	if response.Questions[0].QCLASS != 1 { // 1 = IN (Internet)
		t.Errorf("QCLASS mismatch: got %d, expected 1 (IN)", response.Questions[0].QCLASS)
	}

	// Check the answer section
	if len(response.Answers) != 1 {
		t.Errorf("Expected 1 answer, got %d", len(response.Answers))
	}
	if response.Answers[0].ANAME != "example.com" {
		t.Errorf("ANAME mismatch: got %s, expected example.com", response.Answers[0].ANAME)
	}
	if response.Answers[0].ATYPE != 1 {
		t.Errorf("ATYPE mismatch: got %d, expected 1 (A)", response.Answers[0].ATYPE)
	}
	if response.Answers[0].ACLASS != 1 {
		t.Errorf("ACLASS mismatch: got %d, expected 1 (IN)", response.Answers[0].ACLASS)
	}
	if response.Answers[0].TTL != uint32(67543) {
		t.Errorf("TTL mismatch: got %d, expected 67543", response.Answers[0].TTL)
	}
	if response.Answers[0].RDLENGTH != 4 { // 4 bytes for IPv4 address
		t.Errorf("RDLENGTH mismatch: got %d, expected 4", response.Answers[0].RDLENGTH)
	}
	rdataBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(rdataBytes, response.Answers[0].RDATA)
	if net.IP(rdataBytes).String() != "9.8.17.3" {
		t.Errorf("RDATA mismatch: got %x, expected 7f000001", response.Answers[0].RDATA)
	}
}

func testQueryWithUnimplementedOpcode(t *testing.T, conn *net.UDPConn) {
	query := []byte{
		0xab, 0xef, // Transaction ID
		0x58, 0x00, // Flags: [QR=0, OPCODE=1011, AA=0, TC=0, RD=0], [RA=0, Z=000, RCODE=0000]
		0x00, 0x01, // Questions: 1
		0x00, 0x00, // Answer RRs: 0
		0x00, 0x00, // Authority RRs: 0
		0x00, 0x00, // Additional RRs: 0
		0x07, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, // "example"
		0x03, 0x63, 0x6f, 0x6d, 0x00, // "com"
		0x00, 0x01, // Type: A
		0x00, 0x01, // Class: IN
	}
	_, err := conn.Write(query)
	if err != nil {
		t.Fatalf("Failed to send query: %v", err)
	}

	// Read and parse the response
	packet := make([]byte, 512)
	n, _, err := conn.ReadFromUDP(packet)
	if err != nil {
		t.Fatalf("Failed to read response: %v", err)
	}
	if n == 0 {
		t.Fatalf("No response received from server")
	}
	response, err := dns.ParseDNSMessage(packet[:n])
	if err != nil {
		t.Fatalf("Failed to parse DNS response: %v", err)
	}

	// only check mimicked ID and flags and error code in RDATA
	if response.Header.ID != 0xabef { // Mimic different Transaction ID
		t.Errorf("Transaction ID mismatch: got %x, expected %x", response.Header.ID, 0xabef)
	}
	if response.Header.Flags != 0xD804 { // Expected flags (QR=1, mimic OPCODE and RD, RCODE=4 for unimplemented opcode)
		t.Errorf("Flags mismatch: got %x, expected %x", response.Header.Flags, 0xD804)
	}
}
