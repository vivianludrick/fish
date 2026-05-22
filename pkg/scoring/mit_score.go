package scoring

import "math"

// MIT off-target scoring (Hsu et al. 2013)
// Estimates likelihood of Cas9 cutting at an off-target site.
//
// Per-hit score = positionPenalty * distancePenalty * countPenalty * 100
// Aggregate specificity = 100 / (100 + sum(all hit scores))

// Position-dependent mismatch weights (positions 1-20, PAM-distal to PAM-proximal)
var mitWeights = [20]float64{
	0, 0, 0.014, 0, 0, 0.395, 0.317, 0, 0.389, 0.079,
	0.445, 0.508, 0.613, 0.851, 0.732, 0.828, 0.615, 0.804, 0.685, 0.583,
}

// MitHitScore computes the MIT score for a single off-target hit.
// guide and offTarget are 20bp sequences (PAM excluded).
// Returns score in range [0, 100].
func MitHitScore(guide, offTarget string) float64 {
	if len(guide) != 20 || len(offTarget) != 20 {
		return 0
	}

	var mismatchPositions []int
	for i := range 20 {
		if guide[i] != offTarget[i] {
			mismatchPositions = append(mismatchPositions, i)
		}
	}

	mismatchCount := len(mismatchPositions)
	if mismatchCount == 0 {
		return 100 // perfect match
	}

	// Score 1: product of (1 - weight) for each mismatch position
	positionPenalty := 1.0
	for _, pos := range mismatchPositions {
		positionPenalty *= (1 - mitWeights[pos])
	}

	// Score 2: mean pairwise distance penalty (only when >= 2 mismatches)
	distancePenalty := 1.0
	if mismatchCount >= 2 {
		totalDist := 0.0
		pairs := 0
		for i := 0; i < mismatchCount; i++ {
			for j := i + 1; j < mismatchCount; j++ {
				totalDist += math.Abs(float64(mismatchPositions[i] - mismatchPositions[j]))
				pairs++
			}
		}
		avgDist := totalDist / float64(pairs)
		distancePenalty = 1.0 / ((19.0-avgDist)/19.0*4.0 + 1.0)
	}

	// Score 3: mismatch count penalty
	countPenalty := 1.0 / float64(mismatchCount*mismatchCount)

	return positionPenalty * distancePenalty * countPenalty * 100
}

// MitAggregateSpecificity computes the aggregate specificity score for a guide
// across all its off-target hit scores.
// Returns value in range [0, 100]. Higher = more specific (fewer/weaker off-targets).
func MitAggregateSpecificity(hitScores []float64) float64 {
	sum := 0.0
	for _, s := range hitScores {
		sum += s
	}
	return 100.0 / (100.0 + sum)
}
