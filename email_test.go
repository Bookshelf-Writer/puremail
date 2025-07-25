package puremail

import (
	"context"
	"encoding/binary"
	"hash/crc32"
	"net"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"
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

//

func stubMxLookup(counter *int32) func(ctx context.Context, domain string) ([]*net.MX, error) {
	return func(ctx context.Context, domain string) ([]*net.MX, error) {
		atomic.AddInt32(counter, 1)
		return []*net.MX{{Host: "mx." + domain, Pref: 10}}, nil
	}
}

func TestMxCacheHit(t *testing.T) {
	var calls int32

	old := lookupMX
	lookupMX = stubMxLookup(&calls)
	defer func() { lookupMX = old }()

	c := newMxCache()

	if err := c.hasMX("example.com"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := c.hasMX("example.com"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Errorf("want 1 DNS lookup, got %d", got)
	}
}

func TestMxTTLExpiration(t *testing.T) {
	var calls int32
	old := lookupMX
	lookupMX = stubMxLookup(&calls)
	defer func() { lookupMX = old }()

	c := newMxCache()
	domain := "expire.com"

	if err := c.hasMX(domain); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Manually expire the item.
	c.mu.Lock()
	if ent, ok := c.data[domain]; ok {
		ent.expire = time.Now().Add(-time.Minute)
	}
	c.mu.Unlock()

	if err := c.hasMX(domain); err != nil {
		t.Fatalf("unexpected error after expiry: %v", err)
	}

	if got := atomic.LoadInt32(&calls); got != 2 {
		t.Errorf("want 2 DNS lookups (before & after expiry), got %d", got)
	}
}

func TestMxLRUEviction(t *testing.T) {
	var calls int32
	oldLookup := lookupMX
	lookupMX = stubMxLookup(&calls)
	defer func() { lookupMX = oldLookup }()

	oldCap := VARmxCapacity
	VARmxCapacity = 3
	defer func() { VARmxCapacity = oldCap }()

	c := newMxCache()

	_ = c.hasMX("a.com")
	_ = c.hasMX("b.com")
	_ = c.hasMX("c.com")

	if l := c.lru.Len(); l != 3 {
		t.Fatalf("expected LRU size 3, got %d", l)
	}

	_ = c.hasMX("d.com")

	if l := c.lru.Len(); l != 3 {
		t.Fatalf("expected LRU size 3 after eviction, got %d", l)
	}

	_ = c.hasMX("a.com")

	if got := atomic.LoadInt32(&calls); got != 5 {
		t.Errorf("want 5 DNS lookups in total, got %d", got)
	}
}

func TestMxSingleFlight(t *testing.T) {
	var calls int32
	oldLookup := lookupMX
	lookupMX = func(ctx context.Context, domain string) ([]*net.MX, error) {
		atomic.AddInt32(&calls, 1)
		time.Sleep(40 * time.Millisecond)
		return []*net.MX{{Host: "mx." + domain, Pref: 10}}, nil
	}
	defer func() { lookupMX = oldLookup }()

	c := newMxCache()

	const workers = 20
	wg := sync.WaitGroup{}
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			if err := c.hasMX("parallel.com"); err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		}()
	}
	wg.Wait()

	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Errorf("want 1 DNS lookup, got %d", got)
	}
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

//

func BenchmarkHasMXCached(b *testing.B) {
	var calls int32
	oldLookup := lookupMX
	lookupMX = stubMxLookup(&calls)
	defer func() { lookupMX = oldLookup }()

	c := newMxCache()
	_ = c.hasMX("bench.com")

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = c.hasMX("bench.com")
	}
}

func BenchmarkHasMXMiss(b *testing.B) {
	var calls int32
	oldLookup := lookupMX
	lookupMX = stubMxLookup(&calls)
	defer func() { lookupMX = oldLookup }()

	c := newMxCache()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		domain := "miss" + strconv.Itoa(i) + ".com"
		_ = c.hasMX(domain)
	}
}

func BenchmarkHasMXParallel(b *testing.B) {
	var calls int32
	oldLookup := lookupMX
	lookupMX = func(ctx context.Context, domain string) ([]*net.MX, error) {
		atomic.AddInt32(&calls, 1)
		return []*net.MX{{Host: "mx." + domain, Pref: 10}}, nil
	}
	defer func() { lookupMX = oldLookup }()

	c := newMxCache()

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = c.hasMX("p.com")
		}
	})

	b.ReportMetric(float64(atomic.LoadInt32(&calls)), "dns_calls")
}
