/*
DESCRIPTION
  bitreader.go provides a bit reader implementation that can read or peek from
  an io.Reader data source.

AUTHORS
  Saxon Nelson-Milton <saxon@ausocean.org>, The Australian Ocean Laboratory (AusOcean)

LICENSE

  Copyright (c) 2009 The Go Authors. All rights reserved.

  Redistribution and use in source and binary forms, with or without
  modification, are permitted provided that the following conditions are
  met:

    * Redistributions of source code must retain the above copyright
  notice, this list of conditions and the following disclaimer.
    * Redistributions in binary form must reproduce the above
  copyright notice, this list of conditions and the following disclaimer
  in the documentation and/or other materials provided with the
  distribution.
    * Neither the name of Google Inc. nor the names of its
  contributors may be used to endorse or promote products derived from
  this software without specific prior written permission.

  THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
  "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
  LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
  A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
  OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
  SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
  LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
  DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
  THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
  (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
  OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
*/

// Package bits provides a bit reader implementation that can read or peek from
// an io.Reader data source.
package bits

import (
	"bufio"
	"io"
)

// BitReader is a bit reader that provides methods for reading bits from an
// io.Reader source.
type BitReader struct {
	r    *bufio.Reader
	n    uint64
	bits uint
}

// NewBitReader returns a new BitReader.
func NewBitReader(r io.Reader) *BitReader {
	byter, ok := r.(*bufio.Reader)
	if !ok {
		byter = bufio.NewReader(r)
	}
	return &BitReader{r: byter}
}

// ReadBits reads n bits from the source and returns them the least-significant
// part of a uint64.
// For example, with a source as []byte{0x8f,0xe3} (1000 1111, 1110 0011), we
// would get the following results for consequtive reads with n values:
// n = 4, res = 0x8 (1000)
// n = 2, res = 0x3 (0011)
// n = 4, res = 0xf (1111)
// n = 6, res = 0x23 (0010 0011)
func (br *BitReader) ReadBits(n uint) (uint64, error) {
	for n > br.bits {
		b, err := br.r.ReadByte()
		if err == io.EOF {
			return 0, io.ErrUnexpectedEOF
		}
		if err != nil {
			return 0, err
		}
		br.n <<= 8
		br.n |= uint64(b)
		br.bits += 8
	}

	// br.n looks like this (assuming that br.bits = 14 and bits = 6):
	// Bit: 111111
	//      5432109876543210
	//
	//         (6 bits, the desired output)
	//        |-----|
	//        V     V
	//      0101101101001110
	//        ^            ^
	//        |------------|
	//           br.bits (num valid bits)
	//
	// This the next line right shifts the desired bits into the
	// least-significant places and masks off anything above.
	r := (br.n >> (br.bits - n)) & ((1 << n) - 1)
	br.bits -= n
	return r, nil
}

// PeekBits provides the next n bits returning them in the least-significant
// part of a uint64, without advancing through the source.
// For example, with a source as []byte{0x8f,0xe3} (1000 1111, 1110 0011), we
// would get the following results for consequtive peeks with n values:
// n = 4, res = 0x8 (1000)
// n = 8, res = 0x8f (1000 1111)
// n = 16, res = 0x8fe3 (1000 1111, 1110 0011)
func (br *BitReader) PeekBits(n uint) (uint64, error) {
	byt, err := br.r.Peek(int((n-br.bits)+7) / 8)
	bits := br.bits
	if err != nil {
		if err == io.EOF {
			return 0, io.ErrUnexpectedEOF
		}
		return 0, err
	}
	for i := 0; n > bits; i++ {
		b := byt[i]
		if err != nil {
			return 0, err
		}
		br.n <<= 8
		br.n |= uint64(b)
		bits += 8
	}

	r := (br.n >> (bits - n)) & ((1 << n) - 1)
	return r, nil
}
