package models

import "sort"

// ExonRegion represents a single exon interval (1-based inclusive coordinates matching GTF)
type ExonRegion struct {
	Start int // 1-based inclusive
	End   int // 1-based inclusive
}

// ExonIndex maps chromosome/seqname to sorted, merged exon intervals.
// After MergeOverlapping, intervals are non-overlapping and sorted by Start.
type ExonIndex struct {
	Regions map[string][]ExonRegion // seqname -> sorted exon regions
}

func NewExonIndex() *ExonIndex {
	return &ExonIndex{
		Regions: make(map[string][]ExonRegion),
	}
}

func (idx *ExonIndex) AddExon(seqname string, start, end int) {
	idx.Regions[seqname] = append(idx.Regions[seqname], ExonRegion{Start: start, End: end})
}

// MergeOverlapping sorts and merges overlapping/adjacent exon intervals per chromosome.
func (idx *ExonIndex) MergeOverlapping() {
	for seqname, regions := range idx.Regions {
		if len(regions) == 0 {
			continue
		}

		sort.Slice(regions, func(i, j int) bool {
			return regions[i].Start < regions[j].Start
		})

		merged := []ExonRegion{regions[0]}
		for _, r := range regions[1:] {
			last := &merged[len(merged)-1]
			if r.Start <= last.End+1 {
				if r.End > last.End {
					last.End = r.End
				}
			} else {
				merged = append(merged, r)
			}
		}
		idx.Regions[seqname] = merged
	}
}

// OverlapsRange checks if any exon overlaps with [start, end] (1-based inclusive).
// Uses binary search — O(log n) per query.
func (idx *ExonIndex) OverlapsRange(seqname string, start, end int) bool {
	regions, ok := idx.Regions[seqname]
	if !ok {
		return false
	}

	// Find first region whose End >= start
	i := sort.Search(len(regions), func(i int) bool {
		return regions[i].End >= start
	})

	if i >= len(regions) {
		return false
	}
	return regions[i].Start <= end
}
