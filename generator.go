package main

import (
	"fmt"
	"math/rand/v2"
)

func main() {
	pattern := "CTAATAGGAGAGTATGCTGATGG"

	randomPlaces := []int{}

	for range rand.IntN(6) {
		randomPlaces = append(randomPlaces, rand.IntN(len(pattern)))
	}

	nucleoTideArray := []string{"A", "T", "G", "C"}

	for _, val := range randomPlaces {
		newPattern := pattern[:val] + nucleoTideArray[rand.IntN(len(nucleoTideArray))] + pattern[val+1:]
		fmt.Println(newPattern)
	}
}
