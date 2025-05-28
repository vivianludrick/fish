//go:build ignore

package main

import "fmt"

var input = "ACTAGCTAGCCTAGCTAGCTACG"

func main() {

	nucleotideHexMap := map[byte]string{
		'A': "8",
		'T': "4",
		'G': "2",
		'C': "1",
		'N': "f",
		'a': "8",
		't': "4",
		'g': "2",
		'c': "1",
		'n': "f",
	}

	nucleotideBinaryMap := map[byte]string{
		'A': "1000",
		'T': "0100",
		'G': "0010",
		'C': "0001",
		'N': "1111",
		'a': "1000",
		't': "0100",
		'g': "0010",
		'c': "0001",
		'n': "1111",
	}

	hexoutput := ""
	binaryoutput := ""

	for i := range input {
		hexoutput += nucleotideHexMap[input[i]]
		binaryoutput += nucleotideBinaryMap[input[i]]
	}

	fmt.Println("Length: ", len(input))
	fmt.Println("Hex representation: ", hexoutput)
	fmt.Println("Binary representation: ", binaryoutput)
}
