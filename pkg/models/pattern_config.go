package models

type SegmentConfig struct {
	Size            int `json:"size"`
	AllowedMismatch int `json:"allowedMismatch"`
}

type PatternConfig struct {
	Segments           []*SegmentConfig `json:"segments"`
	MaxMismatchAllowed int              `json:"maxMismatchAllowed"`
	TotalSize          int              `json:"totalSize"`
}
