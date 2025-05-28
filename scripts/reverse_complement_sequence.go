//go:build ignore

package main

import "fmt"

var input = "CGTAGCTAGCTAGGCTAGCTAGT"

var complement = map[rune]rune{
	'A': 'T',
	'C': 'G',
	'G': 'C',
	'T': 'A',
	'N': 'N',
	'a': 'T',
	'c': 'G',
	'g': 'C',
	't': 'A',
	'n': 'N',
}

func main() {
	inputRune := []rune(input)
	// reverse input
	for i := range len(inputRune) / 2 {
		inputRune[i], inputRune[len(input)-1-i] = inputRune[len(input)-1-i], inputRune[i]
	}
	// complement
	rc := ""
	for _, r := range inputRune {
		rc += string(complement[r])
	}
	fmt.Println(rc)
}
