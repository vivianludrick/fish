package scoring

// CFD off-target scoring (Doench et al. 2016)
// Cutting Frequency Determination score.
// Uses position-specific and mismatch-type-specific penalties.
//
// Score = product(mismatch_penalty[pos]) * pam_penalty
// Only mismatch positions contribute; matched positions = 1.0.
//
// Guide and off-target are compared as DNA sequences (same strand).
// Guide DNA T → RNA U for matrix lookup (rX:dY notation).

// cfdMismatchKey encodes guide RNA base + off-target DNA base + position (1-indexed)
type cfdMismatchKey struct {
	guideRNA    byte // 'A','U','G','C'
	offTargetDN byte // 'A','T','G','C'
	position    int  // 1-20
}

// cfgMismatchScores contains the 240 position-specific mismatch penalties.
// Key: (RNA guide base, DNA off-target base, 1-indexed position)
// Only mismatches are stored. Watson-Crick complement cases (rA:dT, rU:dA, rG:dC, rC:dG)
// are absent — those get score 1.0 (RNA still binds via WC pairing).
var cfdMismatchScores map[cfdMismatchKey]float64

// cfdPAMScores contains the 16 PAM dinucleotide penalties.
// Key: 2-char PAM string (positions 2-3 of NGG).
var cfdPAMScores map[string]float64

func init() {
	cfdMismatchScores = map[cfdMismatchKey]float64{
		// rA:dA (guide A, off-target A)
		{'A', 'A', 1}: 1.0, {'A', 'A', 2}: 0.727272727, {'A', 'A', 3}: 0.705882353,
		{'A', 'A', 4}: 0.636363636, {'A', 'A', 5}: 0.363636364, {'A', 'A', 6}: 0.714285714,
		{'A', 'A', 7}: 0.4375, {'A', 'A', 8}: 0.428571429, {'A', 'A', 9}: 0.6,
		{'A', 'A', 10}: 0.882352941, {'A', 'A', 11}: 0.307692308, {'A', 'A', 12}: 0.333333333,
		{'A', 'A', 13}: 0.3, {'A', 'A', 14}: 0.533333333, {'A', 'A', 15}: 0.2,
		{'A', 'A', 16}: 0.0, {'A', 'A', 17}: 0.133333333, {'A', 'A', 18}: 0.5,
		{'A', 'A', 19}: 0.538461538, {'A', 'A', 20}: 0.6,

		// rA:dC (guide A, off-target C)
		{'A', 'C', 1}: 1.0, {'A', 'C', 2}: 0.8, {'A', 'C', 3}: 0.611111111,
		{'A', 'C', 4}: 0.625, {'A', 'C', 5}: 0.72, {'A', 'C', 6}: 0.714285714,
		{'A', 'C', 7}: 0.705882353, {'A', 'C', 8}: 0.733333333, {'A', 'C', 9}: 0.666666667,
		{'A', 'C', 10}: 0.555555556, {'A', 'C', 11}: 0.65, {'A', 'C', 12}: 0.722222222,
		{'A', 'C', 13}: 0.652173913, {'A', 'C', 14}: 0.466666667, {'A', 'C', 15}: 0.65,
		{'A', 'C', 16}: 0.192307692, {'A', 'C', 17}: 0.176470588, {'A', 'C', 18}: 0.4,
		{'A', 'C', 19}: 0.375, {'A', 'C', 20}: 0.764705882,

		// rA:dG (guide A, off-target G)
		{'A', 'G', 1}: 0.857142857, {'A', 'G', 2}: 0.785714286, {'A', 'G', 3}: 0.428571429,
		{'A', 'G', 4}: 0.352941176, {'A', 'G', 5}: 0.5, {'A', 'G', 6}: 0.454545455,
		{'A', 'G', 7}: 0.4375, {'A', 'G', 8}: 0.428571429, {'A', 'G', 9}: 0.571428571,
		{'A', 'G', 10}: 0.333333333, {'A', 'G', 11}: 0.4, {'A', 'G', 12}: 0.263157895,
		{'A', 'G', 13}: 0.210526316, {'A', 'G', 14}: 0.214285714, {'A', 'G', 15}: 0.272727273,
		{'A', 'G', 16}: 0.0, {'A', 'G', 17}: 0.176470588, {'A', 'G', 18}: 0.19047619,
		{'A', 'G', 19}: 0.206896552, {'A', 'G', 20}: 0.227272727,

		// rC:dA (guide C, off-target A)
		{'C', 'A', 1}: 1.0, {'C', 'A', 2}: 0.909090909, {'C', 'A', 3}: 0.6875,
		{'C', 'A', 4}: 0.8, {'C', 'A', 5}: 0.636363636, {'C', 'A', 6}: 0.928571429,
		{'C', 'A', 7}: 0.8125, {'C', 'A', 8}: 0.875, {'C', 'A', 9}: 0.875,
		{'C', 'A', 10}: 0.941176471, {'C', 'A', 11}: 0.307692308, {'C', 'A', 12}: 0.538461538,
		{'C', 'A', 13}: 0.7, {'C', 'A', 14}: 0.733333333, {'C', 'A', 15}: 0.066666667,
		{'C', 'A', 16}: 0.307692308, {'C', 'A', 17}: 0.466666667, {'C', 'A', 18}: 0.642857143,
		{'C', 'A', 19}: 0.461538462, {'C', 'A', 20}: 0.3,

		// rC:dC (guide C, off-target C)
		{'C', 'C', 1}: 0.913043478, {'C', 'C', 2}: 0.695652174, {'C', 'C', 3}: 0.5,
		{'C', 'C', 4}: 0.5, {'C', 'C', 5}: 0.6, {'C', 'C', 6}: 0.5,
		{'C', 'C', 7}: 0.470588235, {'C', 'C', 8}: 0.642857143, {'C', 'C', 9}: 0.619047619,
		{'C', 'C', 10}: 0.388888889, {'C', 'C', 11}: 0.25, {'C', 'C', 12}: 0.444444444,
		{'C', 'C', 13}: 0.136363636, {'C', 'C', 14}: 0.0, {'C', 'C', 15}: 0.05,
		{'C', 'C', 16}: 0.153846154, {'C', 'C', 17}: 0.058823529, {'C', 'C', 18}: 0.133333333,
		{'C', 'C', 19}: 0.125, {'C', 'C', 20}: 0.058823529,

		// rC:dT (guide C, off-target T)
		{'C', 'T', 1}: 1.0, {'C', 'T', 2}: 0.727272727, {'C', 'T', 3}: 0.866666667,
		{'C', 'T', 4}: 0.842105263, {'C', 'T', 5}: 0.571428571, {'C', 'T', 6}: 0.928571429,
		{'C', 'T', 7}: 0.75, {'C', 'T', 8}: 0.65, {'C', 'T', 9}: 0.857142857,
		{'C', 'T', 10}: 0.866666667, {'C', 'T', 11}: 0.75, {'C', 'T', 12}: 0.714285714,
		{'C', 'T', 13}: 0.384615385, {'C', 'T', 14}: 0.35, {'C', 'T', 15}: 0.222222222,
		{'C', 'T', 16}: 1.0, {'C', 'T', 17}: 0.466666667, {'C', 'T', 18}: 0.538461538,
		{'C', 'T', 19}: 0.428571429, {'C', 'T', 20}: 0.5,

		// rG:dA (guide G, off-target A)
		{'G', 'A', 1}: 1.0, {'G', 'A', 2}: 0.636363636, {'G', 'A', 3}: 0.5,
		{'G', 'A', 4}: 0.363636364, {'G', 'A', 5}: 0.3, {'G', 'A', 6}: 0.666666667,
		{'G', 'A', 7}: 0.571428571, {'G', 'A', 8}: 0.625, {'G', 'A', 9}: 0.533333333,
		{'G', 'A', 10}: 0.8125, {'G', 'A', 11}: 0.384615385, {'G', 'A', 12}: 0.384615385,
		{'G', 'A', 13}: 0.3, {'G', 'A', 14}: 0.266666667, {'G', 'A', 15}: 0.142857143,
		{'G', 'A', 16}: 0.0, {'G', 'A', 17}: 0.25, {'G', 'A', 18}: 0.666666667,
		{'G', 'A', 19}: 0.666666667, {'G', 'A', 20}: 0.7,

		// rG:dG (guide G, off-target G)
		{'G', 'G', 1}: 0.714285714, {'G', 'G', 2}: 0.692307692, {'G', 'G', 3}: 0.384615385,
		{'G', 'G', 4}: 0.529411765, {'G', 'G', 5}: 0.785714286, {'G', 'G', 6}: 0.681818182,
		{'G', 'G', 7}: 0.6875, {'G', 'G', 8}: 0.615384615, {'G', 'G', 9}: 0.538461538,
		{'G', 'G', 10}: 0.4, {'G', 'G', 11}: 0.428571429, {'G', 'G', 12}: 0.529411765,
		{'G', 'G', 13}: 0.421052632, {'G', 'G', 14}: 0.428571429, {'G', 'G', 15}: 0.272727273,
		{'G', 'G', 16}: 0.0, {'G', 'G', 17}: 0.235294118, {'G', 'G', 18}: 0.476190476,
		{'G', 'G', 19}: 0.448275862, {'G', 'G', 20}: 0.428571429,

		// rG:dT (guide G, off-target T)
		{'G', 'T', 1}: 0.9, {'G', 'T', 2}: 0.846153846, {'G', 'T', 3}: 0.75,
		{'G', 'T', 4}: 0.9, {'G', 'T', 5}: 0.866666667, {'G', 'T', 6}: 1.0,
		{'G', 'T', 7}: 1.0, {'G', 'T', 8}: 1.0, {'G', 'T', 9}: 0.642857143,
		{'G', 'T', 10}: 0.933333333, {'G', 'T', 11}: 1.0, {'G', 'T', 12}: 0.933333333,
		{'G', 'T', 13}: 0.923076923, {'G', 'T', 14}: 0.75, {'G', 'T', 15}: 0.941176471,
		{'G', 'T', 16}: 1.0, {'G', 'T', 17}: 0.933333333, {'G', 'T', 18}: 0.692307692,
		{'G', 'T', 19}: 0.714285714, {'G', 'T', 20}: 0.9375,

		// rU:dC (guide T→RNA U, off-target C)
		{'U', 'C', 1}: 0.956521739, {'U', 'C', 2}: 0.84, {'U', 'C', 3}: 0.5,
		{'U', 'C', 4}: 0.625, {'U', 'C', 5}: 0.64, {'U', 'C', 6}: 0.571428571,
		{'U', 'C', 7}: 0.588235294, {'U', 'C', 8}: 0.733333333, {'U', 'C', 9}: 0.619047619,
		{'U', 'C', 10}: 0.5, {'U', 'C', 11}: 0.4, {'U', 'C', 12}: 0.5,
		{'U', 'C', 13}: 0.260869565, {'U', 'C', 14}: 0.0, {'U', 'C', 15}: 0.05,
		{'U', 'C', 16}: 0.346153846, {'U', 'C', 17}: 0.117647059, {'U', 'C', 18}: 0.333333333,
		{'U', 'C', 19}: 0.25, {'U', 'C', 20}: 0.176470588,

		// rU:dG (guide T→RNA U, off-target G)
		{'U', 'G', 1}: 0.857142857, {'U', 'G', 2}: 0.857142857, {'U', 'G', 3}: 0.428571429,
		{'U', 'G', 4}: 0.647058824, {'U', 'G', 5}: 1.0, {'U', 'G', 6}: 0.909090909,
		{'U', 'G', 7}: 0.6875, {'U', 'G', 8}: 1.0, {'U', 'G', 9}: 0.923076923,
		{'U', 'G', 10}: 0.533333333, {'U', 'G', 11}: 0.666666667, {'U', 'G', 12}: 0.947368421,
		{'U', 'G', 13}: 0.789473684, {'U', 'G', 14}: 0.285714286, {'U', 'G', 15}: 0.272727273,
		{'U', 'G', 16}: 0.666666667, {'U', 'G', 17}: 0.705882353, {'U', 'G', 18}: 0.428571429,
		{'U', 'G', 19}: 0.275862069, {'U', 'G', 20}: 0.090909091,

		// rU:dT (guide T→RNA U, off-target T)
		{'U', 'T', 1}: 1.0, {'U', 'T', 2}: 0.846153846, {'U', 'T', 3}: 0.714285714,
		{'U', 'T', 4}: 0.476190476, {'U', 'T', 5}: 0.5, {'U', 'T', 6}: 0.866666667,
		{'U', 'T', 7}: 0.875, {'U', 'T', 8}: 0.8, {'U', 'T', 9}: 0.928571429,
		{'U', 'T', 10}: 0.857142857, {'U', 'T', 11}: 0.75, {'U', 'T', 12}: 0.8,
		{'U', 'T', 13}: 0.692307692, {'U', 'T', 14}: 0.619047619, {'U', 'T', 15}: 0.578947368,
		{'U', 'T', 16}: 0.909090909, {'U', 'T', 17}: 0.533333333, {'U', 'T', 18}: 0.666666667,
		{'U', 'T', 19}: 0.285714286, {'U', 'T', 20}: 0.5625,
	}

	cfdPAMScores = map[string]float64{
		"GG": 1.0, "AG": 0.259259259, "CG": 0.107142857, "TG": 0.038961039,
		"GA": 0.069444444, "GC": 0.022222222, "GT": 0.016129032,
		"AA": 0.0, "AC": 0.0, "AT": 0.0,
		"CA": 0.0, "CC": 0.0, "CT": 0.0,
		"TA": 0.0, "TC": 0.0, "TT": 0.0,
	}
}

// dnaToRNA converts guide DNA base to RNA notation for CFD lookup.
// T→U, all others unchanged.
func dnaToRNA(b byte) byte {
	if b == 'T' || b == 't' {
		return 'U'
	}
	if b >= 'a' && b <= 'z' {
		return b - 32 // uppercase
	}
	return b
}

// CfdScore computes the CFD off-target score for a single guide vs off-target pair.
// guide and offTarget are 20bp DNA sequences (same strand, PAM excluded).
// pam is the 2bp PAM dinucleotide of the off-target (e.g. "GG", "AG").
// Returns score in range [0, 1]. Higher = more likely to cut at off-target.
func CfdScore(guide, offTarget string, pam string) float64 {
	if len(guide) != 20 || len(offTarget) != 20 || len(pam) != 2 {
		return 0
	}

	score := 1.0

	for i := range 20 {
		gBase := guide[i]
		oBase := offTarget[i]

		if gBase == oBase {
			continue // match, contributes 1.0
		}

		rnaBase := dnaToRNA(gBase)
		key := cfdMismatchKey{
			guideRNA:    rnaBase,
			offTargetDN: oBase,
			position:    i + 1, // 1-indexed
		}

		if penalty, ok := cfdMismatchScores[key]; ok {
			score *= penalty
		}
		// If key not found (Watson-Crick complement mismatch), score *= 1.0 (no penalty)
	}

	pamUpper := string([]byte{toUpper(pam[0]), toUpper(pam[1])})
	if pamPenalty, ok := cfdPAMScores[pamUpper]; ok {
		score *= pamPenalty
	} else {
		score = 0 // unknown PAM
	}

	return score
}

func toUpper(b byte) byte {
	if b >= 'a' && b <= 'z' {
		return b - 32
	}
	return b
}
