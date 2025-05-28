package models

import "fmt"

type NewInternalPatternSearchConfig struct {
	GuideSequences           [][]uint64 // Guide sequences to search in the genome
	ReverseGuideSequences    [][]uint64 // Guide sequences to search in the genome
	TargetFilePath           string     // the fasta file path of the genome they are targetting
	NewInternalToleranceSpec            // fill from the preset variant
	SelectedBenchmarks       []BenchmarkAlgorithm
	AllowedNs                uint
}

type NewInternalToleranceSpec struct {
	SegmentSpec        []*NewInternalSegmentTolerance // Individual sub-regions of the pattern (e.g., core, suffix, PAM)
	MaxTotalMismatches uint                           // Maximum mismatches allowed across all pattern regions
	TotalGuideLength   uint                           // Total length of the full pattern to match
}

type NewInternalSegmentTolerance struct {
	// either length or lengths
	Lengths []uint

	// either mismatches or variants
	AllowedMismatches uint
	AllowedVariants   [][]uint64
	// TODO: Add the reverse complement of the variants
}

func NewNewInternalPatternSearchConfig(guideSequencesLength int) *NewInternalPatternSearchConfig {
	return &NewInternalPatternSearchConfig{
		GuideSequences:        make([][]uint64, 0, guideSequencesLength),
		ReverseGuideSequences: make([][]uint64, 0, guideSequencesLength),
		NewInternalToleranceSpec: NewInternalToleranceSpec{
			SegmentSpec: make([]*NewInternalSegmentTolerance, 0),
		},
		SelectedBenchmarks: make([]BenchmarkAlgorithm, 0),
	}
}

func NewNewInternalSegmentTolerance() *NewInternalSegmentTolerance {
	return &NewInternalSegmentTolerance{
		Lengths:         make([]uint, 0, 1),  // there will be atleast on length
		AllowedVariants: make([][]uint64, 0), // there is a possibility that there are no variants
	}
}

func (config *NewInternalPatternSearchConfig) Print() {
	fmt.Println("TargetFilePath:", config.TargetFilePath)
	fmt.Println("AllowedNs:", config.AllowedNs)
	fmt.Println("SelectedBenchmarks:", config.SelectedBenchmarks)
	fmt.Println("MaxTotalMismatches:", config.MaxTotalMismatches)
	fmt.Println("TotalGuideLength:", config.TotalGuideLength)
	fmt.Println("SegmentSpec:")
	for _, segment := range config.SegmentSpec {
		fmt.Println("--- New Segment ---")
		fmt.Println("Lengths:", segment.Lengths)
		fmt.Println("AllowedMismatches:", segment.AllowedMismatches)
		fmt.Println("AllowedVariants:", segment.AllowedVariants)
	}
	fmt.Println("GuideSequences:")
	for _, guideSequence := range config.GuideSequences {
		fmt.Println("--- New Guide Sequence ---")
		for _, nucleotide := range guideSequence {
			fmt.Printf("%#x", nucleotide)
		}
		fmt.Println()
	}
	fmt.Println("ReverseGuideSequences:")
	for _, reverseGuideSequence := range config.ReverseGuideSequences {
		fmt.Println("--- New Reverse Guide Sequence ---")
		for _, nucleotide := range reverseGuideSequence {
			fmt.Printf("%#x", nucleotide)
		}
	}
}
