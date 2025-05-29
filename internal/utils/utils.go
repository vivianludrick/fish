package utils

import (
	"fmt"
	"os"
	"time"
	"vivalchemy/cris/internal/bitmaps"
)

func TimeFunction(name string, fn func()) {
	start := time.Now()
	fn()
	fmt.Printf("%s took %v\n", name, time.Since(start))
}

func PrintPattern(pattern []uint64) {
	var str string
	for _, segment := range pattern {
		str += fmt.Sprintf("%016b", segment)
	}
	fmt.Println(str)
}

func DebugPrintln(label string, data any) {
	if os.Getenv("DEBUG") != "true" {
		return
	}
	fmt.Println("--------------------------------------------------------")
	fmt.Println(label)
	fmt.Println("--------------------------------------------------------")
	fmt.Println(data)
}

func DebugRun(label string, fn func()) {
	if os.Getenv("DEBUG") != "true" {
		return
	}
	fmt.Println("--------------------------------------------------------")
	fmt.Println(label)
	fmt.Println("--------------------------------------------------------")
	fn()
}

func Prepend[Type any](slice []Type, elems ...Type) []Type {
	return append(elems, slice...)
}

func ReverseComplement(sequence string) string {
	n := len(sequence)
	rc := make([]byte, n)
	for i, r := range sequence {
		rc[n-1-i] = bitmaps.NucleotideComplementMap[r]
	}
	return string(rc)
}

// Reverse returns a reversed copy of the input slice.
func Reverse[T any](input []T) []T {
	reversed := make([]T, len(input))
	for i, v := range input {
		reversed[len(input)-1-i] = v
	}
	return reversed
}
