package models

import (
	"fmt"
	"vivalchemy/cris/internal/bitmaps"
)

type NewPatternSearchConfig struct {
	GuideSequences     []string          // Guide sequences to search in the genome
	TargetGenome       Genome            // the genome they are targetting
	ToleranceSpec      *NewToleranceSpec // Pattern matching criteria and constraints
	PresetVariant      PresetVariant     // custom/sp-cas9, or other variants
	SelectedBenchmarks []string          //mitscore, doench, etc
	AllowedNs          uint              // 0 if no Ns allowed
}

type NewToleranceSpec struct {
	SegmentSpec        []NewSegmentTolerance // Individual sub-regions of the pattern (e.g., core, suffix, PAM)
	MaxTotalMismatches uint                  // Maximum mismatches allowed across all pattern regions
	TotalGuideLength   uint                  // Total length of the full pattern to match
}

type NewSegmentTolerance struct {
	Length uint
	// either mismatches or variants
	AllowedMismatches uint
	AllowedVariants   []string
}

func (config *NewPatternSearchConfig) Validate() error {
	if len(config.GuideSequences) == 0 {
		return fmt.Errorf("invalid config: no guide sequences provided")
	}
	if _, ok := AvailableGenomes[config.TargetGenome]; !ok {
		return fmt.Errorf("invalid config: the genome is not avaible in our database")
	}
	if config.PresetVariant != Custom {
		val, ok := PresetVariants[config.PresetVariant]
		if !ok {
			return fmt.Errorf("invalid config: invalid preset variant")
		}
		config.ToleranceSpec = &val
	} else {
		if config.ToleranceSpec == nil {
			return fmt.Errorf("invalid config: custom preset requires a tolerance spec")
		}
	}
	if config.AllowedNs > config.ToleranceSpec.TotalGuideLength {
		return fmt.Errorf("invalid config: allowed Ns (%d) exceeds total guide length (%d)",
			config.AllowedNs, config.ToleranceSpec.TotalGuideLength)
	}

	for _, seq := range config.GuideSequences {
		for _, nucleotide := range seq {
			if bitmaps.NucleotideToBitMap[nucleotide] == 0 {
				return fmt.Errorf("invalid nucleotide %c in sequence %s", nucleotide, seq)
			}
		}
	}

	for _, benchmarkAlgorithm := range config.SelectedBenchmarks {
		if _, ok := BenchmarkAlgorithms[benchmarkAlgorithm]; !ok {
			return fmt.Errorf("invalid config: invalid benchmark algorithm: %v", benchmarkAlgorithm)
		}
	}

	// Validate that segment lengths match total guide length
	var totalSegmentsLength uint
	for _, segment := range config.ToleranceSpec.SegmentSpec {
		totalSegmentsLength += segment.Length
	}
	if totalSegmentsLength != config.ToleranceSpec.TotalGuideLength {
		return fmt.Errorf("invalid config: segment lengths (%d) don't match total guide length (%d)",
			totalSegmentsLength, config.ToleranceSpec.TotalGuideLength)
	}

	// verify that each variant length is equal to the length of that segment
	for _, segment := range config.ToleranceSpec.SegmentSpec {
		for _, variant := range segment.AllowedVariants {
			if uint(len(variant)) != segment.Length {
				return fmt.Errorf("invalid config: The variant %v doesn't match the length specified for that segment", variant)
			}
		}
	}

	// verify that the guide sequence and the totalLength match
	for _, guideSequence := range config.GuideSequences {
		if uint(len(guideSequence)) != config.ToleranceSpec.TotalGuideLength {
			return fmt.Errorf("invalid config: The guide sequence %v doesn't match the total guide length specified", guideSequence)
		}
	}

	return nil
}
