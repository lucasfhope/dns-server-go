package dns

import (
	"bytes"
	"encoding/binary"
	"net"
	"strings"
)

func BuildDNSResponse(message DNSMessage) []byte {
	response := new(bytes.Buffer)

	// HEADER
	flags := uint16(0)
	flags |= 1 << 15                                  // QR
	flags |= message.Header.getOpcode() << 11         // OPCODE
	flags |= 0 << 10                                  // AA
	flags |= 0 << 9                                   // TC
	flags |= message.Header.getRecusionDesired() << 8 // RD
	flags |= 0 << 7                                   // RA
	flags |= 0b000 << 4                               // Z
	if message.Header.getOpcode() != 0b0000 {         // RCODE
		flags |= 0b0100
	}

	header := DNSHeader{
		ID:      message.Header.ID,
		Flags:   flags,
		QDCount: uint16(len(message.Questions)),
		ANCount: uint16(len(message.Questions)),
		NSCount: 0,
		ARCount: 0,
	}
	binary.Write(response, binary.BigEndian, header)

	// QUESTIONS
	for _, question := range message.Questions {
		labels := strings.Split(question.QNAME, ".")
		for _, label := range labels {
			response.WriteByte(byte(len(label)))
			response.WriteString(label)
		}
		response.WriteByte(0)

		binary.Write(response, binary.BigEndian, question.QTYPE)
		binary.Write(response, binary.BigEndian, question.QCLASS)
	}

	// ANSWERS
	for _, question := range message.Questions {
		labels := strings.Split(question.QNAME, ".")
		for _, label := range labels {
			response.WriteByte(byte(len(label)))
			response.WriteString(label)
		}
		response.WriteByte(0)

		binary.Write(response, binary.BigEndian, question.QTYPE)
		binary.Write(response, binary.BigEndian, question.QCLASS)

		binary.Write(response, binary.BigEndian, uint32(67543))
		binary.Write(response, binary.BigEndian, uint16(4))
		binary.Write(response, binary.BigEndian, net.ParseIP("9.8.17.3").To4())
	}

	return response.Bytes()
}
