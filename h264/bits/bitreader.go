/*
NAME
  bitreader.go

AUTHORS
  Saxon Nelson-Milton <saxon@ausocean.org>, The Australian Ocean Laboratory (AusOcean)
  mrmod <mcmoranbjr@gmail.com>
*/

// Package bits provides a bitreader interface and implementation.
package bits

import (
	"fmt"
	"io"
)

type BitReader struct {
	r          io.Reader
	bytes      []byte
	byteOffset int
	bitOffset  int
	bitsRead   int
	Debug      bool
}

func NewBitreader(r io.Reader) *BitReader {
	return &BitReader{r: r}
}

func (b *BitReader) Bytes() []byte {
	return b.bytes
}
func (b *BitReader) Fastforward(bits int) {
	b.bitsRead += bits
	b.setOffset()
}
func (b *BitReader) setOffset() {
	b.byteOffset = b.bitsRead / 8
	b.bitOffset = b.bitsRead % 8
}

// TODO: MoreRBSPData Section 7.2 p 62
func (b *BitReader) MoreRBSPData() bool {
	if len(b.bytes)-b.byteOffset == 0 {
		return false
	}
	// Read until the least significant bit of any remaining bytes
	// If the least significant bit is 1, that marks the first bit
	// of the rbspTrailingBits() struct. If the bits read is more
	// than 0, then there is more RBSP data
	buf := make([]int, 1)
	cnt := 0
	for buf[0] != 1 {
		if _, err := b.Read(buf); err != nil {
			return false
		}
		cnt++
	}
	return cnt > 0
}
func (b *BitReader) HasMoreData() bool {
	return len(b.bytes)-b.byteOffset > 0
}

func (b *BitReader) IsByteAligned() bool {
	return b.bitOffset == 0
}

func (b *BitReader) ReadOneBit() int {
	buf := make([]int, 1)
	_, _ = b.Read(buf)
	return buf[0]
}
func (b *BitReader) RewindBits(n int) error {
	if n > 8 {
		nBytes := n / 8
		if err := b.RewindBytes(nBytes); err != nil {
			return err
		}
		b.bitsRead -= n
		b.setOffset()
		return nil
	}
	b.bitsRead -= n
	b.setOffset()
	return nil
}

func (b *BitReader) RewindBytes(n int) error {
	if b.byteOffset-n < 0 {
		return fmt.Errorf("attempted to seek below 0")
	}
	b.byteOffset -= n
	b.bitsRead -= n * 8
	b.setOffset()
	return nil
}

// Get bytes without advancing
func (b *BitReader) PeekBytes(n int) ([]byte, error) {
	if len(b.bytes) >= b.byteOffset+n {
		return b.bytes[b.byteOffset : b.byteOffset+n], nil
	}
	return []byte{}, fmt.Errorf("EOF: not enough bytes to give %d (%d @ offset %d", n, len(b.bytes), b.byteOffset)

}

// io.ByteReader interface
func (b *BitReader) ReadByte() (byte, error) {
	if len(b.bytes) > b.byteOffset {
		bt := b.bytes[b.byteOffset]
		b.byteOffset += 1
		return bt, nil
	}
	return byte(0), fmt.Errorf("EOF:  no more bytes")
}
func (b *BitReader) ReadBytes(n int) ([]byte, error) {
	buf := []byte{}
	for i := 0; i < n; i++ {
		if _b, err := b.ReadByte(); err == nil {
			buf = append(buf, _b)
		} else {
			return buf, err
		}
	}
	return buf, nil
}

func (b *BitReader) Read(buf []int) (int, error) {
	return 0, nil

}
func (b *BitReader) NextField(name string, bits int) int {
	buf := make([]int, bits)
	if _, err := b.Read(buf); err != nil {
		fmt.Printf("error reading %d bits for %s: %v\n", bits, name, err)
		return -1
	}
	return bitVal(buf)
}

// TODO: what does this do ?
func bitVal(bits []int) int {
	t := 0
	for i, b := range bits {
		if b == 1 {
			t += 1 << uint((len(bits)-1)-i)
		}
	}
	// fmt.Printf("\t bitVal: %d\n", t)
	return t
}
