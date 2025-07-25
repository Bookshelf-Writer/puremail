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

var newGroup singleflight.Group

func New(mail string) (*EmailObj, error) {
	v, err, _ := newGroup.Do(mail, func() (any, error) {
		obj, err := parse(mail, false)
		if err != nil {
			return nil, err
		}
		return obj, nil
	})

	if err != nil {
		return nil, err
	}

	return v.(*EmailObj), nil
}

func NewFast(mail string) (*EmailObj, error) {
	v, err, _ := newGroup.Do(mail, func() (any, error) {
		obj, err := parse(mail, true)
		if err != nil {
			return nil, err
		}
		return obj, nil
	})

	if err != nil {
		return nil, err
	}

	return v.(*EmailObj), nil
}
