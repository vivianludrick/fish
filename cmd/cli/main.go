package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"vivalchemy/cris/internal/parsers"
	"vivalchemy/cris/internal/utils"
	"vivalchemy/cris/pkg/encoders"
	"vivalchemy/cris/pkg/extractors"
	"vivalchemy/cris/pkg/models"
	"vivalchemy/cris/pkg/models/pools"
	"vivalchemy/cris/pkg/processors"
)

var (
	BUFFER_SIZE = 1024 * 1024 * 4 // 4MB buffer for better I/O performance
	numWorkers  = max(1, runtime.NumCPU()/2)
)

func main() {
	targetFile := flag.String("file", "./test.fna", "path to FASTA file")
	targetInput := flag.String("input", "", "single 23bp guide+PAM or DNA blob for multi-target extraction")
	exonsOnly := flag.Bool("exons-only", false, "only match within exon regions (requires GTF)")
	gtfFile := flag.String("gtf", "", "path to GTF file (auto-detected from FASTA dir if omitted)")
	cacheDir := flag.String("cache", ".pam_cache", "cache directory")
	flag.Parse()

	// Backward compat: positional args override flags
	args := flag.Args()
	if len(args) > 0 {
		*targetFile = args[0]
	}
	if len(args) > 1 {
		*targetInput = args[1]
	}
	if len(args) > 2 {
		*cacheDir = args[2]
	}

	if *targetInput == "" {
		*targetInput = "CTAATAGGAGAGTATGCTGATGG"
	}

	patternConfig := &models.PatternConfig{
		Segments: []*models.SegmentConfig{
			{Size: 13, AllowedMismatch: 4},
			{Size: 7, AllowedMismatch: 4},
			{Size: 1, AllowedMismatch: 1},
			{Size: 2, AllowedMismatch: 0},
		},
		MaxMismatchAllowed: 5,
		TotalSize:          23,
	}

	guideLen := patternConfig.TotalSize - 3 // 20bp guide, 3bp PAM

	// --- Build encoded targets ---
	var targets []processors.EncodedTarget

	if len(*targetInput) == patternConfig.TotalSize {
		// Single target mode: input is exactly 23bp
		encoded, err := encoders.SegmentAndEncodePattern(*targetInput, patternConfig)
		if err != nil {
			fmt.Println("Error encoding target:", err)
			os.Exit(1)
		}
		targets = append(targets, processors.EncodedTarget{
			Sequence: *targetInput,
			Encoded:  encoded,
			Strand:   "forward",
		})
	} else if len(*targetInput) > patternConfig.TotalSize {
		// Blob mode: extract all possible forward + reverse targets
		extracted := extractors.ExtractTargetsFromBlob(*targetInput, guideLen)
		if len(extracted) == 0 {
			fmt.Println("No valid target sequences found in input blob")
			os.Exit(1)
		}
		fmt.Printf("Extracted %d target sequences from input blob\n", len(extracted))
		for _, et := range extracted {
			encoded, err := encoders.SegmentAndEncodePattern(et.Sequence, patternConfig)
			if err != nil {
				utils.DebugPrintln("Skipping invalid target", fmt.Sprintf("%s: %v", et.Sequence, err))
				continue
			}
			targets = append(targets, processors.EncodedTarget{
				Sequence: et.Sequence,
				Encoded:  encoded,
				Strand:   et.Strand,
			})
		}
		if len(targets) == 0 {
			fmt.Println("No encodable target sequences found in blob")
			os.Exit(1)
		}
		fmt.Printf("Encoded %d targets for matching (%d forward, %d reverse)\n",
			len(targets), countStrand(targets, "forward"), countStrand(targets, "reverse"))
	} else {
		fmt.Printf("Input too short: need at least %d bases (got %d)\n", patternConfig.TotalSize, len(*targetInput))
		os.Exit(1)
	}

	// --- Parse exon index if --exons-only ---
	var exonIndex *models.ExonIndex
	if *exonsOnly {
		gtfPath := *gtfFile
		if gtfPath == "" {
			// Auto-detect: look for genomic.gtf in same directory as FASTA
			gtfPath = filepath.Join(filepath.Dir(*targetFile), "genomic.gtf")
		}
		fmt.Println("Parsing GTF for exon regions:", gtfPath)
		var err error
		exonIndex, err = parsers.ParseGTFExons(gtfPath, BUFFER_SIZE)
		if err != nil {
			fmt.Println("Error parsing GTF file:", err)
			os.Exit(1)
		}
	}

	// --- Pipeline ---
	var processorWg sync.WaitGroup
	var parserWg sync.WaitGroup

	fastaRecordChunksChan := make(chan *pools.FastaRecordChunk, numWorkers*2)
	matchedPatternChan := make(chan *pools.MatchedPattern, numWorkers*2)

	parserWg.Add(1)
	go utils.TimeFunction("Parse FASTA File", func() {
		defer close(fastaRecordChunksChan)
		parsers.ParseFastaFileOrDecodeCache(*targetFile, *cacheDir, BUFFER_SIZE, fastaRecordChunksChan, patternConfig)
		parserWg.Done()
	})

	for range numWorkers {
		processorWg.Add(1)
		go func() {
			processors.ProcessFastaRecordChunks(targets, patternConfig, exonIndex, fastaRecordChunksChan, matchedPatternChan)
			processorWg.Done()
		}()
	}

	go func() {
		defer close(matchedPatternChan)
		processorWg.Wait()
	}()
	processors.ProcessResults(matchedPatternChan)

	parserWg.Wait()
}

func countStrand(targets []processors.EncodedTarget, strand string) int {
	n := 0
	for _, t := range targets {
		if t.Strand == strand {
			n++
		}
	}
	return n
}
