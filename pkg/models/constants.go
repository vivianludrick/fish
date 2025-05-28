package models

// -----------------------------------------------------------------------------
// -----------------------------------------------------------------------------
// ENUMS
// -----------------------------------------------------------------------------
// -----------------------------------------------------------------------------

type Genome string

const (
	DanioRerio     Genome = "danio-rerio"
	AtlanticSalmon Genome = "atlantic-salmon"
)

// relate the genomes to their fastafile paths
var AvailableGenomes = map[Genome]string{
	DanioRerio:     "./genome/danio-rerio",
	AtlanticSalmon: "./genomes/atlantic-salmon",
}

type PresetVariant string

const (
	SpCas9 PresetVariant = "sp-cas9"
	Custom PresetVariant = "custom"
)

var PresetVariants = map[PresetVariant]NewToleranceSpec{
	SpCas9: {
		SegmentSpec: []NewSegmentTolerance{
			{
				Length:            10,
				AllowedMismatches: 4,
			},
			{
				Length:            10,
				AllowedMismatches: 2,
			},
			{
				Length:            3,
				AllowedMismatches: 0,                             // exact match required for this segment
				AllowedVariants:   []string{"NGG", "NAG", "NGA"}, // allow degenerate bases here
			},
		},
		MaxTotalMismatches: 4,
		TotalGuideLength:   23,
	},
	Custom: {}, // empty since the tolerance spec will be provided
}

type BenchmarkAlgorithm int

const (
	Mitscore BenchmarkAlgorithm = iota
	Doench
)

var BenchmarkAlgorithms = map[string]BenchmarkAlgorithm{
	"mitscore": Mitscore,
	"doench":   Doench,
}
