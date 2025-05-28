package models

import (
	"fmt"
	"vivalchemy/cris/internal/bitmaps"
)

type NewPatternSearchConfig struct {
	GuideSequences     []string          // Guide sequences to search in the genome
	TargetGenome       Genome            // the genome they are targetting
	ToleranceSpec      *NewToleranceSpec // Pattern matching criteria and constraints
	SearchVariant      SearchVariant     // custom/sp-cas9, or other variants
	SelectedBenchmarks []string          //mitscore, doench, etc
	AllowedNs          int               // 0 if no Ns allowed
}

type NewToleranceSpec struct {
	SegmentSpec        []NewSegmentTolerance // Individual sub-regions of the pattern (e.g., core, suffix, PAM)
	MaxTotalMismatches int                   // Maximum mismatches allowed across all pattern regions
	TotalGuideLength   int                   // Total length of the full pattern to match
}

type NewSegmentTolerance struct {
	Length int
	// either mismatches or variants
	AllowedMismatches int
	AllowedVariants   []string
}

func (config *NewPatternSearchConfig) Validate() error {
	validators := []func() error{
		config.validateGuideSequences,
		config.validateTargetGenome,
		config.validateSearchVariant,
		config.validateAllowedNs,
		config.validateNucleotides,
		config.validateBenchmarkAlgorithms,
		config.validateSegmentLengths,
		config.validateVariantLengths,
		config.validateGuideSequenceLengths,
	}

	for _, validator := range validators {
		if err := validator(); err != nil {
			return err
		}
	}
	return nil
}

func (config *NewPatternSearchConfig) validateGuideSequences() error {
	if len(config.GuideSequences) == 0 {
		return fmt.Errorf("invalid config: no guide sequences provided")
	}
	return nil
}

func (config *NewPatternSearchConfig) validateTargetGenome() error {
	if _, ok := AvailableGenomes[config.TargetGenome]; !ok {
		return fmt.Errorf("invalid config: the genome is not available in our database")
	}
	return nil
}

func (config *NewPatternSearchConfig) validateSearchVariant() error {
	if config.SearchVariant != SearchVariantCustom {
		val, ok := PresetVariants[config.SearchVariant]
		if !ok {
			return fmt.Errorf("invalid config: invalid preset variant")
		}
		config.ToleranceSpec = &val
	} else {
		if config.ToleranceSpec == nil {
			return fmt.Errorf("invalid config: custom preset requires a tolerance spec")
		}
	}
	return nil
}

func (config *NewPatternSearchConfig) validateAllowedNs() error {
	if config.AllowedNs > config.ToleranceSpec.TotalGuideLength {
		return fmt.Errorf("invalid config: allowed Ns (%d) exceeds total guide length (%d)",
			config.AllowedNs, config.ToleranceSpec.TotalGuideLength)
	}
	return nil
}

func (config *NewPatternSearchConfig) validateNucleotides() error {
	for _, seq := range config.GuideSequences {
		for _, nucleotide := range seq {
			if bitmaps.NucleotideToBitMap[nucleotide] == 0 {
				return fmt.Errorf("invalid nucleotide %c in sequence %s", nucleotide, seq)
			}
		}
	}
	return nil
}

func (config *NewPatternSearchConfig) validateBenchmarkAlgorithms() error {
	for _, benchmarkAlgorithm := range config.SelectedBenchmarks {
		if _, ok := ScoringAlgorithms[benchmarkAlgorithm]; !ok {
			return fmt.Errorf("invalid config: invalid benchmark algorithm: %v", benchmarkAlgorithm)
		}
	}
	return nil
}

func (config *NewPatternSearchConfig) validateSegmentLengths() error {
	var totalSegmentsLength int
	for _, segment := range config.ToleranceSpec.SegmentSpec {
		totalSegmentsLength += segment.Length
	}
	if totalSegmentsLength != config.ToleranceSpec.TotalGuideLength {
		return fmt.Errorf("invalid config: segment lengths (%d) don't match total guide length (%d)",
			totalSegmentsLength, config.ToleranceSpec.TotalGuideLength)
	}
	return nil
}

func (config *NewPatternSearchConfig) validateVariantLengths() error {
	for _, segment := range config.ToleranceSpec.SegmentSpec {
		for _, variant := range segment.AllowedVariants {
			if len(variant) != segment.Length {
				return fmt.Errorf("invalid config: The variant %v doesn't match the length specified for that segment", variant)
			}
		}
	}
	return nil
}

func (config *NewPatternSearchConfig) validateGuideSequenceLengths() error {
	for _, guideSequence := range config.GuideSequences {
		if len(guideSequence) != config.ToleranceSpec.TotalGuideLength {
			return fmt.Errorf("invalid config: The guide sequence %v doesn't match the total guide length specified", guideSequence)
		}
	}
	return nil
}
