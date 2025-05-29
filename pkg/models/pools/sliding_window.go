package pools

import (
	"strings"
	"sync"
	"vivalchemy/cris/internal/bitmaps"
	"vivalchemy/cris/pkg/models"
)

type SlidingWindow struct {
	Segments        []uint64 // Array of bit patterns for each segment
	ReverseSegments []uint64 // Segmenting the data in reverse order
	Config          *models.NewInternalPatternSearchConfig
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

func NewSlidingWindow(config *models.NewInternalPatternSearchConfig) *SlidingWindow {
	sw := slidingWindowPool.Get().(*SlidingWindow)
	sw.Config = config
	sw.Segments = make([]uint64, config.GetNumberOfSegments())
	sw.ReverseSegments = make([]uint64, config.GetNumberOfSegments())
	return sw
}

func (sw *SlidingWindow) Release() {
	// config will remain constant throughout the runtime hence no need to release it
	sw.Segments = sw.Segments[:0]
	sw.ReverseSegments = sw.ReverseSegments[:0]
	slidingWindowPool.Put(sw)
}

func (sw *SlidingWindow) AddNucleotide(nucleotide uint64) {
	segmentCount := len(sw.Segments)
	var overflowBits uint64 = nucleotide

	// General case for other segment counts
	for i := len(sw.Config.SegmentSpec) - 1; i >= 0; i-- {
		for j := len(sw.Config.SegmentSpec[i].Lengths) - 1; j >= 0; j-- {
			segmentSize := sw.Config.SegmentSpec[i].Lengths[j]
			maxBits := segmentSize * 4
			mask := (uint64(1) << maxBits) - 1
			sw.Segments[segmentCount] = (sw.Segments[segmentCount] << 4) | overflowBits
			overflowBits = sw.Segments[segmentCount] >> maxBits
			sw.Segments[segmentCount] &= mask
			segmentCount--
		}
	}

	segmentCount = len(sw.ReverseSegments)
	overflowBits = nucleotide
	// take the first length and add it at last
	for i := range sw.Config.SegmentSpec {
		for j := range sw.Config.SegmentSpec[i].Lengths {
			segmentSize := sw.Config.SegmentSpec[i].Lengths[j]
			maxBits := segmentSize * 4
			mask := (uint64(1) << maxBits) - 1
			sw.ReverseSegments[segmentCount] = (sw.ReverseSegments[segmentCount] << 4) | overflowBits
			overflowBits = sw.ReverseSegments[segmentCount] >> maxBits
			sw.ReverseSegments[segmentCount] &= mask
			segmentCount--
		}
	}
}

func (sw *SlidingWindow) GetSequence() string {
	sb := stringBuilderPool.Get().(*strings.Builder)
	defer func() {
		sb.Reset()
		stringBuilderPool.Put(sb)
	}()

	segmentIndex := 0
	for _, segment := range sw.Config.SegmentSpec {
		for _, subSegmentLength := range segment.Lengths {
			for i := range subSegmentLength {
				nucleotideBits := (sw.Segments[segmentIndex] >> (i * 4)) & 0xF
				sb.WriteByte(bitmaps.BitMapToNucleotide[nucleotideBits])
				segmentIndex++
			}
		}
	}

	return sb.String()
}
