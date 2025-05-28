package utils

import (
	"fmt"
	"os"
	"time"
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
