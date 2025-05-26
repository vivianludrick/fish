package pools

import (
	"strings"
	"sync"
	"vivalchemy/cris/internal/bitmaps"
	"vivalchemy/cris/pkg/models"
)

type SlidingWindow struct {
	Segments []uint64 // Array of bit patterns for each segment
	Config   *models.PatternConfig
}

var slidingWindowPool = sync.Pool{
	New: func() any {
		return &SlidingWindow{}
	},
}

// Configuration for pattern matching
var stringBuilderPool = sync.Pool{
	New: func() any {
		return &strings.Builder{}
	},
}

func NewSlidingWindow(config *models.PatternConfig) *SlidingWindow {
	sw := slidingWindowPool.Get().(*SlidingWindow)
	sw.Config = config
	sw.Segments = make([]uint64, len(config.Segments))
	return sw
}

func (sw *SlidingWindow) Release() {
	// config will remain constant throughout the runtime hence no need to release it
	sw.Segments = sw.Segments[:0]
	slidingWindowPool.Put(sw)
}

func (sw *SlidingWindow) AddNucleotide(nucleotide uint64) {
	segmentCount := len(sw.Config.Segments)
	var overflowBits uint64 = nucleotide

	// Unrolled loop for better performance when we have exactly 2 segments
	if segmentCount == 2 {
		// Process segment 1 (rightmost)
		segmentSize := sw.Config.Segments[1].Size
		maxBits := segmentSize * 4
		mask := (uint64(1) << maxBits) - 1
		sw.Segments[1] = (sw.Segments[1] << 4) | overflowBits
		overflowBits = sw.Segments[1] >> maxBits
		sw.Segments[1] &= mask

		// Process segment 0 (leftmost)
		segmentSize = sw.Config.Segments[0].Size
		maxBits = segmentSize * 4
		mask = (uint64(1) << maxBits) - 1
		sw.Segments[0] = (sw.Segments[0] << 4) | overflowBits
		sw.Segments[0] &= mask
		return
	}

	// General case for other segment counts
	for i := segmentCount - 1; i >= 0; i-- {
		segmentSize := sw.Config.Segments[i].Size
		maxBits := segmentSize * 4
		mask := (uint64(1) << maxBits) - 1
		sw.Segments[i] = (sw.Segments[i] << 4) | overflowBits
		overflowBits = sw.Segments[i] >> maxBits
		sw.Segments[i] &= mask
	}
}

func (sw *SlidingWindow) GetSequence() string {
	sb := stringBuilderPool.Get().(*strings.Builder)
	defer func() {
		sb.Reset()
		stringBuilderPool.Put(sb)
	}()

	for segmentIdx, segment := range sw.Segments[:len(sw.Config.Segments)] {
		segmentSize := sw.Config.Segments[segmentIdx].Size

		// Convert each nucleotide in the segment (from left to right)
		for i := segmentSize - 1; i >= 0; i-- {
			nucleotideBits := (segment >> (i * 4)) & 0xF
			sb.WriteByte(bitmaps.BitMapToNucleotide[nucleotideBits])
		}
	}

	return sb.String()
}
