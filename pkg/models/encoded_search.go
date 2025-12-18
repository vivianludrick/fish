package models

import "fmt"

type EncodedSearch struct {
	EncodedGuides    [][]uint64 // Guide sequences to search in the genome
	EncodedRCGuides  [][]uint64 // Guide sequences to search in the genome
	TargetFilePath   string     // the fasta file path of the genome they are targetting
	EncodedTolerance            // fill from the preset variant
	ScoringModels    []ScoringModels
}

type EncodedTolerance struct {
	EncodedSegments []*EncodedSegment // Individual sub-regions of the pattern (e.g., core, suffix, PAM)
	MaxMismatches   int               // Maximum mismatches allowed across all pattern regions
	GuideLength     int               // Total length of the full pattern to match
	AllowedNs       int
}

type EncodedSegment struct {
	// either length or lengths
	Lengths []int

	// either mismatches or variants
	MaxMismatches   int
	AllowedVariants [][]uint64

	AllowedRCVariants [][]uint64
}

func NewEncodedConfig(guideSequencesLength int) *EncodedSearch {
	return &EncodedSearch{
		EncodedGuides:   make([][]uint64, 0, guideSequencesLength),
		EncodedRCGuides: make([][]uint64, 0, guideSequencesLength),
		EncodedTolerance: EncodedTolerance{
			EncodedSegments: make([]*EncodedSegment, 0),
		},
		ScoringModels: make([]ScoringModels, 0),
	}
}

func NewEncodedSegment() *EncodedSegment {
	return &EncodedSegment{
		Lengths:           make([]int, 0, 1),   // there will be atleast one length
		AllowedVariants:   make([][]uint64, 0), // there is a possibility that there are no variants
		AllowedRCVariants: make([][]uint64, 0), // there is a possibility that there are no variants
	}
}

func (config *EncodedSearch) GetNumberOfSegments() int {
	totalSegments := 0
	for _, segment := range config.EncodedSegments {
		totalSegments += len(segment.Lengths)
	}
	return totalSegments
}

func (config *EncodedSearch) Println() {
	fmt.Println("TargetFilePath:", config.TargetFilePath)
	fmt.Println("AllowedNs:", config.AllowedNs)
	fmt.Println("SelectedBenchmarks:", config.ScoringModels)
	fmt.Println("MaxTotalMismatches:", config.MaxMismatches)
	fmt.Println("TotalGuideLength:", config.GuideLength)
	fmt.Println("SegmentSpec:")
	for _, segment := range config.EncodedSegments {
		fmt.Println("--- New Segment ---")
		fmt.Println("Lengths:", segment.Lengths)
		fmt.Println("AllowedMismatches:", segment.MaxMismatches)
		fmt.Println("AllowedVariants:", segment.AllowedVariants)
		fmt.Println("AllowedReverseComplementVariants:", segment.AllowedRCVariants)
	}
	fmt.Println("GuideSequences:")
	for _, guideSequence := range config.EncodedGuides {
		fmt.Println("--- New Guide Sequence ---")
		for _, nucleotide := range guideSequence {
			fmt.Printf("%#x", nucleotide)
		}
		fmt.Println()
	}
	fmt.Println("ReverseGuideSequences:")
	for _, reverseGuideSequence := range config.EncodedRCGuides {
		fmt.Println("--- New Reverse Guide Sequence ---")
		for _, nucleotide := range reverseGuideSequence {
			fmt.Printf("%#x", nucleotide)
		}
		fmt.Println()
	}
}
