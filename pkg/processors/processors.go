package processors

import (
	"math/bits"
	"strings"
	"vivalchemy/cris/internal/bitmaps"
	"vivalchemy/cris/pkg/models"
	"vivalchemy/cris/pkg/models/pools"
	"vivalchemy/cris/pkg/scoring"
)

var (
	ALLOW_N = false
)

// EncodedTarget holds a single encoded guide sequence ready for pattern matching.
type EncodedTarget struct {
	Sequence string   // original 23bp sequence (20bp guide + 3bp PAM)
	Encoded  []uint64 // bit-encoded segments
	Strand   string   // "forward" or "reverse"
}

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

// ProcessFastaRecordChunks searches chunks against all encoded targets.
// If exonIndex is non-nil, only matches overlapping exon regions are emitted.
func ProcessFastaRecordChunks(targets []EncodedTarget, config *models.PatternConfig, exonIndex *models.ExonIndex, fastaRecordChunksChan <-chan *pools.FastaRecordChunk, matchedPatternChan chan<- *pools.MatchedPattern) {
	for fastaRecordChunk := range fastaRecordChunksChan {
		if len(fastaRecordChunk.Sequence) < config.TotalSize {
			continue
		}

		// Extract seqname (first word of header) for exon lookup
		seqname := extractSeqname(fastaRecordChunk.Header)

		slidingWindow := pools.NewSlidingWindow(config)

		// Fill sliding window to TotalSize-1 so next AddNucleotide produces a full window
		for i := range config.TotalSize - 1 {
			slidingWindow.AddNucleotide(bitmaps.NucleotideToBitMap[fastaRecordChunk.Sequence[i]])
		}

		for i := config.TotalSize - 1; i < len(fastaRecordChunk.Sequence); i++ {
			nucleotide := fastaRecordChunk.Sequence[i]
			bitPattern := bitmaps.NucleotideToBitMap[nucleotide]
			slidingWindow.AddNucleotide(bitPattern)

			matchPos := fastaRecordChunk.Offset + i - config.TotalSize + 2 // 1-based

			// Exon filter: skip positions outside exons
			if exonIndex != nil {
				matchEnd := matchPos + config.TotalSize - 1
				if !exonIndex.OverlapsRange(seqname, matchPos, matchEnd) {
					continue
				}
			}

			for _, target := range targets {
				if ComparePatterns(slidingWindow.Segments, target.Encoded, config) {
					matchedPattern := pools.NewPatternMatch()
					matchedPattern.Header = fastaRecordChunk.Header
					matchedPattern.MatchedSequence = slidingWindow.GetSequence()
					matchedPattern.GuideSequence = target.Sequence
					matchedPattern.Offset = matchPos

					matchedPatternChan <- matchedPattern
				}
			}
		}
		slidingWindow.Release()
	}
}

// extractSeqname returns the first whitespace-delimited token from a FASTA header.
// e.g. "NC_066869.1 Labeo rohita strain..." -> "NC_066869.1"
func extractSeqname(header string) string {
	if idx := strings.IndexAny(header, " \t"); idx >= 0 {
		return header[:idx]
	}
	return header
}

func ProcessResults(matchedPatternChan <-chan *pools.MatchedPattern) {
	for match := range matchedPatternChan {
		scored := scoring.ScoreOffTarget(match.GuideSequence, match.MatchedSequence)
		scored.Header = match.Header
		scored.Offset = match.Offset
		scored.Println()
		match.Release()
	}
}
