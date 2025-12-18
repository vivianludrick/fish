package models

import (
	"fmt"
	"vivalchemy/cris/internal/bitmaps"
)

type SearchRequest struct {
	GuideSequences []string     // Guide sequences to search in the genome
	GenomeID       GenomeID     // the genome they are targetting
	Tolerance      *Tolerance   // Pattern matching criteria and constraints
	MatchVariant   MatchVariant // custom/sp-cas9, or other variants
	ScoringModels  []string     //mitscore, doench, etc
}

type Tolerance struct {
	SegmentTolerance []SegmentTolerance // Individual sub-regions of the pattern (e.g., core, suffix, PAM)
	MaxMismatches    int                // Maximum mismatches allowed across all pattern regions
	GuideLength      int                // Total length of the full pattern to match
	AllowedNs        int
}

type SegmentTolerance struct {
	Length int
	// either mismatches or variants
	AllowedMismatches int
	AllowedVariants   []string
}

func (config *SearchRequest) Validate() error {
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

func (config *SearchRequest) validateGuideSequences() error {
	if len(config.GuideSequences) == 0 {
		return fmt.Errorf("invalid config: no guide sequences provided")
	}
	return nil
}

func (config *SearchRequest) validateTargetGenome() error {
	if _, ok := AvailableGenomes[config.GenomeID]; !ok {
		return fmt.Errorf("invalid config: the genome is not available in our database")
	}
	return nil
}

func (config *SearchRequest) validateSearchVariant() error {
	if config.MatchVariant != SearchVariantCustom {
		val, ok := PresetVariants[config.MatchVariant]
		if !ok {
			return fmt.Errorf("invalid config: invalid preset variant")
		}
		config.Tolerance = &val
	} else {
		if config.Tolerance == nil {
			return fmt.Errorf("invalid config: custom preset requires a tolerance spec")
		}
	}
	return nil
}

func (config *SearchRequest) validateAllowedNs() error {
	if config.Tolerance.AllowedNs > config.Tolerance.GuideLength {
		return fmt.Errorf("invalid config: allowed Ns (%d) exceeds total guide length (%d)",
			config.Tolerance.AllowedNs, config.Tolerance.GuideLength)
	}
	return nil
}

func (config *SearchRequest) validateNucleotides() error {
	for _, seq := range config.GuideSequences {
		for _, nucleotide := range seq {
			if bitmaps.NucleotideToBitMap[nucleotide] == 0 {
				return fmt.Errorf("invalid nucleotide %c in sequence %s", nucleotide, seq)
			}
		}
	}
	return nil
}

func (config *SearchRequest) validateBenchmarkAlgorithms() error {
	for _, benchmarkAlgorithm := range config.ScoringModels {
		if _, ok := ScoringAlgorithms[benchmarkAlgorithm]; !ok {
			return fmt.Errorf("invalid config: invalid benchmark algorithm: %v", benchmarkAlgorithm)
		}
	}
	return nil
}

func (config *SearchRequest) validateSegmentLengths() error {
	var totalSegmentsLength int
	for _, segment := range config.Tolerance.SegmentTolerance {
		totalSegmentsLength += segment.Length
	}
	if totalSegmentsLength != config.Tolerance.GuideLength {
		return fmt.Errorf("invalid config: segment lengths (%d) don't match total guide length (%d)",
			totalSegmentsLength, config.Tolerance.GuideLength)
	}
	return nil
}

func (config *SearchRequest) validateVariantLengths() error {
	for _, segment := range config.Tolerance.SegmentTolerance {
		for _, variant := range segment.AllowedVariants {
			if len(variant) != segment.Length {
				return fmt.Errorf("invalid config: The variant %v doesn't match the length specified for that segment", variant)
			}
		}
	}
	return nil
}

func (config *SearchRequest) validateGuideSequenceLengths() error {
	for _, guideSequence := range config.GuideSequences {
		if len(guideSequence) != config.Tolerance.GuideLength {
			return fmt.Errorf("invalid config: The guide sequence %v doesn't match the total guide length specified", guideSequence)
		}
	}
	return nil
}
