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
	buf := make([]byte, 0, obj.len+5)

	buf = append(buf, byte(len(obj.login)))
	buf = append(buf, obj.login...)

	buf = append(buf, byte(len(obj.domain)))
	buf = append(buf, obj.domain...)

	for _, p := range obj.prefixes {
		buf = append(buf, p.char)
		buf = append(buf, byte(len(p.text)))
		buf = append(buf, p.text...)
	}

	crc := make([]byte, 4)
	binary.LittleEndian.PutUint32(crc, crc32.ChecksumIEEE(buf))
	buf = append(buf, crc...)

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

	if data[0] == 0 {
		return nil, ErrMalformed
	}
	obj.login = string(data[1 : data[0]+1])
	obj.len += len(obj.login)

	if data[data[0]+1] == 0 {
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

				obj.prefixes = append(obj.prefixes, EmailPrefixObj{char: prefix, text: string(payload[i+1 : i+bufLen+1])})

				prefix = 0
				i += bufLen
				obj.len += bufLen + 1
			}
		}
	}

	return obj, nil
}
