package main

import (
	"bufio"
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

func parseFastaFile(fileName string) {
	// open the file
	file, err := os.Open(fileName)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()

	// scan the file and pass it to the
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if line[0] == '>' {
			header := line[1:]
			sequence := make([]byte, 0)
			for scanner.Scan() {
				line := scanner.Text()
				sequence = append(sequence, line...)
			}
			fmt.Println(header)
			fmt.Println(string(sequence))
		}
	}
}

func main() {
	cliArgumentInitialization()

}
