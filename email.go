package puremail

import (
	"golang.org/x/sync/singleflight"
)

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

type EmailObj struct {
	login, domain string
	prefixes      []EmailPrefixObj
	len           int

	conf *ConfigObj
}

//

var (
	parseGroup singleflight.Group
	conf       *ConfigObj
)

func doParse(mail string, fast bool) (*EmailObj, error) {
	if conf.NoCache {
		return parse(mail, fast)
	}

	v, err, _ := parseGroup.Do(mail, func() (any, error) {
		return parse(mail, fast)
	})
	if err != nil {
		return nil, err
	}
	return v.(*EmailObj), nil
}

//

func Init(configuration *ConfigObj) {
	copyConf := *configuration
	conf = &copyConf

	mxInitValue(&copyConf)
}

func InitDefault() {
	Init(DefaultConfig)
}

func New(mail string) (*EmailObj, error)     { return doParse(mail, false) }
func NewFast(mail string) (*EmailObj, error) { return doParse(mail, true) }
