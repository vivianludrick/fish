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
	genomeFilePath     = "./genomes/GCF_905237065.1_Ssal_v3.1_genomic.fna"
	targetPAM          = "CTCCTGTATTTAGGAGGCTCNGG"
	ioBufferSize       = 1024 * 1024 * 4 // 4MB
	cacheDirectory     = ".pam_cache"
	numWorkers         = runtime.NumCPU() / 2
	maxMismatchAllowed = 4
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

type FastaEntry struct {
	Header   string `json:"header"`
	Sequence []byte `json:"sequence"`
}

func (fe *FastaEntry) Release() {
	fe.Header = ""
	fe.Sequence = nil
	fastaEntryPool.Put(fe)
}

func NewFastaEntry() *FastaEntry {
	return fastaEntryPool.Get().(*FastaEntry)
}

var fastaEntryPool = sync.Pool{
	New: func() any {
		return &FastaEntry{}
	},
}

type FastaEntries []*FastaEntry

type PatternMatch struct {
	PAM             string `json:"PAM"`
	MismatchIndices []int  `json:"mismatchIndices"`
	Header          string `json:"header"`
	Offset          uint32 `json:"offset"`
}

func (pm *PatternMatch) Release() {
	pm.PAM = ""
	pm.Header = ""
	pm.MismatchIndices = nil
	patternMatchPool.Put(pm)
}

func NewPatternMatch() *PatternMatch {
	return patternMatchPool.Get().(*PatternMatch)
}

var patternMatchPool = sync.Pool{
	New: func() any {
		return &PatternMatch{}
	},
}

type PatternLocation struct {
	Header string `json:"header"`
	Offset uint32 `json:"offset"`
}

type PatternMatchInfo struct {
	Locations       []*PatternLocation `json:"locations"`
	MismatchIndices []int              `json:"mismatchIndices"`
}

type PatternResults map[string]*PatternMatchInfo

func parseCLIArguments() {
	if len(os.Args) > 1 {
		genomeFilePath = os.Args[1]
	}
	if len(os.Args) > 2 {
		targetPAM = os.Args[2]
	}
	if len(os.Args) > 3 {
		cacheDirectory = os.Args[3]
	}
}

func parseFastaFile(fileName string, bufferSize int, fastaChan chan<- *FastaEntry) {
	file, err := os.Open(fileName)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var currentEntry *FastaEntry

	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == 0 {
			continue
		}
		if line[0] == '>' {
			if currentEntry != nil {
				fastaChan <- currentEntry
			}
			currentEntry = NewFastaEntry()
			currentEntry.Header = line[1:]
			currentEntry.Sequence = make([]byte, 0, bufferSize)
		} else {
			currentEntry.Sequence = append(currentEntry.Sequence, line...)
		}
	}
	if currentEntry != nil {
		fastaChan <- currentEntry
	}
}

func countMismatchAndIndices(seq, target string) (int, []int) {
	mismatch := 0
	indices := make([]int, 0)
	for i := range seq {
		if seq[i] != target[i] {
			mismatch++
			indices = append(indices, i)
		}
	}
	return mismatch, indices
}

func scanForPatternMatches(target string, in <-chan *FastaEntry, out chan<- *PatternMatch) {
	length := len(target)
	for entry := range in {
		seqLen := len(entry.Sequence)
		if seqLen < length {
			continue
		}
		maxStart := seqLen - length
		for i := 0; i <= maxStart; i++ {
			segment := string(entry.Sequence[i : i+length])
			mismatch, indices := countMismatchAndIndices(segment, target)
			if mismatch <= maxMismatchAllowed {
				pm := patternMatchPool.Get().(*PatternMatch)
				pm.PAM = target
				pm.MismatchIndices = indices
				pm.Header = entry.Header
				pm.Offset = uint32(i)
				out <- pm
			}
		}
		entry.Release()
	}
}

func aggregateMatchResults(in <-chan *PatternMatch) *PatternResults {
	results := make(PatternResults)
	for match := range in {
		if _, exists := results[match.PAM]; !exists {
			results[match.PAM] = &PatternMatchInfo{
				Locations:       make([]*PatternLocation, 0),
				MismatchIndices: match.MismatchIndices,
			}
		}
		results[match.PAM].Locations = append(results[match.PAM].Locations, &PatternLocation{
			Header: match.Header,
			Offset: match.Offset,
		})
		match.Release()
	}
	return &results
}

func main() {
	parseCLIArguments()
	log.Println("Genome File:", genomeFilePath)
	log.Println("Target PAM:", targetPAM)
	log.Println("Workers:", numWorkers)
	log.Println("Max Mismatch:", maxMismatchAllowed)

	wg := sync.WaitGroup{}
	fastaChan := make(chan *FastaEntry, numWorkers*2)
	matchChan := make(chan *PatternMatch, numWorkers*2)
	aggregatedResultsChan := make(chan *PatternResults, 1)

	wg.Add(1)
	go timeFunction("Parse FASTA File", func() {
		parseFastaFile(genomeFilePath, ioBufferSize, fastaChan)
		close(fastaChan)
		wg.Done()
	})

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			scanForPatternMatches(targetPAM, fastaChan, matchChan)
			wg.Done()
		}()
	}

	go func() {
		aggregatedResultsChan <- aggregateMatchResults(matchChan)
		close(aggregatedResultsChan)
	}()

	wg.Wait()
	close(matchChan)
	results := <-aggregatedResultsChan

	jsonOutput, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal results: %v", err)
	}
	log.Println("Results:\n", string(jsonOutput))
}

func timeFunction(name string, fn func()) {
	start := time.Now()
	fn()
	log.Printf("%s took %v\n", name, time.Since(start))
}
