package puremail

import (
	"encoding/binary"
	"hash/crc32"
	"strings"
)

// // // // // // // // // //

func (obj *EmailObj) String() string {
	var b strings.Builder
	b.WriteString("[ '")
	b.WriteString(obj.login)
	b.WriteString("@")
	b.WriteString(obj.domain)
	b.WriteString("'")

	if len(obj.prefixes) > 0 {
		b.WriteString(", [")
		for i, p := range obj.prefixes {
			if i > 0 {
				b.WriteString(", ")
			}
			b.WriteString("'")
			b.WriteString(p.String())
			b.WriteString("'")
		}
		b.WriteString("]")
	}
	b.WriteString(" ]")
	return b.String()
}

func (obj *EmailObj) Bytes() []byte {
	total := 1 + len(obj.login) + 1 + len(obj.domain) + 4
	for _, p := range obj.prefixes {
		total += 1 + 1 + len(p.text)
	}
	buf := make([]byte, total)

	i := 0
	buf[i] = byte(len(obj.login))
	i++
	copy(buf[i:], obj.login)
	i += len(obj.login)

	buf[i] = byte(len(obj.domain))
	i++
	copy(buf[i:], obj.domain)
	i += len(obj.domain)

	for _, p := range obj.prefixes {
		buf[i] = p.char
		i++
		buf[i] = byte(len(p.text))
		i++
		copy(buf[i:], p.text)
		i += len(p.text)
	}

	crc := crc32.ChecksumIEEE(buf[:i])
	binary.LittleEndian.PutUint32(buf[i:], crc)
	return buf
}

func Decode(data []byte) (*EmailObj, error) {
	payloadLen := len(data) - 4
	if payloadLen < 2 {
		return nil, ErrTooShort
	}

	wantCRC := binary.LittleEndian.Uint32(data[payloadLen:])
	if crc32.ChecksumIEEE(data[:payloadLen]) != wantCRC {
		return nil, ErrCRC
	}

	pos := 0
	loginLen := int(data[pos])
	pos++

	if loginLen == 0 || pos+loginLen > payloadLen {
		return nil, ErrMalformed
	}
	obj := new(EmailObj)
	obj.login = string(data[pos : pos+loginLen])
	pos += loginLen

	if pos >= payloadLen {
		return nil, ErrMalformed
	}
	domainLen := int(data[pos])
	pos++
	if domainLen == 0 || pos+domainLen > payloadLen {
		return nil, ErrMalformed
	}
	obj.domain = string(data[pos : pos+domainLen])
	pos += domainLen

	if pos < payloadLen {
		var prefix byte
		txtLen := 0

		for pos < payloadLen {
			prefix = data[pos]
			pos++
			if pos >= payloadLen {
				return nil, ErrMalformed
			}

			txtLen = int(data[pos])
			pos++
			if txtLen == 0 || pos+txtLen > payloadLen {
				return nil, ErrMalformed
			}
			obj.prefixes = append(obj.prefixes, EmailPrefixObj{char: prefix, text: string(data[pos : pos+txtLen])})
			pos += txtLen
		}
	}

	return obj, nil
}
