package models

// -----------------------------------------------------------------------------
// -----------------------------------------------------------------------------
// ENUMS
// -----------------------------------------------------------------------------
// -----------------------------------------------------------------------------

type GenomeID string

const (
	GenomeDanioRerio     GenomeID = "danio-rerio"
	GenomeHilsa          GenomeID = "hilsa"
	GenomeAtlanticSalmon GenomeID = "atlantic-salmon"
	GenomeTest           GenomeID = "test" // TODO: remove this
)

// relate the genomes to their fastafile paths
var AvailableGenomes = map[GenomeID]string{
	GenomeDanioRerio:     "./genomes/GCA_000002035.4_GRCz11_genomic.fna",
	GenomeHilsa:          "./genomes/GCA_015244755.2_TenIli1.0_genomic.fna",
	GenomeAtlanticSalmon: "./genomes/GCF_905237065.1_Ssal_v3.1_genomic.fna",
	GenomeTest:           "./input.fna", // TODO: remove this
}

type MatchVariant string

const (
	SearchVariantSpCas9 MatchVariant = "sp-cas9"
	SearchVariantCustom MatchVariant = "custom"
)

var PresetVariants = map[MatchVariant]Tolerance{
	SearchVariantSpCas9: {
		SegmentTolerance: []SegmentTolerance{
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
		MaxMismatches: 4,
		GuideLength:   23,
	},
	SearchVariantCustom: {}, // empty since the tolerance spec will be provided
}

type ScoringModels int

const (
	ScoringMIT ScoringModels = iota
	ScoringDoench
)

var ScoringAlgorithms = map[string]ScoringModels{
	"mitscore": ScoringMIT,
	"doench":   ScoringDoench,
}
