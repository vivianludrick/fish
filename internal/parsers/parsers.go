package parsers

import (
	"bufio"
	"encoding/gob"
	"fmt"
	"io"
	"os"
	"vivalchemy/cris/internal/utils"
	"vivalchemy/cris/pkg/constants"
	"vivalchemy/cris/pkg/models"
	"vivalchemy/cris/pkg/models/pools"
)

var (
// _CHUNK_SIZE = 1024 * 1024 * 4 // 4MB
)

func sendFastaRecordInChunks(config *models.EncodedSearch, record *pools.FastaRecord, fastaRecordChunksChan chan<- *pools.FastaRecordChunk) {
	sequence := record.Sequence
	seqLen := len(sequence)
	step := constants.CHUNK_SIZE - int(config.GuideLength) + 1

	if seqLen < config.GuideLength || step <= 0 {
		// utils.DebugPrintln("Send In Chunks:", "Error: Sequence length is less than config.TotalSize")
		return
	}

	for i := 0; i < seqLen-config.GuideLength; i += step {
		end := min(i+constants.CHUNK_SIZE, seqLen)
		if end-i < config.GuideLength {
			break
		}
		chunk := &pools.FastaRecordChunk{
			Header:   record.Header,
			Sequence: sequence[i:end],
			Offset:   i,
		}
		// fmt.Printf("Chunk %v: %v len(%v)\n", i, string(chunk.Sequence), len(chunk.Sequence))
		fastaRecordChunksChan <- chunk
	}
}

func parseAndEncodeFastaFile(config *models.EncodedSearch, fastaRecordChunksChan chan<- *pools.FastaRecordChunk) error {
	fastaFile, err := os.Open(config.TargetFilePath)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return err
	}
	defer fastaFile.Close()

	scanner := bufio.NewScanner(fastaFile)
	buffer := make([]byte, constants.BUFFER_SIZE)
	scanner.Buffer(buffer, constants.BUFFER_SIZE)

	if err := utils.EnsureDir(constants.CACHE_DIR); err != nil {
		return err
	}
	cacheFileName := utils.GetCacheFileName(constants.CACHE_DIR, config.TargetFilePath)
	cacheFile, err := os.OpenFile(cacheFileName, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return err
	}
	defer cacheFile.Close()

	encoder := gob.NewEncoder(cacheFile)

	var currentEntry *pools.FastaRecord = nil

	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == 0 {
			continue
		}

		if line[0] == '>' {
			if currentEntry != nil {
				if err := encoder.Encode(currentEntry); err != nil {
					fmt.Println("Error encoding record:", err)
				}
				sendFastaRecordInChunks(config, currentEntry, fastaRecordChunksChan)
			}
			currentEntry = pools.NewFastaRecord()
			currentEntry.Header = line[1:]
		} else {
			currentEntry.Sequence = append(currentEntry.Sequence, line...)
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Scanner error:", err)
		return err
	}

	if currentEntry != nil {
		if err := encoder.Encode(currentEntry); err != nil {
			fmt.Println("Error encoding record:", err)
		}
		sendFastaRecordInChunks(config, currentEntry, fastaRecordChunksChan)
	}

	return nil
}

func decodeFastaCacheFile(config *models.EncodedSearch, cacheFileName string, fastaRecordChunksChan chan<- *pools.FastaRecordChunk) error {
	cacheFile, err := os.Open(cacheFileName)
	if err != nil {
		return err
	}
	defer cacheFile.Close()

	decoder := gob.NewDecoder(cacheFile)
	var currentEntry *pools.FastaRecord = nil

	for {
		currentEntry = pools.NewFastaRecord()
		err := decoder.Decode(currentEntry)
		if err == io.EOF {
			currentEntry.Release()
			break
		}
		if err != nil {
			currentEntry.Release()
			return err
		}
		sendFastaRecordInChunks(config, currentEntry, fastaRecordChunksChan)
	}
	return nil
}

// first decode or not then parse and encode
func ParseFastaFileOrDecodeCache(config *models.EncodedSearch, fastaRecordChunksChan chan<- *pools.FastaRecordChunk) error {
	cacheFileName := utils.GetCacheFileName(constants.CACHE_DIR, config.TargetFilePath)

	if !utils.IsFileModifiedAfterCaching(cacheFileName, config.TargetFilePath) {
		if err := decodeFastaCacheFile(config, cacheFileName, fastaRecordChunksChan); err == nil {
			return nil
		}

		fmt.Println("Error decoding cache file")
	}

	// Either cache file doesn't exist or decoding failed
	return parseAndEncodeFastaFile(config, fastaRecordChunksChan)
}

var _ = ParseFastaFileOrDecodeCache
