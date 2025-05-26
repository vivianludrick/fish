package pools

import (
	"fmt"
	"sync"
)

type MatchedPattern struct {
	PAM    string `json:"pam"`
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

func (mp *MatchedPattern) ToResults() {
	fmt.Println(mp.Header)
	fmt.Println(mp.PAM)
	fmt.Println(mp.Offset)
}
