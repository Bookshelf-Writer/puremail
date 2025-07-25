package puremail

import "errors"

// // // // // // // // // //

var (
	ErrLenMax             = errors.New("too many characters")
	ErrManyA              = errors.New("to many @")
	ErrInvalidLogin       = errors.New("invalid email login")
	ErrInvalidLoginChars  = errors.New("invalid email login characters")
	ErrInvalidDomain      = errors.New("invalid email domain")
	ErrInvalidDomainChars = errors.New("invalid email domain characters")
	ErrEndToTag           = errors.New("end to tag")
	ErrEndToEOF           = errors.New("end to EOF")
	ErrPanic              = errors.New("catch panic")

	ErrTooShort  = errors.New("payload is too short")
	ErrCRC       = errors.New("CRC‑32 mismatch")
	ErrMalformed = errors.New("malformed payload")

	ErrNilMX         = errors.New("no MX records found")
	ErrToManyLookups = errors.New("too many lookups")
)
