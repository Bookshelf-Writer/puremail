package puremail

// // // // // // // // // //

type EmailPrefixObj struct {
	char byte
	text string
}

func (p *EmailPrefixObj) String() string {
	return p.text
}

func (p *EmailPrefixObj) Prefix() byte {
	return p.char
}
