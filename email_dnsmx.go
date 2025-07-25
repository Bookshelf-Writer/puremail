package puremail

import (
	"container/list"
	"context"
	"golang.org/x/sync/singleflight"
	"net"
	"sync"
	"time"
)

// // // // // // // // // //

var (
	VARmxTTL      = time.Hour * 24
	VARmxCapacity = 10_000 // max 1.2МБ
	VARdnsTimeout = 3 * time.Second
)

type mxResultObj struct {
	err    error
	expire time.Time
	elem   *list.Element
}

type mxCacheObj struct {
	mu    sync.Mutex
	data  map[string]*mxResultObj
	lru   *list.List
	group singleflight.Group
}

func newMxCache() *mxCacheObj {
	return &mxCacheObj{
		data: make(map[string]*mxResultObj, 128),
		lru:  list.New(),
	}
}

var (
	globalMX = newMxCache()
	lookupMX = net.DefaultResolver.LookupMX
)

//

func (c *mxCacheObj) hasMX(domain string) error {
	now := time.Now()
	c.mu.Lock()

	if ent, ok := c.data[domain]; ok && ent.expire.After(now) {
		c.lru.MoveToFront(ent.elem)
		err := ent.err
		c.mu.Unlock()
		return err
	}
	c.mu.Unlock()

	v, _, _ := c.group.Do(domain, func() (interface{}, error) {
		ctx, cancel := context.WithTimeout(context.Background(), VARdnsTimeout)
		defer cancel()

		mx, err := lookupMX(ctx, domain)
		if err == nil && len(mx) == 0 {
			err = ErrNilMX
		}

		c.mu.Lock()
		if ent, ok := c.data[domain]; ok {
			ent.err, ent.expire = err, now.Add(VARmxTTL)
			c.lru.MoveToFront(ent.elem)
		} else {
			elem := c.lru.PushFront(domain)
			c.data[domain] = &mxResultObj{
				err:    err,
				expire: now.Add(VARmxTTL),
				elem:   elem,
			}
			if c.lru.Len() > VARmxCapacity {
				oldest := c.lru.Back()
				evictDom := oldest.Value.(string)
				delete(c.data, evictDom)
				c.lru.Remove(oldest)
			}
		}
		c.mu.Unlock()
		return err, nil
	})

	err, ok := v.(error)
	if ok {
		return err
	} else {
		return nil
	}
}

func (obj *EmailObj) HasMX() error {
	return globalMX.hasMX(obj.domain)
}
