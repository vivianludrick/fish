package parsers

// import (
// 	"bufio"
// 	"encoding/gob"
// 	"fmt"
// 	"io"
// 	"os"
// 	"vivalchemy/cris/internal/utils"
// 	"vivalchemy/cris/pkg/models"
// 	"vivalchemy/cris/pkg/models/pools"
// )
//
// var (
// 	CHUNK_SIZE = 1024 * 1024 * 4 // 4MB
// )
//
// // no caching only parsing
// func ParseFastaFile(fileName string, bufferSize int, fastaRecordChunksChan chan<- *pools.FastaRecordChunk, config *models.PatternConfig) {
//
// 	file, err := os.Open(fileName)
// 	if err != nil {
// 		fmt.Println("Error opening file:", err)
// 		return
// 	}
// 	defer file.Close()
//
// 	scanner := bufio.NewScanner(file)
// 	buffer := make([]byte, bufferSize)
// 	scanner.Buffer(buffer, bufferSize)
//
// 	var currentEntry *pools.FastaRecord = nil
//
// 	for scanner.Scan() {
// 		line := scanner.Text()
// 		if len(line) == 0 {
// 			continue
// 		}
//
// 		if line[0] == '>' {
// 			if currentEntry != nil {
// 				sendInChunks(fastaRecordChunksChan, currentEntry, CHUNK_SIZE, config)
// 			}
// 			currentEntry = pools.NewFastaRecord()
// 			currentEntry.Header = line[1:]
// 			currentEntry.Sequence = make([]byte, 0, 1024*8)
// 		} else {
// 			currentEntry.Sequence = append(currentEntry.Sequence, line...)
// 		}
// 	}
//
// 	if err := scanner.Err(); err != nil {
// 		fmt.Println("Scanner error:", err)
// 		return
// 	}
//
// 	if currentEntry != nil {
// 		sendInChunks(fastaRecordChunksChan, currentEntry, CHUNK_SIZE, config)
// 	}
// }
//
// func sendInChunks(fastaRecordChunksChan chan<- *pools.FastaRecordChunk, record *pools.FastaRecord, chunkSize int, config *models.PatternConfig) {
// 	sequence := record.Sequence
// 	seqLen := len(sequence)
// 	step := chunkSize - config.TotalSize + 1
//
// 	if seqLen < config.TotalSize || step <= 0 {
// 		// utils.DebugPrintln("Send In Chunks:", "Error: Sequence length is less than config.TotalSize")
// 		return
// 	}
//
// 	// fmt.Println(chunkSize, "chunkSize")
// 	for i := 0; i < seqLen-config.TotalSize; i += step {
// 		end := min(i+chunkSize, seqLen)
// 		if end-i < config.TotalSize {
// 			break
// 		}
// 		chunk := &pools.FastaRecordChunk{
// 			Header:   record.Header,
// 			Sequence: sequence[i:end],
// 			Offset:   i,
// 		}
// 		// fmt.Printf("Chunk %v: %v len(%v)\n", i, string(chunk.Sequence), len(chunk.Sequence))
// 		fastaRecordChunksChan <- chunk
// 	}
// }
//
// func parseAndEncodeFastaFile(fileName string, cacheDir string, bufferSize int, fastaRecordChunksChan chan<- *pools.FastaRecordChunk, config *models.PatternConfig) error {
// 	fastaFile, err := os.Open(fileName)
// 	if err != nil {
// 		fmt.Println("Error opening file:", err)
// 		return err
// 	}
// 	defer fastaFile.Close()
//
// 	scanner := bufio.NewScanner(fastaFile)
// 	buffer := make([]byte, bufferSize)
// 	scanner.Buffer(buffer, bufferSize)
//
// 	if err := utils.EnsureDir(cacheDir); err != nil {
// 		return err
// 	}
// 	cacheFileName := utils.GetCacheFileName(cacheDir, fileName)
// 	cacheFile, err := os.OpenFile(cacheFileName, os.O_RDWR|os.O_CREATE, 0644)
// 	if err != nil {
// 		fmt.Println("Error opening file:", err)
// 		return err
// 	}
// 	defer cacheFile.Close()
//
// 	encoder := gob.NewEncoder(cacheFile)
//
// 	var currentEntry *pools.FastaRecord = nil
//
// 	for scanner.Scan() {
// 		line := scanner.Text()
// 		if len(line) == 0 {
// 			continue
// 		}
//
// 		if line[0] == '>' {
// 			if currentEntry != nil {
// 				if err := encoder.Encode(currentEntry); err != nil {
// 					fmt.Println("Error encoding record:", err)
// 				}
// 				sendInChunks(fastaRecordChunksChan, currentEntry, CHUNK_SIZE, config)
// 			}
// 			currentEntry = pools.NewFastaRecord()
// 			currentEntry.Header = line[1:]
// 		} else {
// 			currentEntry.Sequence = append(currentEntry.Sequence, line...)
// 		}
// 	}
//
// 	if err := scanner.Err(); err != nil {
// 		fmt.Println("Scanner error:", err)
// 		return err
// 	}
//
// 	if currentEntry != nil {
// 		if err := encoder.Encode(currentEntry); err != nil {
// 			fmt.Println("Error encoding record:", err)
// 		}
// 		sendInChunks(fastaRecordChunksChan, currentEntry, CHUNK_SIZE, config)
// 	}
//
// 	return nil
// }
//
// func decodeFastaCacheFile(cacheFileName string, fastaRecordChunksChan chan<- *pools.FastaRecordChunk, config *models.PatternConfig) error {
// 	cacheFile, err := os.Open(cacheFileName)
// 	if err != nil {
// 		return err
// 	}
// 	defer cacheFile.Close()
//
// 	decoder := gob.NewDecoder(cacheFile)
// 	var currentEntry *pools.FastaRecord = nil
//
// 	for {
// 		currentEntry = pools.NewFastaRecord()
// 		err := decoder.Decode(currentEntry)
// 		if err == io.EOF {
// 			currentEntry.Release()
// 			break
// 		}
// 		if err != nil {
// 			currentEntry.Release()
// 			return err
// 		}
// 		sendInChunks(fastaRecordChunksChan, currentEntry, CHUNK_SIZE, config)
// 	}
// 	return nil
// }
//
// // first decode or not then parse and encode
// func ParseFastaFileOrDecodeCache(fileName string, cacheDir string, bufferSize int, fastaRecordChunksChan chan<- *pools.FastaRecordChunk, config *models.PatternConfig) error {
// 	cacheFileName := utils.GetCacheFileName(cacheDir, fileName)
//
// 	if !utils.IsFileModifiedAfterCaching(cacheFileName, fileName) {
// 		if err := decodeFastaCacheFile(cacheFileName, fastaRecordChunksChan, config); err == nil {
// 			return nil
// 		}
//
// 		fmt.Println("Error decoding cache file")
// 	}
//
// 	// Either cache file doesn't exist or decoding failed
// 	return parseAndEncodeFastaFile(fileName, cacheDir, bufferSize, fastaRecordChunksChan, config)
// }

// -------

// // 1.5seconds
// parserWg.Add(1)
// go utils.TimeFunction("Parse FASTA File", func() {
// 	parsers.ParseFastaFile(TARGET_FILE, BUFFER_SIZE, fastaRecordsChan)
// 	defer close(fastaRecordsChan)
// 	parserWg.Done()
// })

// 2seconds
// wg.Add(1)
// go utils.TimeFunction("Parse FASTA File", func() {
// 	parseAndEncodeFastaFile(TARGET_FILE, fastaRecordsChan)
// 	defer close(fastaRecordsChan)
// 	wg.Done()
// })

// 500ms
// wg.Add(1)
// go utils.TimeFunction("Parse FASTA File", func() {
// 	decodeFastaCacheFile(TARGET_FILE, fastaRecordsChan)
// 	defer close(fastaRecordsChan)
// 	wg.Done()
// })

// ensemble
// parserWg.Add(1)
// go utils.TimeFunction("Parse FASTA File", func() {
// 	parsers.ParseFastaFileOrDecodeCache(TARGET_FILE, CACHE_DIR, BUFFER_SIZE, fastaRecordsChan)
// 	defer close(fastaRecordsChan)
// 	parserWg.Done()
// })
