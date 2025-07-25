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
	if len(data) < 4+2 {
		return nil, ErrTooShort
	}

	wantCRC := binary.LittleEndian.Uint32(data[len(data)-4:])
	if crc32.ChecksumIEEE(data[:len(data)-4]) != wantCRC {
		return nil, ErrCRC
	}

	obj := new(EmailObj)
	obj.len = 1

	if data[0] == 0 || int(data[0]) > len(data)-obj.len {
		return nil, ErrMalformed
	}
	obj.login = string(data[1 : data[0]+1])
	obj.len += len(obj.login)

	if data[data[0]+1] == 0 || int(data[data[0]+1]) > len(data)-obj.len {
		return nil, ErrMalformed
	}
	obj.domain = string(data[data[0]+2 : data[0]+2+data[data[0]+1]])
	obj.len += len(obj.domain)

	payload := data[data[0]+2+data[data[0]+1] : len(data)-4]
	if len(data[data[0]+2+data[data[0]+1]:len(data)-4]) != 0 {
		obj.prefixes = make([]EmailPrefixObj, 0, 2)

		var prefix byte
		for i := 0; i < len(payload); i++ {
			if prefix == 0 {
				prefix = payload[i]
				continue
			} else {
				bufLen := int(payload[i])
				if bufLen == 0 || bufLen > len(data)-obj.len {
					return nil, ErrMalformed
				}

				obj.prefixes = append(obj.prefixes, EmailPrefixObj{char: prefix, text: string(payload[i+1 : i+bufLen+1])})

				prefix = 0
				i += bufLen
				obj.len += bufLen + 1
			}
		}
	}

	return obj, nil
}
