package puremail

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
)

// // // // // // // // // //

type testParseObj struct {
	name         string
	input        string
	isShot       bool
	wantLogin    string
	wantDomain   string
	wantPrefixes int
	wantErr      bool
}

func TestParse(t *testing.T) {
	t.Parallel()

	tests := []*testParseObj{
		{
			name:       "simple valid (shot)",
			input:      "user@example.com",
			isShot:     true,
			wantLogin:  "user",
			wantDomain: "example.com",
		},
		{
			name:         "valid with prefixes (full)",
			input:        "alice+dev=go@example.io",
			isShot:       false,
			wantLogin:    "alice",
			wantDomain:   "example.io",
			wantPrefixes: 2,
		},
		{
			name:         "valid with prefixes v2 (full)",
			input:        "bob+promo=gophers@gmail.com",
			isShot:       false,
			wantLogin:    "bob",
			wantDomain:   "gmail.com",
			wantPrefixes: 2,
		},
		{
			name:       "uppercase converted to lower",
			input:      "Bob.Smith@GMAIL.COM",
			isShot:     true,
			wantLogin:  "bob.smith",
			wantDomain: "gmail.com",
		},
		{
			name:    "empty login",
			input:   "@example.org",
			isShot:  true,
			wantErr: true,
		},
		{
			name:    "missing domain",
			input:   "john@",
			isShot:  true,
			wantErr: true,
		},
		{
			name:    "two at signs",
			input:   "a@b@c.com",
			isShot:  true,
			wantErr: true,
		},
		{
			name:    "invalid char in login",
			input:   "me(you)@mail.net",
			isShot:  true,
			wantErr: true,
		},
		{
			name:    "domain fails regexp",
			input:   "foo@-bad-.com",
			isShot:  true,
			wantErr: true,
		},
		{
			name:    "too long (>254)",
			input:   strings.Repeat("x", 255) + "@a.com",
			isShot:  true,
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := parse(tc.input, tc.isShot)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("an error expected but it is not")
				}
				return
			}
			if err != nil {
				t.Fatalf("an unexpected error: %v", err)
			}

			if got.login != tc.wantLogin {
				t.Errorf("login = %q, want %q", got.login, tc.wantLogin)
			}
			if got.domain != tc.wantDomain {
				t.Errorf("domain = %q, want %q", got.domain, tc.wantDomain)
			}
			if len(got.prefixes) != tc.wantPrefixes {
				t.Errorf("len(prefixes) = %d, want %d", len(got.prefixes), tc.wantPrefixes)
			}
			if tc.isShot && len(got.prefixes) != 0 {
				t.Errorf("in shot-mode prefixes should be 0 received %#v", len(got.prefixes))
			}

			if !tc.isShot && strings.ToLower(tc.input) != got.MailFull() {
				t.Errorf("input = %q, want %q", got.MailFull(), tc.input)
			}
		})
	}
}

//

type testParseErrObj struct {
	name    string
	input   string
	isShot  bool
	wantErr error
}

func TestParse_Errors(t *testing.T) {
	tests := []*testParseErrObj{
		{
			name:    "ErrLenMax – line is longer than 254 bytes",
			input:   strings.Repeat("x", 255),
			isShot:  false,
			wantErr: ErrLenMax,
		},
		{
			name:    "ErrInvalidLogin – no login to @",
			input:   "@domain.com",
			isShot:  false,
			wantErr: ErrInvalidLogin,
		},
		{
			name:    "ErrManyA – more than one @",
			input:   "a@b@c.com",
			isShot:  false,
			wantErr: ErrManyA,
		},
		{
			name:    "ErrInvalidLoginChars – prohibited symbol in login",
			input:   "us..er@domain.com",
			isShot:  false,
			wantErr: ErrInvalidLoginChars,
		},
		{
			name:    "ErrInvalidDomain – the domain is empty",
			input:   "user@",
			isShot:  false,
			wantErr: ErrInvalidDomain,
		},
		{
			name:    "ErrInvalidDomainChars – the forbidden character in the domain",
			input:   "user@exa^mple.com",
			isShot:  false,
			wantErr: ErrInvalidDomainChars,
		},
		{
			name:    "ErrEndToTag – the string broke after the tag",
			input:   "user+tag",
			isShot:  false,
			wantErr: ErrEndToTag,
		},
		{
			name:    "ErrEndToEOF – the end of the line without @",
			input:   "justlogin",
			isShot:  false,
			wantErr: ErrEndToEOF,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parse(tt.input, tt.isShot)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("the error was obtained %v, expected %v", err, tt.wantErr)
			}
		})
	}
}

//

type testParseAsyncObj struct {
	adr      string
	hash     [hashBlockSize]byte
	hashFull [hashBlockSize]byte
}

func TestAsync(t *testing.T) {
	addresses := make([]testParseAsyncObj, 1000)
	for i := range addresses {
		adr := fmt.Sprintf("user%03d+prefix%03d=sufix%03d@host0.com", i, i, i)
		obj, _ := parse(adr, false)
		addresses[i] = testParseAsyncObj{adr: adr, hash: obj.Hash(), hashFull: obj.HashFull()}
	}

	var wg sync.WaitGroup
	for _, adrObj := range addresses {
		wg.Add(1)

		go func(adrObj *testParseAsyncObj) {
			defer wg.Done()

			if t.Failed() {
				return
			}

			obj, _ := parse(adrObj.adr, false)
			if obj.MailFull() != adrObj.adr {
				t.Errorf("mail full = %q, want %q", obj.MailFull(), adrObj.adr)
			}
			if obj.Hash() != adrObj.hash {
				t.Errorf("hash = %02x, want %02x", obj.Hash(), adrObj.hash)
			}
			if obj.HashFull() != adrObj.hashFull {
				t.Errorf("hashFull = %02x, want %02x", obj.HashFull(), adrObj.hashFull)
			}
		}(&adrObj)
	}
	wg.Wait()
}

// //

type benchParseObj struct {
	name     string
	template string
	bulk     int
}

func BenchmarkParse(b *testing.B) {
	cases := []*benchParseObj{
		{
			name:     "Simple",
			template: "user@example.com",
			bulk:     1,
		},
		{
			name: "Bulk",
			bulk: 1000,
		},
	}

	for _, tc := range cases {
		addresses := make([]string, tc.bulk)
		if tc.bulk == 1 {
			addresses[0] = tc.template
		} else {
			tc.name += fmt.Sprint(tc.bulk)
			for i := range addresses {
				addresses[i] = fmt.Sprintf("user%03d@host0.com", i)
			}
		}

		for prefix := 0; prefix < 2; prefix++ {
			name := tc.name
			if prefix == 1 {
				name += "WithPrefixes"
			}

			var err error
			b.Run(name, func(b *testing.B) {
				b.ReportAllocs()
				b.ResetTimer()

				for n := 0; n < b.N; n++ {
					for _, addr := range addresses {
						if prefix == 1 {
							_, err = NewFast(addr)
						} else {
							_, err = New(addr)
						}
						if err != nil {
							b.Fatal(err)
						}

					}
				}

			})
		}

	}
}
