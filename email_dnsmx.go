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

type mxObj struct {
	ticker *time.Ticker
	dnsSem *semaphore.Weighted

	shardCounts        uint32
	shards             []mxShardCacheObj
	maxEntriesPerShard int

	confMx *ConfigMxObj
	ctx    context.Context
}

var (
	errNoMX            = &net.DNSError{Err: ErrNilMX.Error(), IsNotFound: true}
	errToManyLookupsMX = &net.DNSError{Err: ErrToManyLookups.Error(), IsNotFound: true}
	lookupMX           = net.DefaultResolver.LookupMX

	mx *mxObj
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

func mxInitValue(conf *ConfigObj) {
	if conf.MX.ShardAbs == 0 || conf.MX.ShardAbs > 31 {
		panic("ShardAbs must be 1..31")
	}

	if conf.MX.ShardMaxSize == 0 {
		conf.MX.ShardMaxSize = 10_000
	}

	if int64(conf.MX.ConcurrencyLimitLookupMX) > int64(^uint(0)>>1) {
		panic("concurrency limit overflows int")
	}
	if conf.MX.TimeoutDnsBurst.Nanoseconds() <= conf.MX.TimeoutDns.Nanoseconds() {
		panic("timeout dns burst is too low")
	}

	ctx, _ := context.WithCancel(conf.Ctx)
	confCopy := *conf

	shardCounts := uint32(2)
	for i := byte(0); i < conf.MX.ShardAbs; i++ {
		shardCounts *= 2
	}

	mx = &mxObj{
		ticker: time.NewTicker(conf.MX.TimeoutRefresh),
		dnsSem: semaphore.NewWeighted(int64(conf.MX.ConcurrencyLimitLookupMX)),

		shardCounts:        shardCounts,
		shards:             make([]mxShardCacheObj, shardCounts),
		maxEntriesPerShard: int(conf.MX.ShardMaxSize),

		ctx:    ctx,
		confMx: &confCopy.MX,
	}

	for i := range mx.shards {
		mx.shards[i].data = make(map[string]*mxEntryObj, 1024)
	}

	go func() {
		for {
			select {
			case <-mx.ticker.C:
				now := time.Now()
				for i := range mx.shards {
					sh := &mx.shards[i]
					sh.mu.Lock()

					for k, v := range sh.data {
						if now.UnixNano() > v.expire {
							sh.group.Forget(k)
							delete(sh.data, k)
						}
					}

					for len(sh.data) > mx.maxEntriesPerShard {
						oldestKey := ""
						oldestExp := int64(^uint64(0) >> 1)
						i := 0

						//todo не забыть переписать на нормальное вытеснение как будет время
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

			case <-mx.ctx.Done():
				return
			}
		}
	}()
}

//

func acquireDNS(ctx context.Context) error {
	return mx.dnsSem.Acquire(ctx, 1)
}

func releaseDNS() { mx.dnsSem.Release(1) }

func nextTTL(positive bool) time.Duration {
	if positive {
		return mx.confMx.TllPos
	}
	return mx.confMx.TllNeg
}

func (obj *EmailObj) HasMX() error {
	idx := crc32.ChecksumIEEE([]byte(obj.domain)) & (mx.shardCounts - 1)
	sh := &mx.shards[int(idx)]

	sh.mu.RLock()
	ent, ok := sh.data[obj.domain]
	sh.mu.RUnlock()

	if ok {
		if time.Now().UnixNano() < ent.expire {
			if time.Until(time.Unix(0, ent.expire)) < mx.confMx.RefreshAhead && ent.err == nil {
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

		ctx, cancel := context.WithTimeout(mx.ctx, mx.confMx.TimeoutDnsBurst)
		err := acquireDNS(ctx)
		cancel()
		if err != nil {
			return errToManyLookupsMX, nil
		}

		ctx, cancel = context.WithTimeout(mx.ctx, mx.confMx.TimeoutDns)
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
