package encoders

import (
	"errors"
	"strconv"
	"vivalchemy/cris/internal/bitmaps"
	"vivalchemy/cris/pkg/models"
)

func SegmentAndEncodePattern(pattern string, config *models.PatternConfig) ([]uint64, error) {
	if len(pattern) != config.TotalSize {
		return nil, errors.New("Target PAM sequence must be exactly " + strconv.Itoa(config.TotalSize) + " nucleotides; got " + strconv.Itoa(len(pattern)))
	}

	segments := make([]uint64, len(config.Segments))
	position := 0
	sizeSoFar := 0

	for segmentIdx, segment := range config.Segments {
		// since a uint64 can support 16segments but we need 1 for carry over after shifting
		if segment.Size < 1 || segment.Size > 15 {
			return nil, errors.New("Invalid segment size, must be between 1 and 15")
		}

		sizeSoFar += segment.Size
		if sizeSoFar > len(pattern) {
			return nil, errors.New("Target PAM sequence is too short")
		}

		for range segment.Size {
			nucleotide := pattern[position]
			mapped := bitmaps.NucleotideToBitMap[nucleotide]
			if mapped == 0 {
				return nil, errors.New("Invalid nucleotide in target PAM sequence at position " + strconv.Itoa(position))
			}
			segments[segmentIdx] = (segments[segmentIdx] << 4) | mapped
			position++
		}
	}

	return segments, nil
}
