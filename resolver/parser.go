package resolver

import (
	"fmt"
	"strings"
)

const (
	TypeA     = 1   // IPv4 address
	TypeNS    = 2   // Name server
	TypeCNAME = 5   // Canonical name
	TypeSOA   = 6   // Start of authority
	TypeMX    = 15  // Mail exchange
	TypeTXT   = 16  // Text record
	TypeAAAA  = 28  // IPv6 address
	TypeOPT   = 41  // EDNS0 option
	TypeANY   = 255 // All records
)

const (
	ClassIN  = 1   // Internet
	ClassCS  = 2   // CSNET
	ClassCH  = 3   // CHAOS
	ClassHS  = 4   // Hesiod
	ClassANY = 255 // Any class
)

func Parser(data []byte) (*DNSMessage, error) {

	if len(data) < 12 {
		return nil, fmt.Errorf("message too short: %d bytes", len(data))
	}

	header := DNSHeader{
		ID:      uint16(data[0])<<8 | uint16(data[1]),
		Flags:   uint16(data[2])<<8 | uint16(data[3]),
		QDCount: uint16(data[4])<<8 | uint16(data[5]),
		ANCount: uint16(data[6])<<8 | uint16(data[7]),
		NSCount: uint16(data[8])<<8 | uint16(data[9]),
		ARCount: uint16(data[10])<<8 | uint16(data[11]),
	}

	questions, err := parseQuestion(data, int(header.QDCount))
	if err != nil {
		return nil, fmt.Errorf("failed to parse questions: %w", err)
	}

	dnsMessage := &DNSMessage{
		Header:    header,
		Questions: questions,
	}
	return dnsMessage, nil
}

func parseQuestion(data []byte, qCount int) ([]DNSQuestion, error) {
	questions := []DNSQuestion{}
	offset := 12
	for i := 0; i < qCount; i++ {
		name, bytesRead, err := parseName(data, offset)
		if err != nil {
			return nil, fmt.Errorf("failed to parse question name: %w", err)
		}
		offset += bytesRead

		if offset+4 > len(data) {
			return nil, fmt.Errorf("incomplete question section")
		}

		qType := uint16(data[offset])<<8 | uint16(data[offset+1])
		qClass := uint16(data[offset+2])<<8 | uint16(data[offset+3])
		offset += 4

		questions = append(questions, DNSQuestion{
			Name:  name,
			Type:  qType,
			Class: qClass,
		})
	}
	return questions, nil
}

func parseName(data []byte, offset int) (string, int, error) {
	if offset >= len(data) {
		return "", 0, fmt.Errorf("offset out of bounds")
	}

	bytesRead := 0
	labels := []string{}

	for {
		if offset+bytesRead >= len(data) {
			return "", bytesRead, fmt.Errorf("unexpected end of data")
		}

		length := int(data[offset+bytesRead])
		bytesRead++

		if length == 0 {
			break
		}

		if length >= 192 {
			if offset+bytesRead >= len(data) {
				return "", bytesRead, fmt.Errorf("incomplete pointer")
			}

			pointerOffset := ((length & 0x3F) << 8) | int(data[offset+bytesRead])
			bytesRead++
			pointedName, _, err := parseName(data, pointerOffset)
			if err != nil {
				return "", bytesRead, fmt.Errorf("failed to follow pointer: %w", err)
			}

			if len(labels) > 0 {
				labels = append(labels, pointedName)
			} else {
				return pointedName, bytesRead, nil
			}
			break
		}

		if length > 63 {
			return "", bytesRead, fmt.Errorf("invalid label length: %d", length)
		}

		if offset+bytesRead+length > len(data) {
			return "", bytesRead, fmt.Errorf("label extends beyond data")
		}

		labels = append(labels, string(data[offset+bytesRead:offset+bytesRead+length]))
		bytesRead += length
	}
	return joinLabels(labels), bytesRead, nil

}

func joinLabels(labels []string) string {
	if len(labels) == 0 {
		return ""
	}
	var builder strings.Builder
	for i, label := range labels {
		if i > 0 {
			builder.WriteByte('.')
		}
		builder.WriteString(label)
	}
	return builder.String()
}

func (q DNSQuestion) String() string {
	typeName := getTypeName(q.Type)
	className := getClassName(q.Class)
	return fmt.Sprintf("%s (Type:%s Class:%s)", q.Name, typeName, className)
}

func getTypeName(qType uint16) string {
	switch qType {
	case TypeA:
		return "A"
	case TypeNS:
		return "NS"
	case TypeCNAME:
		return "CNAME"
	case TypeMX:
		return "MX"
	case TypeTXT:
		return "TXT"
	case TypeAAAA:
		return "AAAA"
	default:
		return fmt.Sprintf("%d", qType)
	}
}

func getClassName(qClass uint16) string {
	switch qClass {
	case ClassIN:
		return "IN"
	case ClassCH:
		return "CH"
	case ClassHS:
		return "HS"
	default:
		return fmt.Sprintf("%d", qClass)
	}
}
