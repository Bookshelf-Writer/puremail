package puremail

import (
	"errors"
)

// // // // // // // // // //

var loginTable [256]bool

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

func parseRecover(err *error) {
	if r := recover(); r != nil {
		*err = errors.Join(ErrPanic, *err)
	}
}

// //

func parse(s string, isShot bool) (obj *EmailObj, err error) {
	if len(s) > 254 {
		return nil, ErrLenMax
	}

	defer parseRecover(&err)

	obj = new(EmailObj)
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
					err = ErrInvalidLogin
					return
				}
				if len(obj.login) > 0 {
					err = ErrManyA
					return
				}

				obj.login = string(buf[:bufLen])
				if !isValidLogin(obj.login) {
					err = ErrInvalidLoginChars
					return
				}
			}

			bufLen = 0
			status = 1

		case '+', '=':
			if len(obj.login) == 0 {
				if bufLen == 0 {
					err = ErrInvalidLogin
					return
				}

				obj.login = string(buf[:bufLen])
				if !isValidLogin(obj.login) {
					err = ErrInvalidLoginChars
					return
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
			err = ErrInvalidDomain
			return
		}
		obj.domain = string(buf[:bufLen])

		if !isValidDomain(obj.domain) {
			err = ErrInvalidDomainChars
			return
		}
		return

	case 2:
		err = ErrEndToTag
		return

	default:
		err = ErrEndToEOF
		return
	}
}
