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

func DebugPrint(label string, data any) {
	if os.Getenv("DEBUG") != "true" {
		return
	}
	fmt.Println("--------------------------------------------------------")
	fmt.Println("         ", label)
	fmt.Println("--------------------------------------------------------")
	fmt.Println(data)
}
