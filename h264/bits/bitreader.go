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

const (
	maxBits  = 64          // Max number of bits we can read in one go.
	maxBytes = maxBits / 8 // Max number of bytes in the cache.
)

type cache []byte

// add adds a byte to the end of the cache.
func (c *cache) add(d []byte) {
	*c = append(*c, d...)
}

// delete removes the byte at the start of the cache.
func (c *cache) delete() {
	copy((*c)[0:], (*c)[1:])
	(*c) = (*c)[:len(*c)-1]
}

// next provides the byte at the start of the cache.
func (c *cache) next() byte {
	return (*c)[0]
}

// BitReader is an io.Reader that provides additional methods for reading bits
// ReadBits, and peeking bits, PeekBits.
type BitReader struct {
	r    io.Reader
	c    *cache
	bits uint
	tmp  []byte
}

// NewBitReader returns a new BitReader.
func NewBitReader(r io.Reader) *BitReader {
	return &BitReader{r: r, bits: 8, tmp: make([]byte, 0, maxBytes), c: (*cache)(&[]byte{})}
}

// Error used by ReadBits.
var errMaxBits = errors.New("can not read more than 64 bits")

// ReadBits reads n bits from the source and returns as an uint64.
// For example, with a source as []byte{0x8f,0xe3} (1000 1111, 1110 0011), we
// would get the following results for consequtive reads with n values:
// n = 4, res = 0x8 (1000)
// n = 2, res = 0x3 (0011)
// n = 4, res = 0xf (1111)
// n = 6, res = 0x23 (0010 0011)
func (b *BitReader) ReadBits(n uint) (uint64, error) {
	if n > maxBits {
		return 0, errMaxBits
	}

	l := 8*(uint(len(*b.c))-1) + b.bits
	if n > l {
		nbytes := math.Ceil(float64(n-l) / 8.0)
		b.tmp = b.tmp[:int(nbytes)]
		_, err := b.r.Read(b.tmp)
		if err != nil {
			return 0, errors.Wrap(err, "could not read more data from source")
		}
		b.c.add(b.tmp)
	}

	var res uint64
	for {
		if n > b.bits {
			res = res << b.bits
			mask := uint64(0xff >> (8 - b.bits))
			res |= uint64(b.c.next()) & mask
			n -= b.bits
			b.bits = 8
			b.c.delete()
			continue
		}

		res = res << n
		mask := uint64(0xff >> (8 - n))
		res |= uint64((b.c.next() >> (b.bits - n))) & mask
		if n == b.bits {
			b.bits = 8
			b.c.delete()
			return res, nil
		}
		b.bits -= n
		return res, nil
	}
}

func (b *BitReader) PeekBits(n int) (uint64, int, error) {
	return 0, 0, nil
}

func (b *BitReader) Read(buf []byte) (int, error) {
	return 0, nil
}
