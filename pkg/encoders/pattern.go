package encoders

import (
	"errors"
	"strconv"
	"vivalchemy/cris/pkg/models"
)

var bitMapArray [256]uint64
var reverseBitMapArray [16]byte

func init() {
	// Initialize lookup arrays
	bitMapArray['A'] = 0x8 // 1000
	bitMapArray['C'] = 0x4 // 0100
	bitMapArray['G'] = 0x2 // 0010
	bitMapArray['T'] = 0x1 // 0001
	bitMapArray['N'] = 0xF // 1111 (ambiguous nucleotide)
	bitMapArray['a'] = 0x8 // Support lowercase
	bitMapArray['c'] = 0x4
	bitMapArray['g'] = 0x2
	bitMapArray['t'] = 0x1
	bitMapArray['n'] = 0xF

	reverseBitMapArray[0x8] = 'A'
	reverseBitMapArray[0x4] = 'C'
	reverseBitMapArray[0x2] = 'G'
	reverseBitMapArray[0x1] = 'T'
	reverseBitMapArray[0xF] = 'N'
}

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
			mapped := bitMapArray[nucleotide]
			if mapped == 0 {
				return nil, errors.New("Invalid nucleotide in target PAM sequence at position " + strconv.Itoa(position))
			}
			segments[segmentIdx] = (segments[segmentIdx] << 4) | mapped
			position++
		}
	}

	return segments, nil
}
