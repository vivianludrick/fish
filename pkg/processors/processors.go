package processors

import (
	"math/bits"
	"vivalchemy/cris/internal/bitmaps"
	"vivalchemy/cris/internal/utils"
	"vivalchemy/cris/pkg/models"
	"vivalchemy/cris/pkg/models/pools"
)

const mask1InNibble = 0x1111111111111111

// Inlined and optimized N counting
func countNs(segment uint64) int {
	tmp := segment & (segment << 1) & mask1InNibble
	return bits.OnesCount64(tmp)
}

// Inlined nibble collapsing - kept as separate function for clarity but will be inlined by compiler
func consolidateMatchingBits(matchingNibblesSet uint64) uint64 {
	matchingNibblesSet |= (matchingNibblesSet >> 1)
	matchingNibblesSet |= (matchingNibblesSet >> 2)
	return matchingNibblesSet & mask1InNibble
}

func matchWithVariants(segment []uint64, variants [][]uint64, subSegmentLengths []int, checkNs bool) bool {
	for _, variant := range variants {
		if len(variant) != len(segment) {
			continue
		}
		matchedVariant := true
		for i, subVariant := range variant {
			guideSubSegment := segment[i]

			var matchingNibblesSet uint64
			var currentSegmentMismatches int

			if checkNs {
				matchingNibblesSet = subVariant & guideSubSegment
				// Inline nibble collapsing
				matchingNibblesSet = consolidateMatchingBits(matchingNibblesSet)
				currentSegmentMismatches = bits.OnesCount64(matchingNibblesSet)
			} else {
				matchingNibblesSet = subVariant ^ guideSubSegment
				// Inline nibble collapsing
				matchingNibblesSet = consolidateMatchingBits(matchingNibblesSet)
				currentSegmentMismatches = subSegmentLengths[i] - bits.OnesCount64(matchingNibblesSet)
			}
			if currentSegmentMismatches > 0 {
				matchedVariant = false
				break
			}
		}
		if matchedVariant {
			return true
		}

	}
	return false
}

func matchWithMismatches(guideSegment []uint64, segment []uint64, subSegmentLengths []int, allowedMismatches int, checkNs bool) int {
	if allowedMismatches == 0 {
		if matchWithVariants(guideSegment, [][]uint64{segment}, subSegmentLengths, checkNs) {
			return 0
		}
		return 1
	}

	var mismatchSoFarInSegment int
	for i := range segment {
		guideSubSegment := guideSegment[i]
		subSegment := segment[i]

		var matchingNibblesSet uint64
		if checkNs {
			matchingNibblesSet = guideSubSegment & subSegment
		} else {
			matchingNibblesSet = guideSubSegment ^ subSegment
		}

		// Inline nibble collapsing
		matchingNibblesSet = consolidateMatchingBits(matchingNibblesSet)

		// Calculate mismatches
		var currentSegmentMismatches int
		if checkNs {
			currentSegmentMismatches = bits.OnesCount64(matchingNibblesSet)
		} else {
			currentSegmentMismatches = subSegmentLengths[i] - bits.OnesCount64(matchingNibblesSet)
		}

		mismatchSoFarInSegment += currentSegmentMismatches
		if mismatchSoFarInSegment > allowedMismatches {
			return mismatchSoFarInSegment
		}
	}
	return mismatchSoFarInSegment
}

func validateGuideInternal(sw *pools.SlidingWindow, guideSequence []uint64, isReverse bool) bool {
	totalMismatches := 0
	currentSegmentIdx := 0

	// Pre-calculate boolean conditions to avoid repeated comparisons
	if sw.Config.AllowedNs > 0 {
		nsCount := 0
		for _, segment := range sw.Segments {
			nsCount += countNs(segment)
			if nsCount > sw.Config.AllowedNs {
				return false
			}
		}
	}

	for segmentIdx, segmentSpec := range sw.Config.SegmentSpec {
		if len(segmentSpec.AllowedVariants) != 0 {
			if isReverse {
				if !matchWithVariants(
					sw.ReverseSegments[currentSegmentIdx:currentSegmentIdx+len(segmentSpec.Lengths)],
					segmentSpec.AllowedReverseComplementVariants,
					utils.Reverse(sw.Config.SegmentSpec[len(sw.Config.SegmentSpec)-segmentIdx-1].Lengths),
					sw.Config.AllowedNs > 0) {
					return false
				}
				currentSegmentIdx += len(segmentSpec.Lengths)
			} else {
				if !matchWithVariants(sw.Segments[currentSegmentIdx:currentSegmentIdx+len(segmentSpec.Lengths)],
					segmentSpec.AllowedVariants,
					segmentSpec.Lengths,
					sw.Config.AllowedNs > 0) {
					return false
				}
				currentSegmentIdx += len(segmentSpec.Lengths)
			}
		} else {
			if isReverse {
				currentMismatches := matchWithMismatches(
					guideSequence[currentSegmentIdx:currentSegmentIdx+len(segmentSpec.Lengths)],
					sw.ReverseSegments[currentSegmentIdx:currentSegmentIdx+len(segmentSpec.Lengths)],
					utils.Reverse(sw.Config.SegmentSpec[len(sw.Config.SegmentSpec)-segmentIdx-1].Lengths),
					segmentSpec.AllowedMismatches,
					sw.Config.AllowedNs > 0)
				if currentMismatches > segmentSpec.AllowedMismatches {
					return false
				}
				totalMismatches += currentMismatches
				currentSegmentIdx += len(segmentSpec.Lengths)
			} else {
				currentMismatches := matchWithMismatches(
					guideSequence[currentSegmentIdx:currentSegmentIdx+len(segmentSpec.Lengths)],
					sw.Segments[currentSegmentIdx:currentSegmentIdx+len(segmentSpec.Lengths)],
					segmentSpec.Lengths,
					segmentSpec.AllowedMismatches,
					sw.Config.AllowedNs > 0)
				if currentMismatches > segmentSpec.AllowedMismatches {
					return false
				}
				totalMismatches += currentMismatches
				currentSegmentIdx += len(segmentSpec.Lengths)
			}
		}
	}

	return totalMismatches <= sw.Config.MaxTotalMismatches
}

// Unified validation function to reduce code duplication
// isReverse is used to reverse the segment order
// func _validateGuideInternal(sw *pools.SlidingWindow, guideSequence []uint64, isReverse bool) bool {
// 	totalMismatches := 0
// 	currentSegmentIdx := len(sw.Segments) - 1
// 	nsCount := 0
// 	totalSegments := len(sw.Segments)
//
// 	// Cache frequently accessed config values
// 	allowedNs := sw.Config.AllowedNs
// 	maxTotalMismatches := sw.Config.MaxTotalMismatches
// 	segmentSpecs := sw.Config.SegmentSpec
// 	segments := sw.Segments
//
// 	// Pre-calculate boolean conditions to avoid repeated comparisons
// 	checkNs := allowedNs > 0
//
// 	for segmentIdx := len(segmentSpecs) - 1; segmentIdx >= 0; segmentIdx-- {
// 		segment := segmentSpecs[segmentIdx]
//
// 		mismatchSoFarInSegment := 0
// 		for i := len(segment.Lengths) - 1; i >= 0; i-- {
// 			currentSegment := segments[currentSegmentIdx]
//
// 			// Optimize N counting - only when needed
// 			if checkNs {
// 				nsCount += countNs(currentSegment)
// 				if nsCount > allowedNs {
// 					return false
// 				}
// 			}
//
// 			// Calculate guide index based on direction
// 			var guideIdx int
// 			if isReverse {
// 				guideIdx = totalSegments - 1 - currentSegmentIdx
// 			} else {
// 				guideIdx = currentSegmentIdx
// 			}
//
// 			// Calculate matching nibbles
// 			var matchingNibblesSet uint64
// 			if checkNs {
// 				matchingNibblesSet = currentSegment & guideSequence[guideIdx]
// 			} else {
// 				matchingNibblesSet = currentSegment ^ guideSequence[guideIdx]
// 			}
//
// 			// Inline nibble collapsing
// 			matchingNibblesSet = consolidateMatchingBits(matchingNibblesSet)
//
// 			// Calculate mismatches
// 			var currentSegmentMismatches int
// 			if checkNs {
// 				currentSegmentMismatches = bits.OnesCount64(matchingNibblesSet)
// 			} else {
// 				currentSegmentMismatches = segment.Lengths[i] - bits.OnesCount64(matchingNibblesSet)
// 			}
//
// 			mismatchSoFarInSegment += currentSegmentMismatches
// 			if mismatchSoFarInSegment > segment.AllowedMismatches {
// 				return false
// 			}
// 			currentSegmentIdx--
// 		}
//
// 		totalMismatches += mismatchSoFarInSegment
// 		if totalMismatches > maxTotalMismatches {
// 			return false
// 		}
// 	}
//
// 	return true
// }

func ValidateGuide(sw *pools.SlidingWindow, guideSequence []uint64) bool {
	return validateGuideInternal(sw, guideSequence, false)
}

func ValidateReverseComplementGuide(sw *pools.SlidingWindow, reverseComplementGuideSequence []uint64) bool {
	return validateGuideInternal(sw, reverseComplementGuideSequence, true)
}

// Optimized pattern validation with early exit strategies
func FindMatchedGuideIndices(sw *pools.SlidingWindow) []int {
	guideSequences := sw.Config.GuideSequences
	reverseGuideSequences := sw.Config.ReverseGuideSequences
	numGuides := len(guideSequences)

	// Pre-allocate with known size
	matchedGuideIndices := make([]int, numGuides)
	matchedCount := 0

	// Check forward guides first with early exit if all matched
	for i := range numGuides {
		if ValidateGuide(sw, guideSequences[i]) {
			matchedGuideIndices[i] = 1
			matchedCount++
		}
	}

	// Only check reverse complement for unmatched guides
	// Early exit if all guides are already matched
	if matchedCount < numGuides {
		for i := range numGuides {
			if matchedGuideIndices[i] == 0 && ValidateReverseComplementGuide(sw, reverseGuideSequences[i]) {
				matchedGuideIndices[i] = -1
				matchedCount++
				// Early exit if all guides are now matched
				if matchedCount == numGuides {
					break
				}
			}
		}
	}

	return matchedGuideIndices
}

func ProcessFastaRecordChunks(config *models.NewInternalPatternSearchConfig, fastaRecordChunksChan <-chan *pools.FastaRecordChunk, matcedPatternChan chan<- *pools.MatchedPattern) {
	// Pre-calculate values outside the loop
	totalGuideLength := config.TotalGuideLength
	nucleotideToBitMap := bitmaps.NucleotideToBitMap

	for fastaRecordChunk := range fastaRecordChunksChan {
		sequenceLen := len(fastaRecordChunk.Sequence)

		// Early exit for short sequences
		if sequenceLen < totalGuideLength {
			continue
		}

		slidingWindow := pools.NewSlidingWindow(config)
		sequence := fastaRecordChunk.Sequence
		header := fastaRecordChunk.Header
		offset := fastaRecordChunk.Offset

		// Fill sliding window to fullsize-1
		for i := range totalGuideLength - 1 {
			slidingWindow.AddNucleotide(nucleotideToBitMap[sequence[i]])
		}

		// Process remaining sequence
		for i := totalGuideLength - 1; i < sequenceLen; i++ {
			nucleotide := sequence[i]
			bitPattern := nucleotideToBitMap[nucleotide]
			slidingWindow.AddNucleotide(bitPattern)

			matchedGuideIndices := FindMatchedGuideIndices(slidingWindow)

			// Process matches with reduced allocations
			for matchedGuideIndex, matchedGuideStatus := range matchedGuideIndices {
				if matchedGuideStatus != 0 { // Simplified condition
					matchedPattern := pools.NewPatternMatch().
						SetHeader(header).
						SetMatchedSequence(slidingWindow.Segments).
						SetGuideSequence(matchedGuideIndex * matchedGuideStatus).
						SetOffset(offset + i - totalGuideLength + 2) // 1-based indexing
					matcedPatternChan <- matchedPattern
				}
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
