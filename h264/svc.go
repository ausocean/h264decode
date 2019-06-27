/*
NAME
  svc.go

DESCRIPTION
  svc.go provides scalable video coding decoding utilities.

AUTHOR
  mrmod <mcmoranbjr@gmail.com>
  Saxon Nelson-Milton <saxon@ausocean.org>, The Australian Ocean Laboratory (AusOcean)
*/

package h264

// Xd gives sample location x difference as per section G.8.6.2.2.2 eq (G-291).
func Xd(xr, refMbW int) int {
	if xr >= refMbW/2 {
		return xr - refMbW
	}
	return xr + 1
}

// Yd gives sample location y difference as per section G.8.6.2.2.2 eq (G-292)
func Yd(yr, refMbH int) int {
	if yr >= refMbH/2 {
		return yr - refMbH
	}
	return yr + 1
}

// Ya is defined by section G.8.6.2.2.2 eq. (G-293).
func Ya(yd int, refMbH uint) int {
	return yd - int((refMbH/2+1))*int(sign(float64(yd)))
}

// RefMbW gives the reference macroblock width as per derivation from section
// G.8.6.2.2 eq. (G-270).
func RefMbW(chromaFlag bool, refLayerMbWidthC uint) uint {
	if !chromaFlag {
		return 16
	}
	return refLayerMbWidthC
}

// RefMbH gives the reference macroblock height as per derivation from section
// G.8.6.2.2 eq. (G-271).
func RefMbH(chromaFlag bool, refLayerMbHeightC uint) uint {
	if !chromaFlag {
		return 16
	}
	return refLayerMbHeightC
}

// Xr is defined by sec. G.8.6.2.2.2 eq. (G-289).
func Xr(x, xOffset int, refMbW uint) int {
	return (x + xOffset) % int(refMbW)
}

// Yr is defined by sec. G.8.6.2.2.2 eq. (G-289).
func Yr(y, yOffset int, refMbH uint) int {
	return (y + yOffset) % int(refMbH)
}

// XOffset is defined by sec. G.8.6.2.2 eq. (G-272).
func XOffset(xRefMin16, refMbW int) int {
	return (((xRefMin16 - 64) >> 8) << 4) - (refMbW >> 1)
}

// YOffset is defined by sec. G.8.6.2.2 eq. (G-273).
func YOffset(yRefMin16, refMbH int) int {
	return (((yRefMin16 - 64) >> 8) << 4) - (refMbH >> 1)
}

func MbWidthC(sps *SPS) int {
	mbWidthC := 16 / SubWidthC(sps)
	if sps.ChromaFormat == 0 || sps.UseSeparateColorPlane {
		mbWidthC = 0
	}
	return mbWidthC
}

func MbHeightC(sps *SPS) int {
	mbHeightC := 16 / SubHeightC(sps)
	if sps.ChromaFormat == 0 || sps.UseSeparateColorPlane {
		mbHeightC = 0
	}
	return mbHeightC
}
