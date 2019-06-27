/*
NAME
  common.go

DESCRIPTION
  common.go provides common decoding functionality and utilities.

AUTHOR
  Saxon Nelson-Milton <saxon@ausocean.org>, The Australian Ocean Laboratory (AusOcean)
*/

package h264

func sign(a float64) int8 {
	if a < 0 {
		return -1
	}
	return 1
}
