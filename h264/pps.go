package h264

import "fmt"
import "math"

// import "strings"

// Specification Page 46 7.3.2.2

type PPS struct {
	ID, SPSID                         int
	EntropyCodingMode                 int
	NumSliceGroupsMinus1              int
	BottomFieldPicOrderInFramePresent bool
	NumSlicGroupsMinus1               int
	SliceGroupMapType                 int
	RunLengthMinus1                   []int
	TopLeft                           []int
	BottomRight                       []int
	SliceGroupChangeDirection         bool
	SliceGroupChangeRateMinus1        int
	PicSizeInMapUnitsMinus1           int
	SliceGroupId                      []int
	NumRefIdxL0DefaultActiveMinus1    int
	NumRefIdxL1DefaultActiveMinus1    int
	WeightedPred                      bool
	WeightedBipred                    int
	PicInitQpMinus26                  int
	PicInitQsMinus26                  int
	ChromaQpIndexOffset               int
	DeblockingFilterControlPresent    bool
	ConstrainedIntraPred              bool
	RedundantPicCntPresent            bool
	Transform8x8Mode                  int
	PicScalingMatrixPresent           bool
	PicScalingListPresent             []bool
	SecondChromaQpIndexOffset         int
}

func NewPPS(sps *SPS, rbsp []byte, showPacket bool) *PPS {
	logger.Printf("debug: PPS RBSP %d bytes %d bits == \n", len(rbsp), len(rbsp)*8)
	logger.Printf("debug: \t%#v\n", rbsp[0:8])
	pps := PPS{}
	b := &BitReader{bytes: rbsp}
	flagField := func() bool {
		if v := b.NextField("", 1); v == 1 {
			return true
		}
		return false
	}

	pps.ID, _ = readUe(nil)
	pps.SPSID, _ = readUe(nil)
	pps.EntropyCodingMode = b.NextField("EntropyCodingModeFlag", 1)
	pps.BottomFieldPicOrderInFramePresent = flagField()
	pps.NumSliceGroupsMinus1, _ = readUe(nil)
	if pps.NumSliceGroupsMinus1 > 0 {
		pps.SliceGroupMapType, _ = readUe(nil)
		if pps.SliceGroupMapType == 0 {
			for iGroup := 0; iGroup <= pps.NumSliceGroupsMinus1; iGroup++ {
				pps.RunLengthMinus1[iGroup], _ = readUe(nil)
			}
		} else if pps.SliceGroupMapType == 2 {
			for iGroup := 0; iGroup < pps.NumSliceGroupsMinus1; iGroup++ {
				pps.TopLeft[iGroup], _ = readUe(nil)
				pps.BottomRight[iGroup], _ = readUe(nil)
			}
		} else if pps.SliceGroupMapType > 2 && pps.SliceGroupMapType < 6 {
			pps.SliceGroupChangeDirection = flagField()
			pps.SliceGroupChangeRateMinus1, _ = readUe(nil)
		} else if pps.SliceGroupMapType == 6 {
			pps.PicSizeInMapUnitsMinus1, _ = readUe(nil)
			for i := 0; i <= pps.PicSizeInMapUnitsMinus1; i++ {
				pps.SliceGroupId[i] = b.NextField(
					fmt.Sprintf("SliceGroupId[%d]", i),
					int(math.Ceil(math.Log2(float64(pps.NumSliceGroupsMinus1+1)))))
			}
		}

	}
	pps.NumRefIdxL0DefaultActiveMinus1, _ = readUe(nil)
	pps.NumRefIdxL1DefaultActiveMinus1, _ = readUe(nil)
	pps.WeightedPred = flagField()
	pps.WeightedBipred = b.NextField("WeightedBipredIDC", 2)
	pps.PicInitQpMinus26, _ = readSe(nil)
	pps.PicInitQsMinus26, _ = readSe(nil)
	pps.ChromaQpIndexOffset, _ = readSe(nil)
	pps.DeblockingFilterControlPresent = flagField()
	pps.ConstrainedIntraPred = flagField()
	pps.RedundantPicCntPresent = flagField()

	logger.Printf("debug: \tChecking for more PPS data")
	if b.HasMoreData() {
		logger.Printf("debug: \tProcessing additional PPS data")
		pps.Transform8x8Mode = b.NextField("Transform8x8ModeFlag", 1)
		pps.PicScalingMatrixPresent = flagField()
		if pps.PicScalingMatrixPresent {
			v := 6
			if sps.ChromaFormat != 3 {
				v = 2
			}
			for i := 0; i < 6+(v*pps.Transform8x8Mode); i++ {
				pps.PicScalingListPresent[i] = flagField()
				if pps.PicScalingListPresent[i] {
					if i < 6 {
						scalingList(
							b,
							ScalingList4x4[i],
							16,
							DefaultScalingMatrix4x4[i])

					} else {
						scalingList(
							b,
							ScalingList8x8[i],
							64,
							DefaultScalingMatrix8x8[i-6])

					}
				}
			}
			pps.SecondChromaQpIndexOffset, _ = readSe(nil)
		}
		b.MoreRBSPData()
		// rbspTrailingBits()
	}

	if showPacket {
		debugPacket("PPS", pps)
	}
	return &pps

}
