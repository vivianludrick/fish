package pools

import (
	"fmt"
	"sync"
)

type MatchedPattern struct {
	GuideSequence   string `json:"guide_sequence"`
	MatchedSequence string `json:"matched_sequence"`
	Header          string `json:"header"`
	Offset          int    `json:"offset"`
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
	mp.MatchedSequence = ""
	mp.Header = ""
	matchedPatternPool.Put(mp)
}

func (mp *MatchedPattern) Println() {
	fmt.Println(string(mp.Header))
	fmt.Println(mp.MatchedSequence)
	fmt.Println(mp.Offset)
}
