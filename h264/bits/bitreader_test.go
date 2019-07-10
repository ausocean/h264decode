/*
NAME
  bitreader_test.go

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

func TestReadBits(t *testing.T) {
	tests := []struct {
		in    []byte
		reads []uint
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
		gotReads := make([]uint64, len(test.reads))

		var err error
		for j, n := range test.reads {
			gotReads[j], err = br.ReadBits(n)
			if err != nil && errors.Cause(err) != test.errs[j] {
				t.Fatalf("did not expect error: %v for read: %d test: %d", err, j, i)
			}
		}

		// Now check reads.
		if !reflect.DeepEqual(gotReads, test.wants) {
			t.Errorf("did not get expected results from ReadBits for test: %d\nGot: %v\nWant: %v\n", i, gotReads, test.wants)
		}
	}
}
