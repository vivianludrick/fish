package extractors

import (
	"strings"
	"vivalchemy/cris/internal/bitmaps"
)

type ExtractedTarget struct {
	Sequence string // 23bp target (20bp guide + 3bp PAM in NGG form)
	Strand   string // "forward" or "reverse"
}

// ExtractTargetsFromBlob scans a DNA blob for all possible SpCas9 target sites.
//
// Forward strand: scan for NGG PAM — for each position where blob[i+guideLen+1]=='G' && blob[i+guideLen+2]=='G',
// extract blob[i : i+totalLen] as a 23bp target (20bp guide + NGG).
//
// Reverse strand: scan for CCN on the forward blob (= NGG on the reverse complement strand).
// For each position where blob[i]=='C' && blob[i+1]=='C', take blob[i : i+totalLen]
// and reverse-complement it to get the 23bp target with NGG at the 3' end.
//
// Deduplicates by sequence.
func ExtractTargetsFromBlob(blob string, guideLen int) []ExtractedTarget {
	blob = strings.ToUpper(strings.TrimSpace(blob))
	// Remove any whitespace/newlines within the blob
	blob = strings.Join(strings.Fields(blob), "")

	totalLen := guideLen + 3 // guide + PAM (NGG = 3bp)

	if len(blob) < totalLen {
		return nil
	}

	seen := make(map[string]bool)
	var targets []ExtractedTarget

	for i := 0; i <= len(blob)-totalLen; i++ {
		window := blob[i : i+totalLen]

		// Forward strand: [20bp guide][N][G][G]
		// Check positions guideLen+1 and guideLen+2 are both G
		if window[guideLen+1] == 'G' && window[guideLen+2] == 'G' {
			if isValidSequence(window) && !seen[window] {
				seen[window] = true
				targets = append(targets, ExtractedTarget{
					Sequence: window,
					Strand:   "forward",
				})
			}
		}

		// Reverse strand: [C][C][N][20bp] on forward blob = [20bp'][N'][G][G] on reverse
		// Check positions 0 and 1 are both C
		if window[0] == 'C' && window[1] == 'C' {
			rc := ReverseComplement(window)
			if isValidSequence(rc) && !seen[rc] {
				seen[rc] = true
				targets = append(targets, ExtractedTarget{
					Sequence: rc,
					Strand:   "reverse",
				})
			}
		}
	}

	return targets
}

// ReverseComplement returns the reverse complement of a DNA sequence.
func ReverseComplement(seq string) string {
	n := len(seq)
	result := make([]byte, n)
	for i := 0; i < n; i++ {
		result[n-1-i] = bitmaps.NucleotideComplementMap[seq[i]]
	}
	return string(result)
}

func isValidSequence(seq string) bool {
	for i := 0; i < len(seq); i++ {
		if bitmaps.NucleotideToBitMap[seq[i]] == 0 {
			return false
		}
	}
	return true
}
