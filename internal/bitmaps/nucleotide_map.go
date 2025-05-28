package bitmaps

var NucleotideToBitMap [128]uint64 // only the ascii ATGCNatgcn
var BitMapToNucleotide [16]byte
var NucleotideComplementMap [128]byte // a-t g-c n-n

func init() {
	// Initialize lookup arrays
	NucleotideToBitMap['A'] = 0x8 // 1000
	NucleotideToBitMap['T'] = 0x4 // 0100
	NucleotideToBitMap['G'] = 0x2 // 0010
	NucleotideToBitMap['C'] = 0x1 // 0001
	NucleotideToBitMap['N'] = 0xF // 1111 (ambiguous nucleotide)
	NucleotideToBitMap['a'] = 0x8 // Support lowercase
	NucleotideToBitMap['t'] = 0x4
	NucleotideToBitMap['g'] = 0x2
	NucleotideToBitMap['c'] = 0x1
	NucleotideToBitMap['n'] = 0xF

	BitMapToNucleotide[0x8] = 'A'
	BitMapToNucleotide[0x4] = 'T'
	BitMapToNucleotide[0x2] = 'G'
	BitMapToNucleotide[0x1] = 'C'
	BitMapToNucleotide[0xF] = 'N'

	NucleotideComplementMap['A'] = 'T'
	NucleotideComplementMap['T'] = 'A'
	NucleotideComplementMap['G'] = 'C'
	NucleotideComplementMap['C'] = 'G'
	NucleotideComplementMap['N'] = 'N'
	NucleotideComplementMap['a'] = 'T'
	NucleotideComplementMap['t'] = 'A'
	NucleotideComplementMap['g'] = 'C'
	NucleotideComplementMap['c'] = 'G'
	NucleotideComplementMap['n'] = 'N'
}
