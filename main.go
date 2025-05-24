package main

import (
	"bufio"
	"encoding/json"
	"log"
	"os"
	"runtime"
	"sync"
	"time"
)

var (
	TARGET_FILE  = "./genomes/GCF_905237065.1_Ssal_v3.1_genomic.fna"
	TARGET_PAM   = "CTCCTGTATTTAGGAGGCTCNGG"
	bufferSize   = 1024 * 1024 * 4 // 4MB buffer for better I/O performance
	cacheDir     = ".pam_cache"
	NUM_WORKERS  = runtime.NumCPU() / 2
	MAX_MISMATCH = 4 // Define a constant for max allowed mismatch
)

// Configuration for pattern matching
type SegmentConfig struct {
	Size            int `json:"size"`
	AllowedMismatch int `json:"allowedMismatch"`
}

type PatternConfig struct {
	Segments           []SegmentConfig `json:"segments"`
	MaxMismatchAllowed int             `json:"maxMismatchAllowed"`
	TotalSize          int             `json:"totalSize"`
}

type FastaRecord struct {
	Header   string `json:"header"`
	Sequence []byte `json:"sequence"`
}

var fastaRecordPool = sync.Pool{
	New: func() any {
		return &FastaRecord{}
	},
}

type FastaFile []*FastaRecord

// [
//
//	"<pattern>":{
//		location: [ {"header" : <header>, "offset", <offset>} ],
//		mismatch: <mismatch_locations>
//	},
//
// ]
type Results map[string]*MatchInfo

type SinglePatternMatch struct {
	PAM             string `json:"PAM"`
	MismatchIndices []int  `json:"mismatchIndices"`
	Header          string `json:"header"`
	Offset          uint32 `json:"offset"`
}

var singlePatternMatchPool = sync.Pool{
	New: func() any {
		return &SinglePatternMatch{}
	},
}

type MatchInfo struct {
	Locations       []*MatchLocation `json:"locations"`
	MismatchIndices []int            `json:"mismatchIndices"` // Store indices of mismatches
}

type MatchLocation struct {
	Header string `json:"header"`
	Offset uint32 `json:"offset"`
}

func cliArgumentInitialization() {
	if len(os.Args) > 1 {
		TARGET_FILE = os.Args[1]
	}
	if len(os.Args) > 2 {
		TARGET_PAM = os.Args[2]
	}
	if len(os.Args) > 3 {
		cacheDir = os.Args[3]
	}
}

func readFile(fileName string, bufferSize int, fastRecordChan chan<- *FastaRecord) {
	file, err := os.Open(fileName)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	var currentFastaRecord *FastaRecord

	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == 0 {
			continue
		}

		if line[0] == '>' {
			if currentFastaRecord != nil {
				fastRecordChan <- currentFastaRecord
				fastaRecordPool.Put(currentFastaRecord)
			}
			currentFastaRecord = fastaRecordPool.Get().(*FastaRecord)
			currentFastaRecord.Header = line[1:]
			currentFastaRecord.Sequence = make([]byte, 0, bufferSize)
		} else {
			currentFastaRecord.Sequence = append(currentFastaRecord.Sequence, line...)
		}
	}

	if currentFastaRecord != nil {
		fastRecordChan <- currentFastaRecord
	}
	close(fastRecordChan)
}

func countMismatchAndIndices(input string, target string) (int, []int) {
	mismatch := 0
	indices := make([]int, 0)
	for i := range input {
		if input[i] != target[i] {
			mismatch++
			indices = append(indices, i)
		}
	}
	return mismatch, indices
}

func processFastaRecord(targetPAM string, fastaRecordChan <-chan *FastaRecord, patternMatchChan chan<- *SinglePatternMatch) {
	var singlePatternMatch *SinglePatternMatch
	pamLen := len(targetPAM)
	for fastaRecord := range fastaRecordChan {
		seqLen := len(fastaRecord.Sequence)
		if seqLen < pamLen {
			continue
		}
		log.Println("seq len:", seqLen, "pam len:", pamLen)
		for i := 0; i <= seqLen-pamLen; i++ {
			segment := string(fastaRecord.Sequence[i : i+pamLen])
			mismatch, mismatchIndices := countMismatchAndIndices(segment, targetPAM)
			if mismatch <= MAX_MISMATCH {
				singlePatternMatch = singlePatternMatchPool.Get().(*SinglePatternMatch)
				singlePatternMatch.MismatchIndices = mismatchIndices
				singlePatternMatch.PAM = targetPAM
				singlePatternMatch.Header = fastaRecord.Header
				singlePatternMatch.Offset = uint32(i)
				patternMatchChan <- singlePatternMatch
			}
		}
		fastaRecordPool.Put(fastaRecord) // Return the processed record to the pool
	}
	close(patternMatchChan) // Close the output channel when all records are processed
}

func processResults(patternMatchChan <-chan *SinglePatternMatch) *Results {
	results := make(Results)
	for matchInfo := range patternMatchChan {
		if _, ok := results[matchInfo.PAM]; !ok {
			results[matchInfo.PAM] = &MatchInfo{
				Locations:       make([]*MatchLocation, 0),
				MismatchIndices: matchInfo.MismatchIndices,
			}
		}
		results[matchInfo.PAM].Locations = append(results[matchInfo.PAM].Locations, &MatchLocation{
			Header: matchInfo.Header,
			Offset: matchInfo.Offset,
		})
		// Recalculate mismatch and indices here if needed for detailed reporting per match
		segment := TARGET_FILE // This is wrong, need access to the original sequence
		_ = segment            // Avoid unused variable error
		// mismatch, indices := countMismatchAndIndices(segment, matchInfo.PAM)
		// results[matchInfo.PAM].MismatchIndices = append(results[matchInfo.PAM].MismatchIndices, indices)

		singlePatternMatchPool.Put(matchInfo)
	}
	return &results
}

func main() {
	cliArgumentInitialization()
	log.Println("Target File:", TARGET_FILE)
	log.Println("Target PAM:", TARGET_PAM)
	log.Println("Number of Workers:", NUM_WORKERS)
	log.Println("Max Mismatch Allowed:", MAX_MISMATCH)

	wg := sync.WaitGroup{}
	fastaRecordChan := make(chan *FastaRecord, NUM_WORKERS*2)
	patternMatchChan := make(chan *SinglePatternMatch, NUM_WORKERS*2)
	resultsChan := make(chan *Results, 1) // Channel to receive the final results

	wg.Add(1)
	go timeFunction("Read File", func() {
		readFile(TARGET_FILE, bufferSize, fastaRecordChan)
		wg.Done()
	})

	for i := 0; i < NUM_WORKERS; i++ {
		wg.Add(1)
		go func() {
			processFastaRecord(TARGET_PAM, fastaRecordChan, patternMatchChan)
			wg.Done()
		}()
	}

	go func() {
		wg.Wait()
		close(patternMatchChan)
		resultsChan <- processResults(patternMatchChan)
		close(resultsChan)
	}()

	results := <-resultsChan
	resultsJSON, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		log.Fatalf("Error marshaling results to JSON: %v", err)
	}
	log.Println("Results:\n", string(resultsJSON))
}

func timeFunction(name string, fn func()) {
	start := time.Now()
	fn()
	log.Printf("%s took %v\n", name, time.Since(start))
}
