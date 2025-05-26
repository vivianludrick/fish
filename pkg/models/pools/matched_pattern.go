package pools

import (
	"fmt"
	"sync"
	"vivalchemy/cris/internal/utils"
)

type MatchedPattern struct {
	PAM    string `json:"pam"`
	Header []byte `json:"header"`
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
	mp.Header = []byte{}
	matchedPatternPool.Put(mp)
}

func (mp *MatchedPattern) ToResults() {
	utils.DebugRun("ToResults", func() {
		fmt.Println(string(mp.Header))
		fmt.Println(mp.PAM)
		fmt.Println(mp.Offset)
	})
}
