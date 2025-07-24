package puremail

import (
	"encoding/binary"
	"hash/crc32"
	"testing"
)

// // // // // // // // // //

func newObj(login, domain string, prefixes ...EmailPrefixObj) *EmailObj {
	return &EmailObj{
		login:    login,
		domain:   domain,
		prefixes: prefixes,
	}
}

type testEncodeObj struct {
	name    string
	obj     *EmailObj
	wantErr error
}

//

func TestBytesAndDecodeRoundtrip(t *testing.T) {
	tests := []*testEncodeObj{
		{
			"Basic e-mail without prefixes",
			newObj("vasyl", "example.com"),
			nil,
		},
		{
			"With two prefixes",
			newObj("ira", "example.org",
				EmailPrefixObj{char: '!', text: "urgent"},
				EmailPrefixObj{char: '#', text: "promo"},
			),
			nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			data := tc.obj.Bytes()

			got, err := Decode(data)
			if err != tc.wantErr {
				t.Fatalf("expected error %v, received %v", tc.wantErr, err)
			}
			if err != nil {
				return
			}

			if got.login != tc.obj.login || got.domain != tc.obj.domain {
				t.Errorf("roundtrip: expected %q@%q, received %q@%q",
					tc.obj.login, tc.obj.domain, got.login, got.domain)
			}
			if len(got.prefixes) != len(tc.obj.prefixes) {
				t.Fatalf("roundtrip: the number of prefixes is different: %d vs %d",
					len(got.prefixes), len(tc.obj.prefixes))
			}
			for i := range got.prefixes {
				if got.prefixes[i] != tc.obj.prefixes[i] {
					t.Errorf("roundtrip: prefix %d not coinciding: %+v vs %+v",
						i, got.prefixes[i], tc.obj.prefixes[i])
				}
			}
		})
	}
}

func TestDecodeErrors(t *testing.T) {
	if _, err := Decode([]byte{1, 'a'}); err != ErrTooShort {
		t.Fatalf("expected ErrTooShort, received %v", err)
	}

	buf := []byte{0, 0, 0, 0, 0, 0}
	binary.LittleEndian.PutUint32(buf[len(buf)-4:], crc32.ChecksumIEEE(buf[:len(buf)-4]))
	if _, err := Decode(buf); err != ErrMalformed {
		t.Fatalf("expected ErrMalformed for login=0, received %v", err)
	}

	obj := newObj("petro", "mail.ua")
	data := obj.Bytes()
	data[len(data)-1] ^= 0xFF
	if _, err := Decode(data); err != ErrCRC {
		t.Fatalf("expected ErrCRC, received %v", err)
	}
}

func FuzzDecode(f *testing.F) {
	seed := newObj("seed", "ex.ua").Bytes()
	f.Add(seed)

	f.Fuzz(func(t *testing.T, data []byte) {
		obj, err := Decode(data)
		if err == nil && obj == nil {
			t.Fatalf("decode without error but turned nil")
		}
	})
}

// //

func BenchmarkEmailBytesPrefix(b *testing.B) {
	obj := newObj("benchmark_user", "bigcorp.com",
		EmailPrefixObj{char: '*', text: "newsletter"},
		EmailPrefixObj{char: '%', text: "confidential"},
	)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = obj.Bytes()
	}
}

func BenchmarkEmailDecodePrefix(b *testing.B) {
	obj := newObj("benchmark_user", "bigcorp.com",
		EmailPrefixObj{char: '*', text: "newsletter"},
		EmailPrefixObj{char: '%', text: "confidential"},
	)
	data := obj.Bytes()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := Decode(data); err != nil {
			b.Fatalf("Decode: %v", err)
		}
	}
}

func BenchmarkEmailBytes(b *testing.B) {
	obj := newObj("benchmark_user", "bigcorp.com")

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = obj.Bytes()
	}
}

func BenchmarkEmailDecode(b *testing.B) {
	obj := newObj("benchmark_user", "bigcorp.com")
	data := obj.Bytes()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := Decode(data); err != nil {
			b.Fatalf("Decode: %v", err)
		}
	}
}

func BenchmarkEmailHash(b *testing.B) {
	obj := newObj("benchmark_user", "bigcorp.com")

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		obj.Hash()
	}
}

func BenchmarkEmailHashFull(b *testing.B) {
	obj := newObj("benchmark_user", "bigcorp.com",
		EmailPrefixObj{char: '*', text: "newsletter"},
		EmailPrefixObj{char: '%', text: "confidential"},
	)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		obj.HashFull()
	}
}
