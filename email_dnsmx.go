package puremail

import (
	"context"
	"golang.org/x/sync/semaphore"
	"golang.org/x/sync/singleflight"
	"hash/crc32"
	"net"
	"sync"
	"time"
)

// // // // // // // // // //

var (
	PosTTL                  = 6 * time.Hour
	NegTTL                  = 15 * time.Minute
	DnsTimeout              = 400 * time.Millisecond
	RefreshTimeout          = 90 * time.Second
	RefreshAhead            = 10 * time.Minute
	ConcurrencyLimit uint32 = 250

	errNoMX            = &net.DNSError{Err: ErrNilMX.Error(), IsNotFound: true}
	errToManyLookupsMX = &net.DNSError{Err: ErrToManyLookups.Error(), IsNotFound: true}
	lookupMX           = net.DefaultResolver.LookupMX

	ticker = time.NewTicker(RefreshTimeout)
	dnsSem = semaphore.NewWeighted(int64(ConcurrencyLimit))
)

const (
	shardBits  = 8
	shardCount = 1 << shardBits
	shardMask  = shardCount - 1

	maxEntriesPerShard = 10_000
)

type mxEntryObj struct {
	expire int64
	err    error
}
type mxShardCacheObj struct {
	mu    sync.RWMutex
	data  map[string]*mxEntryObj
	group singleflight.Group
}

var shards [shardCount]mxShardCacheObj

func init() {
	for i := range shards {
		shards[i].data = make(map[string]*mxEntryObj, 1024)
	}

	go func() {
		for {
			select {
			case <-ticker.C:
				now := time.Now()
				for i := range shards {
					sh := &shards[i]
					sh.mu.Lock()

					for k, v := range sh.data {
						if now.UnixNano() > v.expire {
							sh.group.Forget(k)
							delete(sh.data, k)
						}
					}

					for len(sh.data) > maxEntriesPerShard {
						oldestKey := ""
						oldestExp := int64(^uint64(0) >> 1)
						i := 0

						for k, v := range sh.data {
							if i >= 64 {
								break
							}
							if v.expire < oldestExp {
								oldestExp = v.expire
								oldestKey = k
							}
							i++
						}

						sh.group.Forget(oldestKey)
						delete(sh.data, oldestKey)
					}

					sh.mu.Unlock()
				}
			}
		}
	}()
}

//

func acquireDNS(ctx context.Context) error {
	return dnsSem.Acquire(ctx, 1)
}

func releaseDNS() { dnsSem.Release(1) }

func nextTTL(positive bool) time.Duration {
	if positive {
		return PosTTL
	}
	return NegTTL
}

func (obj *EmailObj) HasMX() error {
	idx := crc32.ChecksumIEEE([]byte(obj.domain)) & shardMask
	sh := &shards[int(idx)]

	sh.mu.RLock()
	ent, ok := sh.data[obj.domain]
	sh.mu.RUnlock()

	if ok {
		if time.Now().UnixNano() < ent.expire {
			if time.Until(time.Unix(0, ent.expire)) < RefreshAhead && ent.err == nil {
				sh.mu.Lock()
				sh.data[obj.domain] = &mxEntryObj{err: ent.err, expire: time.Now().Add(nextTTL(ent.err == nil)).UnixNano()}
				sh.mu.Unlock()
			}
			return ent.err
		}
	}

	v, err, _ := sh.group.Do(obj.domain, func() (any, error) {
		sh.mu.RLock()
		ent = sh.data[obj.domain]
		sh.mu.RUnlock()
		if ent != nil && time.Now().UnixNano() < ent.expire {
			return ent.err, nil
		}

		ctx, cancel := context.WithTimeout(context.Background(), DnsTimeout*4)
		err := acquireDNS(ctx)
		cancel()
		if err != nil {
			return errToManyLookupsMX, nil
		}

		ctx, cancel = context.WithTimeout(context.Background(), DnsTimeout)
		mx, lookupErr := lookupMX(ctx, obj.domain)
		cancel()
		releaseDNS()

		var entryErr error
		if lookupErr != nil || len(mx) == 0 {
			entryErr = errNoMX
		}

		sh.mu.Lock()
		sh.data[obj.domain] = &mxEntryObj{err: entryErr, expire: time.Now().Add(nextTTL(entryErr == nil)).UnixNano()}
		sh.mu.Unlock()

		return entryErr, nil
	})

	if err != nil {
		return err
	}

	err, ok = v.(error)
	if ok {
		return err
	} else {
		return nil
	}
}
