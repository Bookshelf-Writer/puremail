package puremail

import (
	"net"
)

// // // // // // // // // //

func (obj *EmailObj) Prefixes() []EmailPrefixObj {
	return obj.prefixes
}

func (obj *EmailObj) Login() string {
	return obj.login
}

func (obj *EmailObj) Domain() string {
	return obj.domain
}

func (obj *EmailObj) Mail() string {
	return obj.login + "@" + obj.domain
}

func (obj *EmailObj) MailFull() string {
	if len(obj.prefixes) == 0 {
		return obj.Mail()
	}

	b := make([]byte, 0, obj.len)

	b = append(b, obj.login...)

	for _, p := range obj.prefixes {
		b = append(b, p.char)
		b = append(b, p.text...)
	}

	b = append(b, '@')
	b = append(b, obj.domain...)

	return string(b)
}

//

func (obj *EmailObj) HasMX() error {
	mx, err := net.LookupMX(obj.domain)
	if err != nil {
		return err
	}

	if len(mx) == 0 {
		return ErrNilMX
	}

	return nil
}
