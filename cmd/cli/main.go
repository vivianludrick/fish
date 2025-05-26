package main

import (
	"fmt"
	"os"
	// "runtime"
	"sync"
	"vivalchemy/cris/internal/parsers"
	"vivalchemy/cris/internal/utils"
	"vivalchemy/cris/pkg/encoders"
	"vivalchemy/cris/pkg/models"
	"vivalchemy/cris/pkg/models/pools"
	"vivalchemy/cris/pkg/processors"
)

var (
	TARGET_FILE = "./test.fna"
	TARGET_PAM  = "CTAATAGGAGAGTATGCTGATGG"
	BUFFER_SIZE = 1024 * 1024 * 4 // 4MB buffer for better I/O performance
	CACHE_DIR   = ".pam_cache"
	numWorkers  = 1 // max(1, runtime.NumCPU()/2)
)

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

	patternConfig := &models.PatternConfig{
		Segments: []*models.SegmentConfig{
			{Size: 13, AllowedMismatch: 4},
			{Size: 10, AllowedMismatch: 4},
		},
		MaxMismatchAllowed: 4,
		TotalSize:          23,
	}

	targetPattern, err := encoders.SegmentAndEncodePattern(TARGET_PAM, patternConfig)
	if err != nil {
		fmt.Println(err)
		return
	}
	utils.DebugPrint("Target Pattern:", targetPattern)

	var processorWg sync.WaitGroup
	var parserWg sync.WaitGroup

	fastaRecordsChan := make(chan *pools.FastaRecord, numWorkers*2)
	matchedPatternChan := make(chan *pools.MatchedPattern, numWorkers*2)
	// aggregatedResultsChan := make(chan *PatternResults, 1)

	parserWg.Add(1)
	go utils.TimeFunction("Parse FASTA File", func() {
		defer close(fastaRecordsChan)
		parsers.ParseFastaFileOrDecodeCache(TARGET_FILE, CACHE_DIR, BUFFER_SIZE, fastaRecordsChan)
		parserWg.Done()
	})

	for range numWorkers {
		processorWg.Add(1)
		go func() {
			processors.ProcessFastaRecord(targetPattern, patternConfig, fastaRecordsChan, matchedPatternChan)
			processorWg.Done()
		}()
	}

	go func() {
		defer close(matchedPatternChan)
		processorWg.Wait()
	}()
	processors.ProcessResults(matchedPatternChan)

	// waiters
	parserWg.Wait()
}
