package parsers

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"vivalchemy/cris/pkg/models"
)

// ParseGTFExons reads a GTF file and builds an ExonIndex containing all exon intervals.
// GTF format: tab-separated with columns: seqname, source, feature, start, end, score, strand, frame, attributes
// Only rows with feature == "exon" are kept. Coordinates are 1-based inclusive (standard GTF).
func ParseGTFExons(gtfFilePath string, bufferSize int) (*models.ExonIndex, error) {
	file, err := os.Open(gtfFilePath)
	if err != nil {
		return nil, fmt.Errorf("error opening GTF file %s: %w", gtfFilePath, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	buf := make([]byte, bufferSize)
	scanner.Buffer(buf, bufferSize)

	index := models.NewExonIndex()
	exonCount := 0

	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == 0 || line[0] == '#' {
			continue
		}

		// Fast check: skip lines that don't contain "exon" before splitting
		feature, seqname, start, end, ok := parseGTFExonLine(line)
		if !ok || feature != "exon" {
			continue
		}

		index.AddExon(seqname, start, end)
		exonCount++
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scanner error reading GTF: %w", err)
	}

	index.MergeOverlapping()
	fmt.Printf("Parsed %d exon records across %d sequences\n", exonCount, len(index.Regions))
	return index, nil
}

// parseGTFExonLine extracts feature, seqname, start, end from a GTF line.
// Returns ok=false if the line is malformed.
func parseGTFExonLine(line string) (feature string, seqname string, start int, end int, ok bool) {
	// GTF columns: seqname \t source \t feature \t start \t end \t ...
	// We need columns 0, 2, 3, 4
	fields := strings.SplitN(line, "\t", 6) // only need first 5 columns
	if len(fields) < 5 {
		return "", "", 0, 0, false
	}

	feature = fields[2]
	if feature != "exon" {
		return feature, "", 0, 0, true // valid line but not exon — caller checks feature
	}

	seqname = fields[0]

	start, err := strconv.Atoi(fields[3])
	if err != nil {
		return "", "", 0, 0, false
	}

	end, err = strconv.Atoi(fields[4])
	if err != nil {
		return "", "", 0, 0, false
	}

	return feature, seqname, start, end, true
}
