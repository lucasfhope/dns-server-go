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

	offsets := make(map[string]uint)

	// QUESTIONS
	for _, question := range message.Questions {
		writeQname(response, question.QNAME, offsets)

		binary.Write(response, binary.BigEndian, question.QTYPE)
		binary.Write(response, binary.BigEndian, question.QCLASS)
	}

	// ANSWERS
	for _, question := range message.Questions {
		writeQname(response, question.QNAME, offsets)

		binary.Write(response, binary.BigEndian, question.QTYPE)
		binary.Write(response, binary.BigEndian, question.QCLASS)

		binary.Write(response, binary.BigEndian, uint32(67543))
		binary.Write(response, binary.BigEndian, uint16(4))
		binary.Write(response, binary.BigEndian, net.ParseIP("9.8.17.3").To4())
	}

	return response.Bytes()
}

func writeQname(w *bytes.Buffer, name string, offsets map[string]uint) {
	labels := strings.Split(name, ".")
	for i, label := range labels {
		// Write a pointer to any existing name
		remainingName := strings.Join(labels[i:], ".")
		if offset, exists := offsets[remainingName]; exists {
			pointer := 0xC000 | offset
			binary.Write(w, binary.BigEndian, uint16(pointer))
			return
		}
		if i == 0 {
			offsets[name] = uint(w.Len())
		}
		offsets[remainingName] = uint(w.Len())
		w.WriteByte(byte(len(label)))
		w.WriteString(label)
	}
	w.WriteByte(0)
}
