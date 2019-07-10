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
	maxBytes = maxBits / 8 // Max number of bytes we will need in the tmp slice.
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
func (c *cache) get(i int) byte {
	return (*c)[i]
}

// BitReader is an io.Reader that provides additional methods for reading bits
// ReadBits, and peeking bits, PeekBits.
type BitReader struct {
	r    io.Reader // The data source.
	c    *cache    // Cache to hold data we're in the middle of reading.
	bits uint      // Denotes the number of bits left in the start byte of the cache.
	tmp  []byte    // Used to get data from the source and copy into the cache.
}

// NewBitReader returns a new BitReader.
func NewBitReader(r io.Reader) *BitReader {
	return &BitReader{
		r:    r,
		bits: 8,
		tmp:  make([]byte, 0, maxBytes),
		c:    (*cache)(&[]byte{}),
	}
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
	var i int
	return readBits(b, n, &b.bits, b.c, &i, func() { b.c.delete() })
}

// PeekBits provides the next n bits, but without advancing through the source.
// For example, with a source as []byte{0x8f,0xe3} (1000 1111, 1110 0011), we
// would get the following results for consequtive peeks with n values:
// n = 4, res = 0x8 (1000)
// n = 8, res = 0x8f (1000 1111)
// n = 16, res = 0x8fe3 (1000 1111, 1110 0011)
func (b *BitReader) PeekBits(n uint) (uint64, error) {
	bits := b.bits
	var i int
	return readBits(b, n, &bits, b.c, &i, func() { i++ })
}

func readBits(b *BitReader, n uint, bits *uint, c *cache, i *int, advance func()) (uint64, error) {
	if n > maxBits {
		return 0, errMaxBits
	}

	err := b.resizeCache(n)
	if err != nil {
		return 0, errors.Wrap(err, "could not resize cache")
	}

	var res uint64
	for {
		if n > *bits {
			res = appendBits(res, *bits, 8-*bits, 0, b.c.get(*i))
			n -= *bits
			*bits = 8
			advance()
			continue
		}

		res = appendBits(res, n, 8-*bits, *bits-n, b.c.get(*i))
		if n == *bits {
			*bits = 8
			advance()
			return res, nil
		}
		*bits -= n
		return res, nil
	}
}

func appendBits(res uint64, rshift, mshift, bshift uint, from byte) uint64 {
	res = res << rshift
	mask := uint64(0xff >> mshift)
	res |= uint64(from>>bshift) & mask
	return res
}

func (b *BitReader) resizeCache(n uint) error {
	l := 8*(uint(len(*b.c))-1) + b.bits
	if n > l {
		nbytes := math.Ceil(float64(n-l) / 8.0)
		b.tmp = b.tmp[:int(nbytes)]
		_, err := b.r.Read(b.tmp)
		if err != nil {
			return errors.Wrap(err, "could not read more data from source")
		}
		b.c.add(b.tmp)
	}
	return nil
}
