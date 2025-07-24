package puremail

import (
	"golang.org/x/crypto/blake2b"
	"hash"
	"io"
	"sync"
)

// // // // // // // // // //

const hashBlockSize = 20

var blakePool = sync.Pool{
	New: func() any {
		h, _ := blake2b.New(hashBlockSize, nil)
		return h
	},
}

func getHasher() hash.Hash {
	return blakePool.Get().(hash.Hash)
}

func putHasher(h hash.Hash) {
	h.Reset()
	blakePool.Put(h)
}

//

func (obj *EmailObj) sum(includePrefixes bool) (out [hashBlockSize]byte) {
	h := getHasher()
	defer putHasher(h)

	io.WriteString(h, obj.login)
	io.WriteString(h, obj.domain)

	if includePrefixes {
		var cbuf [1]byte
		for _, p := range obj.prefixes {
			io.WriteString(h, p.text)
			cbuf[0] = p.char
			h.Write(cbuf[:])
		}
	}
	h.Sum(out[:0])
	return
}

func (obj *EmailObj) Hash() [hashBlockSize]byte     { return obj.sum(false) }
func (obj *EmailObj) HashFull() [hashBlockSize]byte { return obj.sum(true) }
