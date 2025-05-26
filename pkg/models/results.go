package models

//	{
//		<guide_sequence>:{
//			<matched_sequence>:{
//				locations:[],
//				mismatches_locations: []unique
//			}
//		}
//	}
type Result map[string]MatchedSequences

type MatchedSequences map[string]SequenceDetails

type SequenceDetails struct {
	Locations         []*MatchLocation `json:"locations"`
	MismatchPositions []uint           `json:"mismatch_positions"`
}

type MatchLocation struct {
	Header string `json:"header"`
	Offset int    `json:"offset"`
}
