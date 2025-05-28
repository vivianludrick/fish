package encoders

import (
	"vivalchemy/cris/internal/bitmaps"
	"vivalchemy/cris/internal/utils"
	"vivalchemy/cris/pkg/models"
)

func ValidateAndEncodeConfig(config *models.NewPatternSearchConfig) (*models.NewInternalPatternSearchConfig, error) { // Validation
	if err := config.Validate(); err != nil {
		return nil, err
	}

	internalPSC := models.NewNewInternalPatternSearchConfig(len(config.GuideSequences))

	// input validation is already done
	targetFilePath, _ := models.AvailableGenomes[config.TargetGenome]
	internalPSC.TargetFilePath = targetFilePath

	internalPSC.AllowedNs = config.AllowedNs

	for _, benchMark := range config.SelectedBenchmarks {
		// validation for benchmark is already done
		val, _ := models.ScoringAlgorithms[benchMark]
		internalPSC.SelectedBenchmarks = append(internalPSC.SelectedBenchmarks, val)
	}

	internalPSC.MaxTotalMismatches = config.ToleranceSpec.MaxTotalMismatches
	internalPSC.TotalGuideLength = config.ToleranceSpec.TotalGuideLength

	// copy over the allowedMismatches
	for _, segment := range config.ToleranceSpec.SegmentSpec {
		internalSegmentTolerance := models.NewNewInternalSegmentTolerance()
		internalSegmentTolerance.AllowedMismatches = segment.AllowedMismatches

		total15Divisions := segment.Length / 15
		overFlowlength := segment.Length % 15
		for range total15Divisions {
			internalSegmentTolerance.Lengths = append(internalSegmentTolerance.Lengths, 15)
		}
		internalSegmentTolerance.Lengths = append(internalSegmentTolerance.Lengths, overFlowlength)

		for _, variant := range segment.AllowedVariants {
			var lengthPassedSoFar int
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
		var lengthPassedSoFar int
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
