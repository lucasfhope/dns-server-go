package dns

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strings"
)

func ParseDNSMessage(packet []byte) (DNSMessage, error) {
	header, position, err := parseDNSHeader(packet)
	if err != nil {
		return DNSMessage{}, err
	}

	var questions []DNSQuestion
	for i := 0; i < int(header.QDCount); i++ {
		question, newPosition, err := parseDNSQuestion(packet, position)
		if err != nil {
			return DNSMessage{}, err
		}
		questions = append(questions, question)
		position = newPosition
	}

	var answers []DNSAnswer
	for i := 0; i < int(header.ANCount); i++ {
		answer, newPosition, err := parseDNSAnswer(packet, position)
		if err != nil {
			return DNSMessage{}, err
		}
		answers = append(answers, answer)
		position = newPosition
	}

	message := DNSMessage{
		Header:    header,
		Questions: questions,
		Answers:   answers,
	}
	return message, nil
}

func parseDNSHeader(packet []byte) (DNSHeader, uint, error) {
	var header DNSHeader
	if err := binary.Read(bytes.NewReader(packet), binary.BigEndian, &header); err != nil {
		return DNSHeader{}, 0, fmt.Errorf("[Parse Header Error] %w", err)
	}
	return header, 12, nil
}

func parseDNSQuestion(packet []byte, position uint) (DNSQuestion, uint, error) {
	var question DNSQuestion
	qname, position, err := parseQNAME(packet, position, nil)
	if err != nil {
		return DNSQuestion{}, 0, fmt.Errorf("[Parse Question Error] %w", err)
	}

	if err := binary.Read(bytes.NewReader(packet[position:]), binary.BigEndian, &question.QTYPE); err != nil {
		return DNSQuestion{}, 0, fmt.Errorf("[Parse Question Error] %w", err)
	}
	position += 2

	if err := binary.Read(bytes.NewReader(packet[position:]), binary.BigEndian, &question.QCLASS); err != nil {
		return DNSQuestion{}, 0, fmt.Errorf("[Parse Question Error] %w", err)
	}
	position += 2

	question.QNAME = qname
	return question, position, nil
}

func parseDNSAnswer(packet []byte, position uint) (DNSAnswer, uint, error) {
	var answer DNSAnswer
	qname, position, err := parseQNAME(packet, position, nil)
	if err != nil {
		return DNSAnswer{}, 0, fmt.Errorf("[Parse Answer Error] %w", err)
	}

	if err := binary.Read(bytes.NewReader(packet[position:]), binary.BigEndian, &answer.ATYPE); err != nil {
		return DNSAnswer{}, 0, fmt.Errorf("[Parse Answer Error] %w", err)
	}
	position += 2

	if err := binary.Read(bytes.NewReader(packet[position:]), binary.BigEndian, &answer.ACLASS); err != nil {
		return DNSAnswer{}, 0, fmt.Errorf("[Parse Answer Error] %w", err)
	}
	position += 2

	if err := binary.Read(bytes.NewReader(packet[position:]), binary.BigEndian, &answer.TTL); err != nil {
		return DNSAnswer{}, 0, fmt.Errorf("[Parse Answer Error] %w", err)
	}
	position += 4

	if err := binary.Read(bytes.NewReader(packet[position:]), binary.BigEndian, &answer.RDLENGTH); err != nil {
		return DNSAnswer{}, 0, fmt.Errorf("[Parse Answer Error] %w", err)
	}
	position += 2

	answer.ANAME = qname
	answer.RDATA = binary.BigEndian.Uint32(packet[position : position+uint(answer.RDLENGTH)])
	position += uint(answer.RDLENGTH)

	return answer, position, nil
}

func parseQNAME(packet []byte, position uint, visitedOffsets map[uint]bool) (string, uint, error) {
	if visitedOffsets == nil {
		visitedOffsets = make(map[uint]bool)
	}

	var labels []string
	for {
		length := uint(packet[position])
		position++

		// POINTER
		if length&0xC0 == 0xC0 {
			offset := ((length & 0x3F) << 8) | uint(packet[position])
			position++

			// Prevent infinite loop
			if visitedOffsets[offset] {
				return "", 0, fmt.Errorf("circular pointer reference detected")
			}
			visitedOffsets[offset] = true

			referencedName, _, err := parseQNAME(packet, offset, visitedOffsets)
			if err != nil {
				return "", 0, err
			}
			labels = append(labels, referencedName)
			break
		}

		if length == 0 {
			break
		}
		if position+length > uint(len(packet)) {
			return "", 0, fmt.Errorf("QNAME label length exceeds packet size")
		}
		labels = append(labels, string(packet[position:position+length]))
		position += uint(length)
	}
	return strings.Join(labels, "."), position, nil
}
