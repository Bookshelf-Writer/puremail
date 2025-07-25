package puremail

import "sync"

// // // // // // // // // //

var (
	loginTable [256]bool
	emailPool  = sync.Pool{New: func() any { return &EmailObj{} }}
)

func init() {
	for _, c := range []byte("abcdefghijklmnopqrstuvwxyz0123456789!#$%&'*+/=?^_`{|}~-") {
		loginTable[c] = true
	}
}

func isLoginChar(c byte) bool { return loginTable[c] }

func isValidLabel(label string) bool {
	if len(label) == 0 || len(label) > 63 {
		return false
	}

	if len(label) >= 5 && label[:4] == "xn--" {
		if len(label) > 63 {
			return false
		}
		for i := 4; i < len(label); i++ {
			c := label[i]
			if !('a' <= c && c <= 'z' || '0' <= c && c <= '9' || c == '-') {
				return false
			}
		}
		return true
	}

	first, last := label[0], label[len(label)-1]
	if !('a' <= first && first <= 'z' || '0' <= first && first <= '9') {
		return false
	}
	if !('a' <= last && last <= 'z' || '0' <= last && last <= '9') {
		return false
	}
	for i := 1; i < len(label)-1; i++ {
		c := label[i]
		if !('a' <= c && c <= 'z' || '0' <= c && c <= '9' || c == '-') {
			return false
		}
	}
	return true
}

func clearEmailPool(e *EmailObj) {
	e.login, e.domain = "", ""
	e.len = 0
	e.prefixes = e.prefixes[:0]
	emailPool.Put(e)
}

//

func isValidLogin(s string) bool {
	if len(s) == 0 {
		return false
	}

	segLen := 0
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == '.' {
			if segLen == 0 {
				return false
			}
			segLen = 0
			continue
		}
		if !isLoginChar(c) {
			return false
		}
		segLen++
	}
	return segLen != 0
}

func isValidDomain(s string) bool {
	if len(s) == 0 {
		return false
	}

	start := 0
	for i := 0; i <= len(s); i++ {
		var c byte
		if i == len(s) {
			c = '.'
		} else {
			c = s[i]
		}

		if c == '.' {
			if !isValidLabel(s[start:i]) {
				return false
			}
			start = i + 1
		}
	}
	return true
}

// //

func parseInto(obj *EmailObj, s string, isShot bool) error {
	if len(s) > 254 {
		return ErrLenMax
	}

	obj.len = len(s)

	var buf [254]byte
	bufLen := 0
	var status, tag byte

	for i := 0; i < len(s); i++ {
		c := s[i]

		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}

		switch c {
		case '@':
			if tag != 0 && !isShot {
				obj.prefixes = append(obj.prefixes, EmailPrefixObj{char: tag, text: string(buf[:bufLen])})
				tag = 0
			} else {
				if bufLen == 0 {
					return ErrInvalidLogin
				}
				if len(obj.login) > 0 {
					return ErrManyA
				}

				obj.login = string(buf[:bufLen])
				if !isValidLogin(obj.login) {
					return ErrInvalidLoginChars
				}
			}

			bufLen = 0
			status = 1

		case '+', '=':
			if len(obj.login) == 0 {
				if bufLen == 0 {
					return ErrInvalidLogin
				}

				obj.login = string(buf[:bufLen])
				if !isValidLogin(obj.login) {
					return ErrInvalidLoginChars
				}

			} else if tag != 0 && !isShot {
				obj.prefixes = append(obj.prefixes, EmailPrefixObj{char: tag, text: string(buf[:bufLen])})
			}
			tag = c
			bufLen = 0
			status = 2

		default:
			buf[bufLen] = c
			bufLen++
		}
	}

	if tag != 0 {
		if !isShot {
			obj.prefixes = append(obj.prefixes,
				EmailPrefixObj{char: tag, text: string(buf[:bufLen])})
		}
		tag = 0
		bufLen = 0
	}

	switch status {
	case 1:
		if bufLen == 0 {
			return ErrInvalidDomain
		}
		obj.domain = string(buf[:bufLen])

		if !isValidDomain(obj.domain) {
			return ErrInvalidDomainChars
		}
		return nil

	case 2:
		return ErrEndToTag

	default:
		return ErrEndToEOF
	}
}

func parse(s string, withPrefixes bool) (*EmailObj, error) {
	e := emailPool.Get().(*EmailObj)
	if err := parseInto(e, s, withPrefixes); err != nil {
		clearEmailPool(e)
		return nil, err
	}
	return e, nil
}
