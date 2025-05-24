package main

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"
	"vivalchemy/cris/internal/parsers"
	"vivalchemy/cris/internal/utils"
	"vivalchemy/cris/pkg/encoders"
	"vivalchemy/cris/pkg/models"
)

var (
	TARGET_FILE = "./genomes/GCA_015244755.2_TenIli1.0_genomic.fna"
	TARGET_PAM  = "CTCCTGTATTTAGGAGGCTCNGG"
	BUFFER_SIZE = 1024 * 1024 * 4 // 4MB buffer for better I/O performance
	CACHE_DIR   = ".pam_cache"
	numWorkers  = max(1, runtime.NumCPU()/2)
)

var bitMapArray [256]uint64
var reverseBitMapArray [16]byte

func init() {
	// Initialize lookup arrays
	bitMapArray['A'] = 0x8 // 1000
	bitMapArray['C'] = 0x4 // 0100
	bitMapArray['G'] = 0x2 // 0010
	bitMapArray['T'] = 0x1 // 0001
	bitMapArray['N'] = 0xF // 1111 (ambiguous nucleotide)
	bitMapArray['a'] = 0x8 // Support lowercase
	bitMapArray['c'] = 0x4
	bitMapArray['g'] = 0x2
	bitMapArray['t'] = 0x1
	bitMapArray['n'] = 0xF

	reverseBitMapArray[0x8] = 'A'
	reverseBitMapArray[0x4] = 'C'
	reverseBitMapArray[0x2] = 'G'
	reverseBitMapArray[0x1] = 'T'
	reverseBitMapArray[0xF] = 'N'
}

// Configuration for pattern matching

type MatchedPattern struct {
	PAM    string `json:"pam"`
	Header string `json:"header"`
	Offset uint32 `json:"offset"`
}

var matchedPatternPool = sync.Pool{
	New: func() any {
		return &MatchedPattern{}
	},
}

func NewPatternMatch() *MatchedPattern {
	return matchedPatternPool.Get().(*MatchedPattern)
}

func (mp *MatchedPattern) Release() {
	mp.PAM = ""
	mp.Header = ""
	matchedPatternPool.Put(mp)
}

func (mp *MatchedPattern) ToResults() {
	fmt.Println(mp.Header)
	fmt.Println(mp.PAM)
	fmt.Println(mp.Offset)
}

type SlidingWindow struct {
	segments []uint64 // Array of bit patterns for each segment
	config   *models.PatternConfig
}

var slidingWindowPool = sync.Pool{
	New: func() any {
		return &SlidingWindow{}
	},
}

var stringBuilderPool = sync.Pool{
	New: func() any {
		return &strings.Builder{}
	},
}

func NewSlidingWindow(config *models.PatternConfig) *SlidingWindow {
	sw := slidingWindowPool.Get().(*SlidingWindow)
	sw.config = config
	sw.segments = make([]uint64, len(config.Segments))
	return sw
}

func (sw *SlidingWindow) Release() {
	// config will remain constant throughout the runtime hence no need to release it
	sw.segments = sw.segments[:0]
	slidingWindowPool.Put(sw)
}

func (sw *SlidingWindow) AddNucleotide(nucleotide uint64) {
	segmentCount := len(sw.config.Segments)
	var overflowBits uint64 = nucleotide

	// Unrolled loop for better performance when we have exactly 2 segments
	if segmentCount == 2 {
		// Process segment 1 (rightmost)
		segmentSize := sw.config.Segments[1].Size
		maxBits := segmentSize * 4
		mask := (uint64(1) << maxBits) - 1
		sw.segments[1] = (sw.segments[1] << 4) | overflowBits
		overflowBits = sw.segments[1] >> maxBits
		sw.segments[1] &= mask

		// Process segment 0 (leftmost)
		segmentSize = sw.config.Segments[0].Size
		maxBits = segmentSize * 4
		mask = (uint64(1) << maxBits) - 1
		sw.segments[0] = (sw.segments[0] << 4) | overflowBits
		sw.segments[0] &= mask
		return
	}

	// General case for other segment counts
	for i := segmentCount - 1; i >= 0; i-- {
		segmentSize := sw.config.Segments[i].Size
		maxBits := segmentSize * 4
		mask := (uint64(1) << maxBits) - 1
		sw.segments[i] = (sw.segments[i] << 4) | overflowBits
		overflowBits = sw.segments[i] >> maxBits
		sw.segments[i] &= mask
	}
}

func (sw *SlidingWindow) getSequence() string {
	sb := stringBuilderPool.Get().(*strings.Builder)
	defer func() {
		sb.Reset()
		stringBuilderPool.Put(sb)
	}()

	for segmentIdx, segment := range sw.segments[:len(sw.config.Segments)] {
		segmentSize := sw.config.Segments[segmentIdx].Size

		// Convert each nucleotide in the segment (from left to right)
		for i := segmentSize - 1; i >= 0; i-- {
			nucleotideBits := (segment >> (i * 4)) & 0xF
			sb.WriteByte(reverseBitMapArray[nucleotideBits])
		}
	}

	return sb.String()
}

func comparePatterns(pattern1 []uint64, pattern2 []uint64, config *models.PatternConfig) bool {
	totalMismatches := 0
	// checking from back
	for i := len(config.Segments) - 1; i >= 0; i-- {
		currentSegmentMismatches := 0
		xor := pattern1[i] ^ pattern2[i]
		// early exit if no mismatches
		if xor == 0 {
			continue
		}

		// since it is 4-bit we have to and it and then shift
		firstBits := xor & 0x8888_8888_8888_8888  // 1000s
		secondBits := xor & 0x4444_4444_4444_4444 // 0100s
		thirdBits := xor & 0x2222_2222_2222_2222  // 0010s
		fourthBits := xor & 0x1111_1111_1111_1111 // 0001s
		xor = firstBits | (secondBits << 1) | (thirdBits << 2) | (fourthBits << 3)

		for xor != 0 {
			currentSegmentMismatches++
			xor &= xor - 1
		}

		totalMismatches += currentSegmentMismatches
		if currentSegmentMismatches > config.Segments[i].AllowedMismatch || totalMismatches > config.MaxMismatchAllowed {
			return false
		}
	}
	return true
}

func processFastaRecord(targetPattern []uint64, config *models.PatternConfig, fastaRecordChan <-chan *models.FastaRecord, matcedPatternChan chan<- *MatchedPattern) {
	for fastaRecord := range fastaRecordChan {
		slidingWindow := NewSlidingWindow(config)

		var i uint32
		for i = range uint32(len(fastaRecord.Sequence)) {
			nucleotide := fastaRecord.Sequence[i]
			// get the bit pattern for the current nucleotide
			bitPattern := bitMapArray[nucleotide]
			slidingWindow.AddNucleotide(bitPattern)

			if comparePatterns(slidingWindow.segments, targetPattern, config) {
				matchedPattern := NewPatternMatch()
				matchedPattern.Header = fastaRecord.Header
				matchedPattern.PAM = slidingWindow.getSequence()
				matchedPattern.Offset = i + 1 // 1-based indexing

				matcedPatternChan <- matchedPattern
			}
		}
		slidingWindow.Release()
	}
}

func processResults(matcedPatternChan <-chan *MatchedPattern) {
	for matcedPattern := range matcedPatternChan {
		matcedPattern.ToResults()
		matcedPattern.Release()
	}
}

func cliArgumentInitialization() {
	if len(os.Args) > 1 {
		TARGET_FILE = os.Args[1]
	}
	if len(os.Args) > 2 {
		TARGET_PAM = os.Args[2]
	}
	if len(os.Args) > 3 {
		CACHE_DIR = os.Args[3]
	}
}

func main() {
	cliArgumentInitialization()

	patternConfig := &models.PatternConfig{
		Segments: []*models.SegmentConfig{
			{Size: 13, AllowedMismatch: 4},
			{Size: 10, AllowedMismatch: 2},
		},
		MaxMismatchAllowed: 4,
		TotalSize:          23,
	}
	targetPattern, err := encoders.SegmentAndEncodePattern(TARGET_PAM, patternConfig)
	if err != nil {
		fmt.Println(err)
		return
	}

	var processorWg sync.WaitGroup
	var parserWg sync.WaitGroup

	fastaRecordsChan := make(chan *models.FastaRecord, numWorkers*2)
	matchedPatternChan := make(chan *MatchedPattern, numWorkers*2)
	// aggregatedResultsChan := make(chan *PatternResults, 1)

	parserWg.Add(1)
	go utils.TimeFunction("Parse FASTA File", func() {
		defer close(fastaRecordsChan)
		parsers.ParseFastaFileOrDecodeCache(TARGET_FILE, CACHE_DIR, BUFFER_SIZE, fastaRecordsChan)
		parserWg.Done()
	})

	for range numWorkers {
		processorWg.Add(1)
		go func() {
			processFastaRecord(targetPattern, patternConfig, fastaRecordsChan, matchedPatternChan)
			processorWg.Done()
		}()
	}

	go func() {
		defer close(matchedPatternChan)
		processorWg.Wait()
	}()
	processResults(matchedPatternChan)

	// waiters
	parserWg.Wait()
}
