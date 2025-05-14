package dns

type DNSMessage struct {
	Header    DNSHeader
	Questions []DNSQuestion
	Answers   []DNSAnswer
}

type DNSHeader struct {
	ID    uint16 // Packet Identifier
	Flags uint16
	// Query Response (1 bit)
	// OPCODE (4 bits),
	// Authoritative Answer (1 bit),
	// Truncated (1 bit),
	// Recursion Desired (1 bit),
	// Recursion Available (1 bit),
	// Reserved (3 bits)
	// Response Code (4 bits)
	QDCount uint16 // Question Count
	ANCount uint16 // Answer Record Count
	NSCount uint16 // Authority Record Count
	ARCount uint16 // Additional Record
}

func (h DNSHeader) getOpcode() uint16 {
	return (h.Flags >> 11) & 0b1111
}

func (h DNSHeader) getRecusionDesired() uint16 {
	return (h.Flags >> 8) & 0b1
}

type DNSQuestion struct {
	QNAME  string // Domain Name
	QTYPE  uint16 // Type of query
	QCLASS uint16 // Class of query
}

type DNSAnswer struct {
	ANAME    string // Domain Name
	ATYPE    uint16 // Type of answer
	ACLASS   uint16 // Class of answer
	TTL      uint32 // Time to Live
	RDLENGTH uint16 // Length of the resource data
	RDATA    uint32 // Resource data
}
