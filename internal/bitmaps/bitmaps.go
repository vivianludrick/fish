package bitmaps

var NucleotideToBitMap [256]uint64
var BitMapToNucleotide [16]byte

func init() {
	// Initialize lookup arrays
	NucleotideToBitMap['A'] = 0x8 // 1000
	NucleotideToBitMap['C'] = 0x4 // 0100
	NucleotideToBitMap['G'] = 0x2 // 0010
	NucleotideToBitMap['T'] = 0x1 // 0001
	NucleotideToBitMap['N'] = 0xF // 1111 (ambiguous nucleotide)
	NucleotideToBitMap['a'] = 0x8 // Support lowercase
	NucleotideToBitMap['c'] = 0x4
	NucleotideToBitMap['g'] = 0x2
	NucleotideToBitMap['t'] = 0x1
	NucleotideToBitMap['n'] = 0xF

	BitMapToNucleotide[0x8] = 'A'
	BitMapToNucleotide[0x4] = 'C'
	BitMapToNucleotide[0x2] = 'G'
	BitMapToNucleotide[0x1] = 'T'
	BitMapToNucleotide[0xF] = 'N'
}
