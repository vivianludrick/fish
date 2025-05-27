package processors

import (
	"math/bits"
	"vivalchemy/cris/internal/bitmaps"
	"vivalchemy/cris/pkg/models"
	"vivalchemy/cris/pkg/models/pools"
)

// TODO: Fetch the guide sequence here
var (
	GUIDE_SEQUENCE = "CTAATAGGAGAGTATGCTGATGG"
	ALLOW_N        = false
)

func ComparePatterns(pattern1 []uint64, pattern2 []uint64, config *models.PatternConfig) bool {
	totalMismatches := 0
	// check in reverse order
	for i := len(config.Segments) - 1; i >= 0; i-- {
		if ALLOW_N {
			matchingNibblesSet := pattern1[i] & pattern2[i]

			// collapse each nibble to 1 if any bit is set
			matchingNibblesSet |= (matchingNibblesSet >> 1) // 1000 -> 1100
			matchingNibblesSet |= (matchingNibblesSet >> 2) // 1100 -> 1111 shifted twice
			matchingNibblesSet &= 0x1111111111111111        // keep only the lsb

			// check if the number of mismatches is within the allowed range
			currentSegmentMismatches := config.Segments[i].Size - bits.OnesCount64(matchingNibblesSet)
			totalMismatches += currentSegmentMismatches
			if currentSegmentMismatches > config.Segments[i].AllowedMismatch || totalMismatches > config.MaxMismatchAllowed {
				return false
			}
		} else {
			matchingNibblesSet := pattern1[i] ^ pattern2[i]

			// collapse each nibble to 1 if any bit is set
			matchingNibblesSet |= (matchingNibblesSet >> 1) // 1000 -> 1100
			matchingNibblesSet |= (matchingNibblesSet >> 2) // 1100 -> 1111 shifted twice
			matchingNibblesSet &= 0x1111111111111111        // keep only the lsb

			// check if the number of mismatches is within the allowed range
			currentSegmentMismatches := bits.OnesCount64(matchingNibblesSet)
			totalMismatches += currentSegmentMismatches
			if currentSegmentMismatches > config.Segments[i].AllowedMismatch || totalMismatches > config.MaxMismatchAllowed {
				return false
			}
		}
	}
	return true
}

// func ComparePatterns2(pattern1 []uint64, pattern2 []uint64, config *models.PatternConfig) bool {
// 	totalMismatches := 0
//
// 	for i := len(config.Segments) - 1; i >= 0; i-- {
// 		currentSegmentMismatches := 0
// 		p1 := pattern1[i]
// 		p2 := pattern2[i]
//
// 		for nibble := range 16 {
// 			shift := uint(nibble * 4)
// 			n1 := (p1 >> shift) & 0xF
// 			n2 := (p2 >> shift) & 0xF
//
// 			if n1 != n2 {
// 				currentSegmentMismatches++
// 				totalMismatches++
//
// 				if currentSegmentMismatches > config.Segments[i].AllowedMismatch ||
// 					totalMismatches > config.MaxMismatchAllowed {
// 					return false
// 				}
// 			}
// 		}
// 	}
// 	return true
// }

func ProcessFastaRecordChunks(targetPattern []uint64, config *models.PatternConfig, fastaRecordChunksChan <-chan *pools.FastaRecordChunk, matcedPatternChan chan<- *pools.MatchedPattern) {
	for fastaRecordChunk := range fastaRecordChunksChan {
		// fmt.Println("Processing Record:", fastaRecordChunk.Header)
		if len(fastaRecordChunk.Sequence) < config.TotalSize {
			continue
		}

		slidingWindow := pools.NewSlidingWindow(config)

		// fillup the sliding window to fullsize-1 so that next time we add we get full and start comparin there and there
		// utils.DebugPrint("Processing Record: "+fastaRecord.Header, nil)
		for i := range config.TotalSize - 1 {
			slidingWindow.AddNucleotide(bitmaps.NucleotideToBitMap[fastaRecordChunk.Sequence[i]])
			// fmt.Println("Segments", slidingWindow.Segments, "\tSequence:", slidingWindow.GetSequence())
		}

		for i := config.TotalSize - 1; i < len(fastaRecordChunk.Sequence); i++ {
			nucleotide := fastaRecordChunk.Sequence[i]
			// get the bit pattern for the current nucleotide
			bitPattern := bitmaps.NucleotideToBitMap[nucleotide]
			slidingWindow.AddNucleotide(bitPattern)
			// fmt.Println("Processing Sequence: ", slidingWindow.GetSequence())

			// fmt.Println("Segments", slidingWindow.Segments, "\tSequence:", slidingWindow.GetSequence())
			if ComparePatterns(slidingWindow.Segments, targetPattern, config) {
				matchedPattern := pools.NewPatternMatch()
				matchedPattern.Header = fastaRecordChunk.Header
				matchedPattern.MatchedSequence = slidingWindow.GetSequence()
				matchedPattern.GuideSequence = GUIDE_SEQUENCE
				matchedPattern.Offset = fastaRecordChunk.Offset + i - config.TotalSize + 2 // 1-based indexing, starting index at 22 spaces to left

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
