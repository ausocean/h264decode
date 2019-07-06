package h264

import (
	"io"
	"os"
)

type H264Reader struct {
	IsStarted    bool
	Stream       io.Reader
	NalUnits     []*BitReader
	VideoStreams []*VideoStream
	DebugFile    *os.File
	*BitReader
}

func (h *H264Reader) BufferToReader(cntBytes int) error {
	buf := make([]byte, cntBytes)
	if _, err := h.Stream.Read(buf); err != nil {
		logger.Printf("error: while reading %d bytes: %v\n", cntBytes, err)
		return err
	}
	h.bytes = append(h.bytes, buf...)
	if h.DebugFile != nil {
		h.DebugFile.Write(buf)
	}
	h.byteOffset += cntBytes
	return nil
}

func (h *H264Reader) Discard(cntBytes int) error {
	buf := make([]byte, cntBytes)
	if _, err := h.Stream.Read(buf); err != nil {
		logger.Printf("error: while discarding %d bytes: %v\n", cntBytes, err)
		return err
	}
	h.byteOffset += cntBytes
	return nil
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

func (h *H264Reader) Start() {
	for {
		nalUnit, _ := h.readNalUnit()
		switch nalUnit.Type {
		case NALU_TYPE_SPS:
			// TODO: handle this error
			sps, _ := NewSPS(nalUnit.rbsp, false)
			h.VideoStreams = append(
				h.VideoStreams,
				&VideoStream{SPS: sps},
			)
		case NALU_TYPE_PPS:
			videoStream := h.VideoStreams[len(h.VideoStreams)-1]
			// TODO: handle this error
			videoStream.PPS, _ = NewPPS(videoStream.SPS, nalUnit.RBSP(), false)
		case NALU_TYPE_SLICE_IDR_PICTURE:
			fallthrough
		case NALU_TYPE_SLICE_NON_IDR_PICTURE:
			videoStream := h.VideoStreams[len(h.VideoStreams)-1]
			logger.Printf("info: frame number %d\n", len(videoStream.Slices))
			// TODO: handle this error
			sliceContext, _ := NewSliceContext(videoStream, nalUnit, nalUnit.RBSP(), true)
			videoStream.Slices = append(videoStream.Slices, sliceContext)
		}
	}
}

func (r *H264Reader) readNalUnit() (*NalUnit, *BitReader) {
	// Read to start of NAL
	logger.Printf("debug: Seeking NAL %d start\n", len(r.NalUnits))
	r.LogStreamPosition()
	for !isStartSequence(r.Bytes()) {
		if err := r.BufferToReader(1); err != nil {
			return nil, nil
		}
	}
	/*
		if !r.IsStarted {
			logger.Printf("debug: skipping initial NAL zero byte spaces\n")
			r.LogStreamPosition()
			// Annex B.2 Step 1
			if err := r.Discard(1); err != nil {
				logger.Printf("error: while discarding empty byte (Annex B.2:1): %v\n", err)
				return nil
			}
			if err := r.Discard(2); err != nil {
				logger.Printf("error: while discarding start code prefix one 3bytes (Annex B.2:2): %v\n", err)
				return nil
			}
		}
	*/
	_, startOffset, _ := r.StreamPosition()
	logger.Printf("debug: Seeking next NAL start\n")
	r.LogStreamPosition()
	// Read to start of next NAL
	_, so, _ := r.StreamPosition()
	for so == startOffset || !isStartSequence(r.Bytes()) {
		_, so, _ = r.StreamPosition()
		if err := r.BufferToReader(1); err != nil {
			return nil, nil
		}
	}
	// logger.Printf("debug: PreRewind %#v\n", r.Bytes())
	// Rewind back the length of the start sequence
	// r.RewindBytes(4)
	// logger.Printf("debug: PostRewind %#v\n", r.Bytes())
	_, endOffset, _ := r.StreamPosition()
	logger.Printf("debug: found NAL unit with %d bytes from %d to %d\n", endOffset-startOffset, startOffset, endOffset)
	nalUnitReader := &BitReader{bytes: r.Bytes()[startOffset:]}
	r.NalUnits = append(r.NalUnits, nalUnitReader)
	r.LogStreamPosition()
	logger.Printf("debug: NAL Header: %#v\n", nalUnitReader.Bytes()[0:8])
	nalUnit := NewNalUnit(nalUnitReader.Bytes(), len(nalUnitReader.Bytes()))
	return nalUnit, nalUnitReader
}

func isStartSequence(packet []byte) bool {
	if len(packet) < len(InitialNALU) {
		return false
	}
	naluSegment := packet[len(packet)-4:]
	for i := range InitialNALU {
		if naluSegment[i] != InitialNALU[i] {
			return false
		}
	}
	return true
}

func isStartCodeOnePrefix(buf []byte) bool {
	for i, b := range buf {
		if i < 2 && b != byte(0) {
			return false
		}
		// byte 3 may be 0 or 1
		if i == 3 && b != byte(0) || b != byte(1) {
			return false
		}
	}
	logger.Printf("debug: found start code one prefix byte\n")
	return true
}

func isEmpty3Byte(buf []byte) bool {
	if len(buf) < 3 {
		return false
	}
	for _, i := range buf[len(buf)-3:] {
		if i != 0 {
			return false
		}
	}
	return true
}
