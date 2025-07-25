package puremail

import (
	"golang.org/x/sync/singleflight"
)

// // // // // // // // // //

type EmailObj struct {
	login, domain string
	prefixes      []EmailPrefixObj
	len           int
}

//

var (
	NoCache  = true
	newGroup singleflight.Group
)

func doParse(mail string, fast bool) (*EmailObj, error) {
	if NoCache {
		return parse(mail, fast)
	}

	v, err, _ := newGroup.Do(mail, func() (any, error) {
		return parse(mail, fast)
	})
	if err != nil {
		return nil, err
	}
	return v.(*EmailObj), nil
}

//

func New(mail string) (*EmailObj, error)     { return doParse(mail, false) }
func NewFast(mail string) (*EmailObj, error) { return doParse(mail, true) }
