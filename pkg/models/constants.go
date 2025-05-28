package models

// -----------------------------------------------------------------------------
// -----------------------------------------------------------------------------
// ENUMS
// -----------------------------------------------------------------------------
// -----------------------------------------------------------------------------

type Genome string

const (
	GenomeDanioRerio     Genome = "danio-rerio"
	GenomeHilsa          Genome = "hilsa"
	GenomeAtlanticSalmon Genome = "atlantic-salmon"
	GenomeTest           Genome = "test" // TODO: remove this
)

// relate the genomes to their fastafile paths
var AvailableGenomes = map[Genome]string{
	GenomeDanioRerio:     "./genomes/GCA_000002035.4_GRCz11_genomic.fna",
	GenomeHilsa:          "./genomes/GCA_015244755.2_TenIli1.0_genomic.fna",
	GenomeAtlanticSalmon: "./genomes/GCF_905237065.1_Ssal_v3.1_genomic.fna",
	GenomeTest:           "./input.fna", // TODO: remove this
}

type SearchVariant string

const (
	SearchVariantSpCas9 SearchVariant = "sp-cas9"
	SearchVariantCustom SearchVariant = "custom"
)

var PresetVariants = map[SearchVariant]NewToleranceSpec{
	SearchVariantSpCas9: {
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
	SearchVariantCustom: {}, // empty since the tolerance spec will be provided
}

type ScoringAlgorithm int

const (
	ScoringMIT ScoringAlgorithm = iota
	ScoringDoench
)

var ScoringAlgorithms = map[string]ScoringAlgorithm{
	"mitscore": ScoringMIT,
	"doench":   ScoringDoench,
}
