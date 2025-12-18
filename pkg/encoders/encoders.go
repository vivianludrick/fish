package encoders

import (
	"vivalchemy/cris/internal/bitmaps"
	"vivalchemy/cris/internal/utils"
	"vivalchemy/cris/pkg/models"
)

func ValidateAndEncodeConfig(config *models.SearchRequest) (*models.EncodedSearch, error) {
	// Validation
	if err := config.Validate(); err != nil {
		return nil, err
	}

	internalPSC := models.NewEncodedConfig(len(config.GuideSequences))

	// Encode basic configuration
	encodeBasicConfig(config, internalPSC)

	// Encode tolerance specifications and segments
	if err := encodeToleranceSpec(*config.Tolerance, internalPSC); err != nil {
		return nil, err
	}

	// Encode guide sequences
	encodeGuideSequences(config.GuideSequences, internalPSC)

	return internalPSC, nil
}

func encodeBasicConfig(config *models.SearchRequest, internalPSC *models.EncodedSearch) {
	// Set target file path
	targetFilePath, _ := models.AvailableGenomes[config.GenomeID]
	internalPSC.TargetFilePath = targetFilePath
	internalPSC.AllowedNs = config.Tolerance.AllowedNs

	// Encode selected benchmarks
	for _, benchMark := range config.ScoringModels {
		val, _ := models.ScoringAlgorithms[benchMark]
		internalPSC.ScoringModels = append(internalPSC.ScoringModels, val)
	}

	// Set tolerance values
	internalPSC.MaxMismatches = config.Tolerance.MaxMismatches
	internalPSC.GuideLength = config.Tolerance.GuideLength
}

func encodeToleranceSpec(toleranceSpec models.Tolerance, internalPSC *models.EncodedSearch) error {
	for _, segment := range toleranceSpec.SegmentTolerance {
		internalSegment, err := encodeSegment(segment)
		if err != nil {
			return err
		}
		internalPSC.EncodedSegments = append(internalPSC.EncodedSegments, internalSegment)
	}
	return nil
}

func encodeSegment(segment models.SegmentTolerance) (*models.EncodedSegment, error) {
	internalSegmentTolerance := models.NewEncodedSegment()
	internalSegmentTolerance.MaxMismatches = segment.AllowedMismatches

	// Calculate segment lengths (15-base chunks + overflow)
	internalSegmentTolerance.Lengths = calculateSegmentLengths(segment.Length)

	// Will be ignore if there are no variants as the looping will be skipped
	// Encode all allowed variants for this segment
	for _, variant := range segment.AllowedVariants {
		encodedVariant := encodeVariantSequence(variant, internalSegmentTolerance.Lengths)
		internalSegmentTolerance.AllowedVariants = append(internalSegmentTolerance.AllowedVariants, encodedVariant)

		// reverse complement variants
		encodedReverseComplementVariant := encodeVariantSequence(utils.ReverseComplement(variant), utils.Reverse(internalSegmentTolerance.Lengths))
		internalSegmentTolerance.AllowedRCVariants = append(internalSegmentTolerance.AllowedRCVariants, encodedReverseComplementVariant)
	}

	return internalSegmentTolerance, nil
}

func calculateSegmentLengths(totalLength int) []int {
	var lengths []int
	total15Divisions := totalLength / 15
	overFlowLength := totalLength % 15

	// Add 15-base segments
	for range total15Divisions {
		lengths = append(lengths, 15)
	}

	// Add overflow segment if exists
	if overFlowLength > 0 {
		lengths = append(lengths, overFlowLength)
	}

	return lengths
}

func encodeVariantSequence(variant string, lengths []int) []uint64 {
	var encodedVariants []uint64
	var lengthPassedSoFar int

	for _, length := range lengths {
		var encodedVariantSegment uint64
		for indexInCurrentLength := range length {
			encodedVariantSegment = (encodedVariantSegment << 4) |
				bitmaps.NucleotideToBitMap[variant[lengthPassedSoFar+indexInCurrentLength]]
		}
		encodedVariants = append(encodedVariants, encodedVariantSegment)
		lengthPassedSoFar += length
	}

	return encodedVariants
}

func encodeGuideSequences(guideSequences []string, internalPSC *models.EncodedSearch) {
	for _, guideSequence := range guideSequences {
		encodedGuide, encodedReverseGuide := encodeGuideSequence(guideSequence, internalPSC.EncodedSegments)
		internalPSC.EncodedGuides = append(internalPSC.EncodedGuides, encodedGuide)
		internalPSC.EncodedRCGuides = append(internalPSC.EncodedRCGuides, encodedReverseGuide)
	}
}

func encodeGuideSequence(guideSequence string, segmentSpec []*models.EncodedSegment) ([]uint64, []uint64) {
	var encodedGuide []uint64
	var encodedReverseGuide []uint64
	var lengthPassedSoFar int

	for _, segment := range segmentSpec {
		for _, length := range segment.Lengths {
			encodedSegment, encodedReverseSegment := encodeGuideSegment(
				guideSequence, lengthPassedSoFar, length)

			encodedGuide = append(encodedGuide, encodedSegment)
			encodedReverseGuide = utils.Prepend(encodedReverseGuide, encodedReverseSegment)
			lengthPassedSoFar += length
		}
	}

	return encodedGuide, encodedReverseGuide
}

func encodeGuideSegment(guideSequence string, startPos, length int) (uint64, uint64) {
	var encodedGuideSegment uint64
	var encodedReverseGuideSegment uint64

	for indexInLength := range length {
		nucleotide := guideSequence[startPos+indexInLength]

		// Forward encoding
		encodedGuideSegment = (encodedGuideSegment << 4) | bitmaps.NucleotideToBitMap[nucleotide]

		// Reverse complement encoding
		complement := bitmaps.NucleotideComplementMap[nucleotide]
		encodedReverseGuideSegment = (encodedReverseGuideSegment >> 4) |
			(bitmaps.NucleotideToBitMap[complement] << (4 * 15))
	}

	// Shift the reverse complement by the remaining spaces
	encodedReverseGuideSegment = encodedReverseGuideSegment >> (4 * (16 - length))

	return encodedGuideSegment, encodedReverseGuideSegment
}
