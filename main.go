package main

import (
	"fmt"
	"os"
	"time"
)

var (
	TARGET_FILE = "./GCA_015244755.2_TenIli1.0_genomic.fna"
	TARGET_PAM  = "CTCCTGTATTTAGGAGGCTCNGG"
	BUFFER_SIZE = 1024 * 1024 * 4 // 4MB buffer for better I/O performance
	CACHE_DIR   = ".pam_cache"
)

type fastaRecord struct {
	Header   string
	Sequence []byte
}

func cliArgumentInitialization() {
	if len(os.Args) > 1 {
		TARGET_FILE = os.Args[1]
	}
	if len(os.Args) > 2 {
		TARGET_PAM = os.Args[2]
	}
	if len(os.Args) > 3 {
		CACHE_DIR = os.Args[3]
	}
}

func main() {
	cliArgumentInitialization()

}

// --- utils
func timeFunction(name string, fn func()) {
	start := time.Now()
	fn()
	fmt.Printf("%s took %v\n", name, time.Since(start))
}
