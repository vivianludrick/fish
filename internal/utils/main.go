package utils

import (
	"fmt"
	"time"
)

func TimeFunction(name string, fn func()) {
	start := time.Now()
	fn()
	fmt.Printf("%s took %v\n", name, time.Since(start))
}
