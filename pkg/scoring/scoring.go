package scoring

import "fmt"

// ScoredMatch holds a matched off-target site with computed scores.
type ScoredMatch struct {
	GuideSequence   string  `json:"guide_sequence"`   // full guide+PAM (e.g. 23bp)
	MatchedSequence string  `json:"matched_sequence"`  // full off-target+PAM
	Header          string  `json:"header"`
	Offset          int     `json:"offset"`
	MitScore        float64 `json:"mit_score"`  // 0-100, higher = more likely to cut
	CfdScore        float64 `json:"cfd_score"`  // 0-1, higher = more likely to cut
}

// ScoreOffTarget computes MIT and CFD scores for a guide vs off-target match.
// Both sequences include the PAM (last 3bp). Guide portion = first (len-3) bases.
func ScoreOffTarget(guideSeq, matchedSeq string) ScoredMatch {
	sm := ScoredMatch{
		GuideSequence:   guideSeq,
		MatchedSequence: matchedSeq,
	}

	seqLen := len(guideSeq)
	if seqLen < 5 || len(matchedSeq) != seqLen {
		return sm
	}

	guideCore := guideSeq[:seqLen-3]    // 20bp guide without PAM
	matchCore := matchedSeq[:seqLen-3]   // 20bp off-target without PAM

	// PAM dinucleotide from the off-target (last 2 chars)
	pamDinuc := matchedSeq[seqLen-2:]

	// Truncate/pad to 20bp for scoring (MIT/CFD expect exactly 20)
	if len(guideCore) > 20 {
		guideCore = guideCore[len(guideCore)-20:]
		matchCore = matchCore[len(matchCore)-20:]
	}

	sm.MitScore = MitHitScore(guideCore, matchCore)
	sm.CfdScore = CfdScore(guideCore, matchCore, pamDinuc)

	return sm
}

func (sm *ScoredMatch) Println() {
	fmt.Printf("%s\t%s\toffset:%d\tMIT:%.2f\tCFD:%.4f\n",
		sm.Header, sm.MatchedSequence, sm.Offset, sm.MitScore, sm.CfdScore)
}
