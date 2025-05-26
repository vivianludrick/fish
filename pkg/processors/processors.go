package processors

import (
	"fmt"
	"vivalchemy/cris/internal/bitmaps"
	"vivalchemy/cris/internal/utils"
	"vivalchemy/cris/pkg/models"
	"vivalchemy/cris/pkg/models/pools"
)

// TODO: Fetch the guide sequence here
var (
	GUIDE_SEQUENCE = "CTAATAGGAGAGTATGCTGATGG"
)

func ComparePatterns(pattern1 []uint64, pattern2 []uint64, config *models.PatternConfig) bool {
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

func ProcessFastaRecord(targetPattern []uint64, config *models.PatternConfig, fastaRecordChan <-chan *pools.FastaRecord, matcedPatternChan chan<- *pools.MatchedPattern) {
	for fastaRecord := range fastaRecordChan {
		if len(fastaRecord.Sequence) < config.TotalSize {
			continue
		}

		slidingWindow := pools.NewSlidingWindow(config)

		// fillup the sliding window to fullsize-1 so that next time we add we get full and start comparin there and there
		utils.DebugPrint("Processing Record: "+fastaRecord.Header, nil)
		for i := range config.TotalSize - 1 {
			slidingWindow.AddNucleotide(bitmaps.NucleotideToBitMap[fastaRecord.Sequence[i]])
			fmt.Println("Segments", slidingWindow.Segments, "\tSequence:", slidingWindow.GetSequence())
		}

		for i := config.TotalSize - 1; i < len(fastaRecord.Sequence); i++ {
			nucleotide := fastaRecord.Sequence[i]
			// get the bit pattern for the current nucleotide
			bitPattern := bitmaps.NucleotideToBitMap[nucleotide]
			slidingWindow.AddNucleotide(bitPattern)

			fmt.Println("Segments", slidingWindow.Segments, "\tSequence:", slidingWindow.GetSequence())
			if ComparePatterns(slidingWindow.Segments, targetPattern, config) {
				matchedPattern := pools.NewPatternMatch()
				matchedPattern.Header = fastaRecord.Header
				matchedPattern.MatchedSequence = slidingWindow.GetSequence()
				matchedPattern.GuideSequence = GUIDE_SEQUENCE
				matchedPattern.Offset = i - config.TotalSize + 2 // 1-based indexing, starting index at 22 spaces to left

				matcedPatternChan <- matchedPattern
			}
		}
		slidingWindow.Release()
	}
}

func ProcessResults(matcedPatternChan <-chan *pools.MatchedPattern) {
	for matcedPattern := range matcedPatternChan {
		matcedPattern.Println()
		matcedPattern.Release()
	}
}
