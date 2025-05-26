package main

import (
	"fmt"
	"os"
	"runtime"
	"sync"
	"vivalchemy/cris/internal/bitmaps"
	"vivalchemy/cris/internal/parsers"
	"vivalchemy/cris/internal/utils"
	"vivalchemy/cris/pkg/encoders"
	"vivalchemy/cris/pkg/models"
	"vivalchemy/cris/pkg/models/pools"
)

var (
	TARGET_FILE = "./test.fna"
	TARGET_PAM  = "CTAATAGGAGAGTATGCTGATGG"
	BUFFER_SIZE = 1024 * 1024 * 4 // 4MB buffer for better I/O performance
	CACHE_DIR   = ".pam_cache"
	numWorkers  = max(1, runtime.NumCPU()/2)
)

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

func processFastaRecord(targetPattern []uint64, config *models.PatternConfig, fastaRecordChan <-chan *pools.FastaRecord, matcedPatternChan chan<- *pools.MatchedPattern) {
	for fastaRecord := range fastaRecordChan {
		slidingWindow := pools.NewSlidingWindow(config)

		var i uint32
		for i = range uint32(len(fastaRecord.Sequence)) {
			nucleotide := fastaRecord.Sequence[i]
			// get the bit pattern for the current nucleotide
			bitPattern := bitmaps.NucleotideToBitMap[nucleotide]
			slidingWindow.AddNucleotide(bitPattern)

			if comparePatterns(slidingWindow.Segments, targetPattern, config) {
				matchedPattern := pools.NewPatternMatch()
				matchedPattern.Header = fastaRecord.Header
				matchedPattern.PAM = slidingWindow.GetSequence()
				matchedPattern.Offset = i - 21 // 1-based indexing, starting index at 22 spaces to left

				matcedPatternChan <- matchedPattern
			}
		}
		slidingWindow.Release()
	}
}

func processResults(matcedPatternChan <-chan *pools.MatchedPattern) {
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

	fastaRecordsChan := make(chan *pools.FastaRecord, numWorkers*2)
	matchedPatternChan := make(chan *pools.MatchedPattern, numWorkers*2)
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
