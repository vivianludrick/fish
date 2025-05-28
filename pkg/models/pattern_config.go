package models

import (
	"fmt"
	"vivalchemy/cris/internal/bitmaps"
	"vivalchemy/cris/internal/utils"
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

func converter(config *NewPatternSearchConfig) (*NewInternalPatternSearchConfig, error) { // Validation
	if err := config.Validate(); err != nil {
		return nil, err
	}

	internalPSC := NewNewInternalPatternSearchConfig(len(config.GuideSequences))

	// input validation is already done
	targetFilePath, _ := AvailableGenomes[config.TargetGenome]
	internalPSC.TargetFilePath = targetFilePath

	internalPSC.AllowedNs = config.AllowedNs

	for _, benchMark := range config.SelectedBenchmarks {
		// validation for benchmark is already done
		val, _ := BenchmarkAlgorithms[benchMark]
		internalPSC.SelectedBenchmarks = append(internalPSC.SelectedBenchmarks, val)
	}

	internalPSC.MaxTotalMismatches = config.ToleranceSpec.MaxTotalMismatches
	internalPSC.TotalGuideLength = config.ToleranceSpec.TotalGuideLength

	// copy over the allowedMismatches
	for _, segment := range config.ToleranceSpec.SegmentSpec {
		internalSegmentTolerance := NewNewInternalSegmentTolerance()
		internalSegmentTolerance.AllowedMismatches = segment.AllowedMismatches

		total15Divisions := segment.Length / 15
		overFlowlength := segment.Length % 15
		for range total15Divisions {
			internalSegmentTolerance.Lengths = append(internalSegmentTolerance.Lengths, 15)
		}
		internalSegmentTolerance.Lengths = append(internalSegmentTolerance.Lengths, overFlowlength)

		for _, variant := range segment.AllowedVariants {
			var lengthPassedSoFar uint
			var encodedVariants []uint64
			for _, length := range internalSegmentTolerance.Lengths {
				var encodedVariantSegment uint64
				for indexInCurrentLength := range length {
					encodedVariantSegment = (encodedVariantSegment << 4) | bitmaps.NucleotideToBitMap[variant[lengthPassedSoFar+indexInCurrentLength]]
				}
				encodedVariants = append(encodedVariants, encodedVariantSegment)
				lengthPassedSoFar += length
			}
			internalSegmentTolerance.AllowedVariants = append(internalSegmentTolerance.AllowedVariants, encodedVariants)
		}

		internalPSC.SegmentSpec = append(internalPSC.SegmentSpec, internalSegmentTolerance)
	}

	for _, guideSequence := range config.GuideSequences {
		var lengthPassedSoFar uint
		var encodedGuide []uint64        // 00000ATGC
		var encodedreverseGuide []uint64 // 0000CGTA
		for _, segment := range internalPSC.SegmentSpec {
			for _, length := range segment.Lengths {
				var encodedGuideSegment uint64
				var encodedReverseGuideSegment uint64
				for indexInLength := range length {
					encodedGuideSegment = (encodedGuideSegment << 4) |
						bitmaps.NucleotideToBitMap[guideSequence[lengthPassedSoFar+indexInLength]]
					encodedReverseGuideSegment = (encodedReverseGuideSegment >> 4) |
						(bitmaps.NucleotideToBitMap[bitmaps.NucleotideComplementMap[guideSequence[lengthPassedSoFar+indexInLength]]] << (4 * 15))
				}
				// shift the reverse complement by the remaining spaces
				encodedReverseGuideSegment = encodedReverseGuideSegment >> (4 * (16 - length))
				encodedGuide = append(encodedGuide, encodedGuideSegment)
				encodedreverseGuide = utils.Prepend(encodedreverseGuide, encodedReverseGuideSegment)
				lengthPassedSoFar += length
			}
		}
		internalPSC.GuideSequences = append(internalPSC.GuideSequences, encodedGuide)
		internalPSC.ReverseGuideSequences = append(internalPSC.ReverseGuideSequences, encodedreverseGuide)
	}

	return internalPSC, nil
}
