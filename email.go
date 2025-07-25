package puremail

// // // // // // // // // //

type EmailObj struct {
	login, domain string
	prefixes      []EmailPrefixObj
	len           int
}

//

func New(mail string) (*EmailObj, error) {
	return parse(mail, false)
}

func NewFast(mail string) (*EmailObj, error) {
	return parse(mail, true)
}
