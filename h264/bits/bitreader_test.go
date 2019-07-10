/*
DESCRIPTION
  bitreader_test.go provides testing for functionality defined in bitreader.go.

AUTHORS
  Saxon Nelson-Milton <saxon@ausocean.org>, The Australian Ocean Laboratory (AusOcean)
*/

package bits

import (
	"bytes"
	"io"
	"reflect"
	"testing"

	"github.com/pkg/errors"
)

// TestReadBits checks that BitReader.ReadBits behaves as expected.
func TestReadBits(t *testing.T) {
	tests := []struct {
		in    []byte   // The bytes the source io.Reader will be initialised with.
		reads []uint   // The values of n for the reads we wish to do.
		wants []uint64 // The results we expect for each ReadBits call.
		errs  []error  // The error expected from each ReadBits call.
	}{
		{
			[]byte{0xff},
			[]uint{8},
			[]uint64{0xff},
			[]error{nil},
		},
		{
			[]byte{0xff},
			[]uint{4, 4},
			[]uint64{0x0f, 0x0f},
			[]error{nil, nil},
		},
		{
			[]byte{0xff},
			[]uint{1, 7},
			[]uint64{0x01, 0x7f},
			[]error{nil, nil},
		},
		{
			[]byte{0xff},
			[]uint{1, 7, 4},
			[]uint64{0x01, 0x7f, 0x00},
			[]error{nil, nil, io.EOF},
		},
		{
			[]byte{0xff, 0xff},
			[]uint{8, 8},
			[]uint64{0xff, 0xff},
			[]error{nil, nil},
		},
		{
			[]byte{0xff, 0xff},
			[]uint{4, 8, 4},
			[]uint64{0x0f, 0xff, 0x0f},
			[]error{nil, nil, nil},
		},
		{
			[]byte{0xff, 0xff},
			[]uint{16},
			[]uint64{0xffff},
			[]error{nil},
		},
		{
			[]byte{0x8f, 0xe3},
			[]uint{4, 2, 4, 6},
			[]uint64{0x8, 0x3, 0xf, 0x23},
			[]error{nil, nil, nil, nil},
		},
	}

	for i, test := range tests {
		br := NewBitReader(bytes.NewReader(test.in))

		// Holds the results from the reads.
		gotReads := make([]uint64, len(test.reads))

		// For each value of n defined in test.reads, we call br.ReadBits, collect
		// the result and check the error.
		var err error
		for j, n := range test.reads {
			gotReads[j], err = br.ReadBits(n)
			if err != nil && errors.Cause(err) != test.errs[j] {
				t.Fatalf("did not expect error: %v for read: %d test: %d", err, j, i)
			}
		}

		// Now we can check the read results.
		if !reflect.DeepEqual(gotReads, test.wants) {
			t.Errorf("did not get expected results from ReadBits for test: %d\nGot: %v\nWant: %v\n", i, gotReads, test.wants)
		}
	}
}

// TestPeekBits checks that BitReader.PeekBits behaves as expected.
func TestPeekBits(t *testing.T) {
	tests := []struct {
		in    []byte
		peeks []uint
		wants []uint64
		errs  []error
	}{
		{
			[]byte{0xff},
			[]uint{8},
			[]uint64{0xff},
			[]error{nil},
		},
		{
			[]byte{0x8f, 0xe3},
			[]uint{4, 8, 16},
			[]uint64{0x8, 0x8f, 0x8fe3},
			[]error{nil, nil, nil},
		},
		{
			[]byte{0x8f, 0xe3, 0x8f, 0xe3},
			[]uint{32},
			[]uint64{0x8fe38fe3},
			[]error{nil},
		},
		{
			[]byte{0x8f, 0xe3},
			[]uint{3, 5, 10},
			[]uint64{0x4, 0x11, 0x23f},
			[]error{nil, nil, nil},
		},
		{
			[]byte{0xff},
			[]uint{1, 7, 10},
			[]uint64{0x01, 0x7f, 0x00},
			[]error{nil, nil, io.EOF},
		},
	}

	for i, test := range tests {
		br := NewBitReader(bytes.NewReader(test.in))

		// Holds the results from the peeks.
		gotPeeks := make([]uint64, len(test.peeks))

		// For each value of n defined in test.peeks, we call br.PeekBits, collect
		// the result and check the error.
		var err error
		for j, n := range test.peeks {
			gotPeeks[j], err = br.PeekBits(n)
			if err != nil && errors.Cause(err) != test.errs[j] {
				t.Fatalf("did not expect error: %v for peek: %d test: %d", err, j, i)
			}
		}

		// Now we can check the peek results.
		if !reflect.DeepEqual(gotPeeks, test.wants) {
			t.Errorf("did not get expected results from PeekBits for test: %d\nGot: %v\nWant: %v\n", i, gotPeeks, test.wants)
		}
	}
}
