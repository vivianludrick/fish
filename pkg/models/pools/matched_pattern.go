package pools

import (
	"fmt"
	"sync"
)

// MatchedPattern is a struct that holds the details of the matched pattern
type MatchedPattern struct {
	// The index of the guide sequence in the internalConfig with which the pattern
	// was matched
	GuideSequence int `json:"guide_sequence"`
	// The uint64 representation of the matched pattern segmented based on the
	// internalConfig segment lengths
	MatchedSequence []uint64 `json:"matched_sequence"`
	// The header of the fasta record that was matched
	Header string `json:"header"`
	// The offset of the matched pattern in that particular fasta record
	Offset int `json:"offset"`
	// Whether the matched pattern is a reverse complement of the guide sequence
	ReverseComplement bool `json:"reverse_complement"`
}

var matchedPatternPool = sync.Pool{
	New: func() any {
		return &MatchedPattern{
			MatchedSequence: make([]uint64, 0),
		}
	},
}

// NewPatternMatch returns a new MatchedPattern from the pool
func NewPatternMatch() *MatchedPattern {
	return matchedPatternPool.Get().(*MatchedPattern)
}

// SetGuideSequence sets the guide sequence index of the matched pattern
func (mp *MatchedPattern) SetGuideSequence(guideSequence int) *MatchedPattern {
	mp.GuideSequence = guideSequence
	return mp
}

// SetMatchedSequence sets the matched sequence of the matched pattern
func (mp *MatchedPattern) SetMatchedSequence(matchedSequence []uint64) *MatchedPattern {
	mp.MatchedSequence = matchedSequence
	return mp
}

// SetHeader sets the header of the matched pattern
func (mp *MatchedPattern) SetHeader(header string) *MatchedPattern {
	mp.Header = header
	return mp
}

// SetOffset sets the offset of the matched pattern
func (mp *MatchedPattern) SetOffset(offset int) *MatchedPattern {
	mp.Offset = offset
	return mp
}

// SetReverseComplement sets the reverse complement of the matched pattern
func (mp *MatchedPattern) SetReverseComplement(reverseComplement bool) *MatchedPattern {
	mp.ReverseComplement = reverseComplement
	return mp
}

// Release the MatchedPattern back to the pool
func (mp *MatchedPattern) Release() {
	mp.MatchedSequence = mp.MatchedSequence[:0]
	mp.Header = ""
	matchedPatternPool.Put(mp)
}

// Println pretty prints the MatchedPattern
func (mp *MatchedPattern) Println() {
	fmt.Println(string(mp.Header))
	fmt.Println(mp.MatchedSequence)
	fmt.Println(mp.Offset)
}
