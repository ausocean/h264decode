/*
NAME
  parse.go

DESCRIPTION
  parse.go provides parsing processes for syntax elements of different
  descriptors specified in 7.2 of ITU-T H.264.

AUTHOR
  Saxon Nelson-Milton <saxon@ausocean.org>

LICENSE
  Copyright (C) 2019 the Australian Ocean Lab (AusOcean)

  It is free software: you can redistribute it and/or modify them
  under the terms of the GNU General Public License as published by the
  Free Software Foundation, either version 3 of the License, or (at your
  option) any later version.

  It is distributed in the hope that it will be useful, but WITHOUT
  ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
  FITNESS FOR A PARTICULAR PURPOSE. See the GNU General Public License
  for more details.

  You should have received a copy of the GNU General Public License
  along with revid in gpl.txt. If not, see http://www.gnu.org/licenses.
*/

package h264

import (
	"math"

	"github.com/icza/bitio"
)

// readUe parses a syntax element of ue(v) descriptor, i.e. an unsigned integer
// Exp-Golomb-coded element.
//
// Specified in 9.1 of ITU-T H.264.
func readUe(r bitio.Reader) (uint, error) {
	nZeros := -1
	var err error
	for b := uint64(0); b == 0; nZeros++ {
		b, err = r.ReadBits(1)
		if err != nil {
			return 0, err
		}
	}
	rem, err := r.ReadBits(byte(nZeros))
	if err != nil {
		return 0, err
	}
	return uint(math.Pow(float64(2), float64(nZeros)) - 1 + float64(rem)), nil
}
