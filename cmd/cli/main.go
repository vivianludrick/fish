package main

import (
	"fmt"
	"os"
	"runtime"
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
	PAM    string `json:"PAM"`
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

type SlidingWindow struct {
	segments        []uint64 // Array of bit patterns for each segment
	nucleotideCount uint64   // Track how many nucleotides we've processed
	config          *models.PatternConfig
}

var slidingWindowPool = sync.Pool{
	New: func() any {
		return &SlidingWindow{}
	},
}

func NewSlidingWindow() *SlidingWindow {
	return slidingWindowPool.Get().(*SlidingWindow)
}

func (sw *SlidingWindow) Release() {
	sw.segments = sw.segments[:0]
	sw.nucleotideCount = 0
	slidingWindowPool.Put(sw)
}

func (sw *SlidingWindow) AddNucleotide(nucleotide uint64) {
	segmentCount := len(sw.config.Segments)

	var overflowBits uint64 = nucleotide // new nucleotide comes in as lowest 4 bits

	// Process each segment from right to left
	for i := segmentCount - 1; i >= 0; i-- {
		segmentSize := sw.config.Segments[i].Size
		maxBits := segmentSize * 4
		mask := (uint64(1) << maxBits) - 1
		// Shift current segment left by 4, add overflow bits
		sw.segments[i] = (sw.segments[i] << 4) | overflowBits
		// Compute overflow for next (left) segment
		overflowBits = sw.segments[i] >> (maxBits)
		// Mask current segment to fit within its allowed size
		sw.segments[i] &= mask
	}
}

func processFastaRecord(targetPattern []uint64, config *models.PatternConfig, fastaRecordChan <-chan *models.FastaRecord, matcedPatternChan chan<- *MatchedPattern) {
	for fastaRecord := range fastaRecordChan {

		for nucleotide := range fastaRecord.Sequence {
			// get the bit pattern for the current nucleotide
			bitPattern := bitMapArray[nucleotide]

		}
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
		Segments: []models.SegmentConfig{
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
		parsers.ParseFastaFileOrDecodeCache(TARGET_FILE, CACHE_DIR, BUFFER_SIZE, fastaRecordsChan)
		defer close(fastaRecordsChan)
		parserWg.Done()
	})

	for range numWorkers {
		processorWg.Add(1)
		go func() {
			processFastaRecord(targetPattern, nil, fastaRecordsChan, matchedPatternChan)
			processorWg.Done()
		}()
	}

	processorWg.Wait()
	parserWg.Wait()
	var _ = printPattern
}

// ---
func printPattern(pattern []uint64) {
	var str string
	for _, segment := range pattern {
		str += fmt.Sprintf("%016b", segment)
	}
	fmt.Println(str)
}
