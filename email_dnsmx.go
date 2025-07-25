package puremail

import (
	"context"
	"golang.org/x/sync/singleflight"
	"hash/crc32"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

// // // // // // // // // //

var (
	PosTTL       = 6 * time.Hour
	NegTTL       = 15 * time.Minute
	DnsTimeout   = 1500 * time.Millisecond
	RefreshAhead = 10 * time.Minute

	errNoMX  = &net.DNSError{Err: ErrNilMX.Error(), IsNotFound: true}
	lookupMX = net.DefaultResolver.LookupMX
)

type mxEntryObj struct {
	err    error
	expire int64
}
type mxShardCacheObj struct {
	mu    sync.RWMutex
	data  map[string]*mxEntryObj
	group singleflight.Group
}

const (
	shardBits  = 8
	shardCount = 1 << shardBits
	shardMask  = shardCount - 1
)

var shards [shardCount]mxShardCacheObj

func init() {
	for i := range shards {
		shards[i].data = make(map[string]*mxEntryObj, 1024)
	}
}

//

func nextTTL(positive bool) time.Duration {
	if positive {
		return PosTTL
	}
	return NegTTL
}

func (obj *EmailObj) HasMX() error {
	now := time.Now()
	idx := crc32.ChecksumIEEE([]byte(obj.domain)) & shardMask
	sh := &shards[idx]

	sh.mu.RLock()
	ent := sh.data[obj.domain]
	sh.mu.RUnlock()

	if ent != nil {
		exp := atomic.LoadInt64(&ent.expire)
		if now.UnixNano() < exp {
			if time.Until(time.Unix(0, exp)) < RefreshAhead {
				atomic.StoreInt64(&ent.expire, now.Add(nextTTL(ent.err == nil)).UnixNano())
			}
			return ent.err
		}
	}

	v, _, _ := sh.group.Do(obj.domain, func() (any, error) {
		sh.mu.RLock()
		ent2 := sh.data[obj.domain]
		sh.mu.RUnlock()
		if ent2 != nil && now.UnixNano() < atomic.LoadInt64(&ent2.expire) {
			return ent2.err, nil
		}

		ctx, cancel := context.WithTimeout(context.Background(), DnsTimeout)
		defer cancel()

		mx, lookupErr := lookupMX(ctx, obj.domain)

		var entryErr error
		if lookupErr != nil || len(mx) == 0 {
			entryErr = errNoMX
		}

		newEntry := &mxEntryObj{
			err:    entryErr,
			expire: time.Now().Add(nextTTL(entryErr == nil)).UnixNano(),
		}

		sh.mu.Lock()
		sh.data[obj.domain] = newEntry
		sh.mu.Unlock()

		return entryErr, nil
	})

	err, ok := v.(error)
	if ok {
		return err
	} else {
		return nil
	}
}
