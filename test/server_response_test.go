package server_response_test

import (
	"net"
	"testing"
	"time"

	"github.com/codecrafters-io/dns-server-starter-go/app/mydns"
)

func TestDNSServerResponse(t *testing.T) {
	go func() {
		mydns.StartDNSServer("")
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
	testCanParseCompressedQName(t, conn)
	testRespondsWithCorrectPointers(t, conn)
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

	response, _ := sendMessageAndParseResponse(t, conn, query)

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
	for i, b := range response.Answers[0].RDATA { // 0.0.0.0
		if b != 0x00 {
			t.Errorf("RDATA mismatch at byte %d: got %x, expected 0x00", i, b)
		}
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

	response, _ := sendMessageAndParseResponse(t, conn, query)

	// only check mimicked ID and flags and error code in RDATA
	if response.Header.ID != 0xabef { // Mimic different Transaction ID
		t.Errorf("Transaction ID mismatch: got %x, expected %x", response.Header.ID, 0xabef)
	}
	if response.Header.Flags != 0xD804 { // Expected flags (QR=1, mimic OPCODE and RD, RCODE=4 for unimplemented opcode)
		t.Errorf("Flags mismatch: got %x, expected %x", response.Header.Flags, 0xD804)
	}
}

func testCanParseCompressedQName(t *testing.T, conn *net.UDPConn) {

	query := []byte{
		0x12, 0x34, // Transaction ID
		0x01, 0x00, // Flags: [QR=0, OPCODE=0000, AA=0, TC=0, RD=1], [RA=0, Z=000, RCODE=0000]
		0x00, 0x02, // Questions: 2
		0x00, 0x00, // Answer RRs: 0
		0x00, 0x00, // Authority RRs: 0
		0x00, 0x00, // Additional RRs: 0
		0x07, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, // "example"
		0x03, 0x63, 0x6f, 0x6d, 0x00, // "com"
		0x00, 0x01, // Type: A
		0x00, 0x01, // Class: IN
		0xc0, 0x0c, // Pointer to "example.com"
		0x00, 0x01, // Type: A
		0x00, 0x01, // Class: IN
	}

	response, _ := sendMessageAndParseResponse(t, conn, query)

	// Check the question section
	if len(response.Questions) != 2 {
		t.Errorf("Expected 2 questions, got %d", len(response.Questions))
	}
	for _, question := range response.Questions {
		if question.QNAME != "example.com" {
			t.Errorf("QNAME mismatch: got %s, expected example.com", question.QNAME)
		}
		if question.QTYPE != 1 { // 1 = A (IPv4 address)
			t.Errorf("QTYPE mismatch: got %d, expected 1 (A)", question.QTYPE)
		}
		if question.QCLASS != 1 { // 1 = IN (Internet)
			t.Errorf("QCLASS mismatch: got %d, expected 1 (IN)", question.QCLASS)
		}
	}

	// Check the answer section
	if len(response.Answers) != 2 {
		t.Errorf("Expected 1 answer, got %d", len(response.Answers))
	}
	for _, answer := range response.Answers {
		if answer.ANAME != "example.com" {
			t.Errorf("ANAME mismatch: got %s, expected example.com", answer.ANAME)
		}
		if answer.ATYPE != 1 {
			t.Errorf("ATYPE mismatch: got %d, expected 1 (A)", answer.ATYPE)
		}
		if answer.ACLASS != 1 {
			t.Errorf("ACLASS mismatch: got %d, expected 1 (IN)", answer.ACLASS)
		}
		if answer.TTL != uint32(67543) {
			t.Errorf("TTL mismatch: got %d, expected 67543", answer.TTL)
		}
		if answer.RDLENGTH != 4 { // 4 bytes for IPv4 address
			t.Errorf("RDLENGTH mismatch: got %d, expected 4", answer.RDLENGTH)
		}
		for i, b := range response.Answers[0].RDATA { // 0.0.0.0
			if b != 0x00 {
				t.Errorf("RDATA mismatch at byte %d: got %x, expected 0x00", i, b)
			}
		}
	}
}

func testRespondsWithCorrectPointers(t *testing.T, conn *net.UDPConn) {
	query := []byte{
		0x12, 0x34, // Transaction ID
		0x01, 0x00, // Flags: [QR=0, OPCODE=0000, AA=0, TC=0, RD=1], [RA=0, Z=000, RCODE=0000]
		0x00, 0x02, // Questions: 2
		0x00, 0x00, // Answer RRs: 0
		0x00, 0x00, // Authority RRs: 0
		0x00, 0x00, // Additional RRs: 0

		// QNAME: example.site.com
		0x07, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, // "example"
		0x04, 0x73, 0x69, 0x74, 0x65, // "site"
		0x03, 0x63, 0x6f, 0x6d, 0x00, // "com"
		0x00, 0x01, // Type: A
		0x00, 0x01, // Class: IN

		// QNAME: site.com (using compression, pointer to "site.com")
		0xc0, 0x14, // Pointer to "site.com" (offset 12)
		0x00, 0x01, // Type: A
		0x00, 0x01, // Class: IN
	}

	response, packet := sendMessageAndParseResponse(t, conn, query)

	// Check the question section
	if len(response.Questions) != 2 {
		t.Errorf("Expected 2 questions, got %d", len(response.Questions))
	}
	if response.Questions[0].QNAME != "example.site.com" {
		t.Errorf("QNAME mismatch: got %s, expected example.site.com", response.Questions[0].QNAME)
	}
	if response.Questions[1].QNAME != "site.com" {
		t.Errorf("QNAME mismatch: got %s, expected site.com", response.Questions[1].QNAME)
	}

	// Check the answer section
	if len(response.Answers) != 2 {
		t.Errorf("Expected 2 answers, got %d", len(response.Answers))
	}
	if response.Answers[0].ANAME != "example.site.com" {
		t.Errorf("QNAME mismatch: got %s, expected example.site.com", response.Answers[0].ANAME)
	}
	if response.Answers[1].ANAME != "site.com" {
		t.Errorf("QNAME mismatch: got %s, expected site.com", response.Answers[1].ANAME)
	}

	expectedResponseBody := []byte{
		// Question 1: example.site.com
		0x07, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, // "example"
		0x04, 0x73, 0x69, 0x74, 0x65, // "site"
		0x03, 0x63, 0x6f, 0x6d, 0x00, // "com"
		0x00, 0x01, // Type: A
		0x00, 0x01, // Class: IN

		// Question 2: site.com (pointer)
		0xc0, 0x14, // Pointer to "site.com"
		0x00, 0x01, // Type: A
		0x00, 0x01, // Class: IN

		// Answer 1: example.site.com
		0xc0, 0x0c, // Pointer to "example.site.com"
		0x00, 0x01, // Type: A
		0x00, 0x01, // Class: IN
		0x00, 0x01, 0x07, 0xd7, // TTL: 67543 (0x00010877)
		0x00, 0x04, // RDLENGTH: 4
		0x00, 0x00, 0x00, 0x00, // RDATA: 0.0.0.0

		// Answer 2: site.com (pointer)
		0xc0, 0x14, // Pointer to "site.com"
		0x00, 0x01, // Type: A
		0x00, 0x01, // Class: IN
		0x00, 0x01, 0x07, 0xd7, // TTL: 67543 (0x00010877)
		0x00, 0x04, // RDLENGTH: 4
		0x00, 0x00, 0x00, 0x00, // RDATA: 0.0.0.0
	}

	compareBytes(t, packet[12:], expectedResponseBody)
}

///////////////////////
// Helper functions ///
///////////////////////

func sendMessageAndParseResponse(t *testing.T, conn *net.UDPConn, message []byte) (mydns.DNSMessage, []byte) {
	// Send the message to the server
	_, err := conn.Write(message)
	if err != nil {
		t.Fatalf("Failed to send message: %v", err)
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
	response, err := mydns.ParseDNSMessage(packet[:n])
	if err != nil {
		t.Fatalf("Failed to parse DNS response: %v", err)
	}

	return response, packet[:n]
}

func compareBytes(t *testing.T, actual []byte, expected []byte) {
	if len(actual) != len(expected) {
		t.Errorf("Response length mismatch: got %d bytes, expected %d bytes", len(actual), len(expected))
	}

	for i := 0; i < len(expected) && i < len(actual); i++ {
		if actual[i] != expected[i] {
			t.Errorf("Byte mismatch at position %d: got 0x%02x, expected 0x%02x", i, actual[i], expected[i])
		}
	}
}
