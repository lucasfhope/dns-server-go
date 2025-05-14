package dns

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strings"
)

func ParseDNSMessage(packet []byte) (DNSMessage, error) {
	header, packet, err := parseDNSHeader(packet)
	if err != nil {
		return DNSMessage{}, err
	}

	var questions []DNSQuestion
	for i := 0; i < int(header.QDCount); i++ {
		question, remainingPacket, err := parseDNSQuestion(packet)
		if err != nil {
			return DNSMessage{}, err
		}
		questions = append(questions, question)
		packet = remainingPacket
	}

	var answers []DNSAnswer
	for i := 0; i < int(header.ANCount); i++ {
		answer, remainingPacket, err := parseDNSAnswer(packet)
		if err != nil {
			return DNSMessage{}, err
		}
		answers = append(answers, answer)
		packet = remainingPacket
	}

	message := DNSMessage{
		Header:    header,
		Questions: questions,
		Answers:   answers,
	}
	return message, nil
}

func parseDNSHeader(packet []byte) (DNSHeader, []byte, error) {
	var header DNSHeader
	if err := binary.Read(bytes.NewReader(packet), binary.BigEndian, &header); err != nil {
		return DNSHeader{}, packet, fmt.Errorf("[Parse Header Error] %w", err)
	}
	return header, packet[12:], nil
}

func parseDNSQuestion(packet []byte) (DNSQuestion, []byte, error) {
	var question DNSQuestion
	qname, n, err := parseQNAME(packet)
	if err != nil {
		return DNSQuestion{}, packet, fmt.Errorf("[Parse Question Error] %w", err)
	}
	packet = packet[n:]

	if err := binary.Read(bytes.NewReader(packet), binary.BigEndian, &question.QTYPE); err != nil {
		return DNSQuestion{}, packet, fmt.Errorf("[Parse Question Error] %w", err)
	}
	packet = packet[2:]

	if err := binary.Read(bytes.NewReader(packet), binary.BigEndian, &question.QCLASS); err != nil {
		return DNSQuestion{}, packet, fmt.Errorf("[Parse Question Error] %w", err)
	}
	packet = packet[2:]

	question.QNAME = qname
	return question, packet, nil
}

func parseDNSAnswer(packet []byte) (DNSAnswer, []byte, error) {
	var answer DNSAnswer
	qname, n, err := parseQNAME(packet)
	if err != nil {
		return DNSAnswer{}, packet, fmt.Errorf("[Parse Answer Error] %w", err)
	}
	packet = packet[n:]

	if err := binary.Read(bytes.NewReader(packet), binary.BigEndian, &answer.ATYPE); err != nil {
		return DNSAnswer{}, packet, fmt.Errorf("[Parse Answer Error] %w", err)
	}
	packet = packet[2:]

	if err := binary.Read(bytes.NewReader(packet), binary.BigEndian, &answer.ACLASS); err != nil {
		return DNSAnswer{}, packet, fmt.Errorf("[Parse Answer Error] %w", err)
	}
	packet = packet[2:]

	if err := binary.Read(bytes.NewReader(packet), binary.BigEndian, &answer.TTL); err != nil {
		return DNSAnswer{}, packet, fmt.Errorf("[Parse Answer Error] %w", err)
	}
	packet = packet[4:]

	if err := binary.Read(bytes.NewReader(packet), binary.BigEndian, &answer.RDLENGTH); err != nil {
		return DNSAnswer{}, packet, fmt.Errorf("[Parse Answer Error] %w", err)
	}
	packet = packet[2:]

	answer.ANAME = qname
	answer.RDATA = binary.BigEndian.Uint32(packet[:4])
	packet = packet[4:]

	return answer, packet, nil
}

func parseQNAME(packet []byte) (string, int, error) {
	var labels []string
	var n int
	for {
		length := int(packet[n])
		n++
		if length == 0 {
			break
		}
		if n+length > len(packet) {
			return "", 0, fmt.Errorf("QNAME label length exceeds packet size")
		}
		labels = append(labels, string(packet[n:n+length]))
		n += int(length)
	}
	return strings.Join(labels, "."), n, nil
}
