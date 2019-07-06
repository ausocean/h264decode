/*
NAME
  bitreader.go

AUTHORS
  Saxon Nelson-Milton <saxon@ausocean.org>, The Australian Ocean Laboratory (AusOcean)
*/

// Package bits provides a bitreader interface and implementation.
package bits

import (
	"io"
	"math"

	"github.com/pkg/errors"
)

type BitReader struct {
	r     io.Reader
	cache []byte
	bits  uint
}

func NewBitReader(r io.Reader) *BitReader {
	return &BitReader{r: r, bits: 8}
}

var errMaxBits = errors.New("can not read more than 64 bits")

func (b *BitReader) ReadBits(n uint) (uint64, error) {
	const maxBits = 64
	if n > maxBits {
		return 0, errMaxBits
	}

	cl := b.cacheLen()
	if n > cl {
		nbytes := math.Ceil(float64(n-cl) / 8.0)
		tmp := make([]byte, int(nbytes))
		_, err := b.r.Read(tmp)
		if err != nil {
			return 0, errors.Wrap(err, "could not read more data from source")
		}
		b.cache = append(b.cache, tmp...)
	}

	var res uint64
	for {
		if n > b.bits {
			res = res << b.bits
			mask := uint64(0xff >> (8 - b.bits))
			res |= uint64(b.cache[0]) & mask
			n -= b.bits
			b.bits = 8
			b.cache = b.cache[1:]
			continue
		}

		res = res << n
		mask := uint64(0xff >> (8 - n))
		res |= uint64((b.cache[0] >> (b.bits - n))) & mask
		if n == b.bits {
			b.bits = 8
			b.cache = b.cache[1:]
			return res, nil
		}
		b.bits -= n
		return res, nil
	}
}

func (b *BitReader) cacheLen() uint {
	return 8*(uint(len(b.cache))-1) + b.bits
}

func (b *BitReader) PeekBits(n int) (uint64, int, error) {
	return 0, 0, nil
}

func (b *BitReader) Read(buf []byte) (int, error) {
	return 0, nil
}
