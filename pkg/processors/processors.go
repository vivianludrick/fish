package processors

import (
	"vivalchemy/cris/internal/bitmaps"
	"vivalchemy/cris/pkg/models"
	"vivalchemy/cris/pkg/models/pools"
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
		slidingWindow := pools.NewSlidingWindow(config)

		var i uint32
		for i = range uint32(len(fastaRecord.Sequence)) {
			nucleotide := fastaRecord.Sequence[i]
			// get the bit pattern for the current nucleotide
			bitPattern := bitmaps.NucleotideToBitMap[nucleotide]
			slidingWindow.AddNucleotide(bitPattern)

			if ComparePatterns(slidingWindow.Segments, targetPattern, config) {
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

func ProcessResults(matcedPatternChan <-chan *pools.MatchedPattern) {
	for matcedPattern := range matcedPatternChan {
		matcedPattern.ToResults()
		matcedPattern.Release()
	}
}
