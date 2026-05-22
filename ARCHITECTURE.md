# CRIS Architecture

## Project Overview
**vivalchemy/cris** - CRISPR guide sequence search and validation tool for fish genomes (zebrafish, atlantic-salmon).

Searches genomic sequences for PAM patterns (e.g., SpCas9 guide sequences) with configurable mismatch tolerances and segment-based constraints. Includes scoring algorithms (Mitscore, Doench).

## Key Characteristics
- Go 1.24.3 module
- Parallel processing (workers = CPU/2)
- Memory pooling for FASTA records (reduce GC pressure)
- GOB-based caching for parsed genomic sequences
- Bit-encoded segment representation (4 bits per nucleotide)

## Data Flow

```
FASTA File
    ↓
ParseFastaFileOrDecodeCache (parsers.go)
    ├─ Check cache validity
    ├─ If valid: decodeFastaCacheFile (GOB)
    └─ If invalid: parseAndEncodeFastaFile (parse + encode)
    ↓
FastaRecordChunk (Header, Sequence, Offset)
    ↓ (parallel workers)
ProcessFastaRecordChunks (processors.go)
    ├─ SegmentAndEncodePattern applied to guide sequences
    ├─ Pattern matching with mismatch tolerance per segment
    ↓
MatchedPattern (results)
    ↓
ProcessResults
    ↓
Results output
```

## Core Models

### Pattern Configuration (`pkg/models/pattern_config.go`)
- **NewPatternSearchConfig**: User input (guide sequences, genome, tolerance spec, preset variant, benchmarks)
- **NewToleranceSpec**: Segment specs + max total mismatches + total guide length
- **NewSegmentTolerance**: Individual segment (length, allowed mismatches/variants)
- **Converter**: Transforms NewPatternSearchConfig → NewInternalPatternSearchConfig (bit-encoded)

### Constants (`pkg/models/constants.go`)
- **Genomes**: `danio-rerio`, `atlantic-salmon` (map to FASTA paths)
- **PresetVariants**: `sp-cas9` (23bp: 10+10+3, mismatches 4+2+0), `custom`
- **BenchmarkAlgorithms**: `mitscore`, `doench`

### Object Pools (`pkg/models/pools/fasta_record.go`)
- **FastaRecord**: Header + Sequence
- **FastaRecordChunk**: Header + Sequence + Offset (for overlapping chunks)
- Both use `sync.Pool` for reuse

### Matched Pattern (`pkg/models/pools/matched_pattern.go`)
- Results container for matches found during processing

## File I/O & Caching

**Parser** (`internal/parsers/parsers.go`) - 1.5KB
- **ParseFastaFile**: Direct line-by-line parsing, sends overlapping chunks (step = chunkSize - patternLength + 1)
- **parseAndEncodeFastaFile**: Parse + encode to GOB cache simultaneously
- **decodeFastaCacheFile**: Decode from GOB, reconstruct chunks
- **ParseFastaFileOrDecodeCache**: Smart selection based on file modification time

Benchmarks (via comments):
- Direct parse: 1.5s
- Parse + encode: 2s
- Cache decode: 500ms
- Smart selection: 500ms (if cache valid)

## Encoders & Bitmaps

**Encoders** (`pkg/encoders/encoders.go`)
- **SegmentAndEncodePattern**: Converts pattern to segment-based bit representation

**Nucleotide Bitmap** (`internal/bitmaps/nucleotide_map.go`)
- Maps A/T/G/C → 4-bit values
- Includes complement mapping for reverse strand search

## Processors

**Processors** (`pkg/processors/processors.go`)
- **ProcessFastaRecordChunks**: Main worker (receives chunks, searches for patterns)
- **ProcessResults**: Scores each match (MIT + CFD) and prints scored output

## Scoring System (`pkg/scoring/`)

Three scoring algorithms, two in pure Go, one via Python HTTP server:

### MIT Score (`mit_score.go`) — Off-target, Pure Go
- Hsu et al. 2013 positional mismatch model
- 20-element weight array + arithmetic
- `MitHitScore(guide20, offTarget20) → 0-100`
- `MitAggregateSpecificity(hitScores) → 0-100` (across all off-targets for a guide)
- Formula: `positionPenalty * distancePenalty * countPenalty * 100`

### CFD Score (`cfd_score.go`) — Off-target, Pure Go
- Doench et al. 2016 Cutting Frequency Determination
- 240-entry mismatch matrix (12 mismatch types × 20 positions) + 16 PAM penalties
- `CfdScore(guide20, offTarget20, pamDinuc) → 0-1`
- Position-specific AND mismatch-type-specific penalties
- Watson-Crick complement mismatches get no penalty (RNA still binds via WC pairing)

### Doench / Rule Set 2 (`doench_client.go`) — On-target, Python Server
- Azimuth GBR model (100 trees, max_depth 3) via `scoring_server/`
- Input: 30-mer context (4nt upstream + 20nt guide + NGG + 3nt downstream)
- `DoenchScore(sequences) → map[seq]score` (0-100 scale)
- Go client talks to Python Flask server at `DOENCH_SERVER_URL` (default: `http://127.0.0.1:8080`)
- Uses retrained Python 3 model from [CrisprScoringHub](https://github.com/Interventional-Genomics-Unit/CrisprScoringHub)
- Setup: `cd scoring_server && ./setup.sh && source venv/bin/activate && python server.py`

### Scoring Data Flow
```
MatchedPattern (23bp guide + 23bp off-target)
    ├─ Extract 20bp guide core + 20bp off-target core + PAM dinuc
    ├─ MitHitScore(guide20, offTarget20) → MIT score
    ├─ CfdScore(guide20, offTarget20, pam) → CFD score
    └─ [optional] DoenchScore(30mer) → on-target activity score
    ↓
ScoredMatch (Header, Offset, MatchedSeq, MIT, CFD)
```

## CLI Entry Point

**cmd/cli/main.go** - 88 lines
- Default: TARGET_FILE="./test.fna", TARGET_PAM="CTAATAGGAGAGTATGCTGATGG"
- Config: 23bp pattern (13+7+1+2 segments, mismatches 4+4+1+0, max 5 total)
- Workers: `runtime.NumCPU() / 2`
- Buffer: 4MB I/O buffer
- Pipeline:
  1. Parser thread (sends chunks)
  2. N worker threads (search chunks)
  3. Result aggregator thread

## Performance Optimizations
1. Object pooling (sync.Pool) → reduced GC pressure
2. GOB caching → 500ms cache decode vs 1.5s parse
3. 4MB I/O buffer + buffered scanner → faster file reading
4. Bit-encoded segments (4 bits/base) → compact representation
5. Overlapping chunks (step = chunkSize - patternLength + 1) → no missed patterns
6. Parallel workers (CPU/2) → CPU-bound parallelization

## Utilities & Helpers

**Utils** (`internal/utils/`)
- **utils.go**: TimeFunction (perf timing), PrintPattern (debug binary), DebugPrintln/DebugRun (DEBUG env check), Prepend (generic)
- **file.go**: GetCacheFileName, EnsureDir, DoesFileExists, IsFileModifiedAfterCaching (MOD time check)

**Stubs** (`internal/stubs/stubs.go`)
- FastaChanConsumerStub: test consumer for FASTA channel

## Scripts

**Development Scripts** (`scripts/`)
- **reverse_complement_sequence.go** (build:ignore): compute reverse complement of test sequence
- **sequence_converter.go** (build:ignore): visualize nucleotide → hex/binary encoding

## Data Models

### Results Structure (`pkg/models/results.go`)
```
Result[guide_sequence][matched_sequence]
  └── SequenceDetails
        ├── Locations: [{header, offset}, ...]
        └── MismatchPositions: [uint, ...]
```

### Matched Pattern Pool (`pkg/models/pools/matched_pattern.go`)
- **MatchedPattern**: guide_sequence, matched_sequence, header, offset
- JSON-serializable, object pooled

### Sliding Window (`pkg/models/pools/sliding_window.go`)
- Maintains bit-encoded segments as nucleotides are streamed
- Special-case optimization for 2-segment patterns (unrolled loop)
- AddNucleotide: shifts bits, cascades overflow
- GetSequence: reconstructs ACGT string from bit patterns

### Internal Config (`pkg/models/internal_pattern_config.go`)
- **NewInternalPatternSearchConfig**: processed config with pre-encoded guides + reverse guides
- **NewInternalSegmentTolerance**: Lengths (split large segments), AllowedMismatches, AllowedVariants (pre-encoded)
- Print() for debugging

### Legacy Models (`pkg/models/pattern_legacy.go`)
- Old PatternConfig/SegmentConfig (kept for reference)

## Genomic Data

Downloaded via NCBI Datasets (genomes/ncbi_dataset/data/):
- **Labeo rohita** (rohu, asian carp) - IGBB_LRoh.1.0 assembly
  - **GCA_022985175.1** (GenBank): 1,140,899,367 bytes FASTA + 382,270,778 bytes GTF
  - **GCF_022985175.1** (RefSeq): 1,141,024,730 bytes FASTA + 448,858,168 bytes GTF
  - Chromosome-level, 31,701 genes, N50=1.3MB
  - Metadata: data_summary.tsv, assembly_data_report.jsonl, sequence_report.jsonl

Additional genomes (not in repo, referenced in mprocs.yaml):
- GCA_015244755.2 TenIli1.0 (tenualosa ilisha/indian oil sardine)
- GCF_049306965.1 GRCz12 (zebrafish)
- GCF_905237065.1 Ssal_v3.1 (atlantic salmon)

Note: genomes/ excluded from git (.gitignore), distributed separately

## Build & Testing

**mprocs.yaml** - multiprocess dev runner:
- `test`: single run with DEBUG
- `btop`: resource monitor
- Benchmarks (hy/*): hyperfine -5 on 3 species
- Regular runs: time wrapper on same 3 species
- DEBUG env enabled for all

**Go Version**: 1.26.2

## Test Data

**input.fna** (153 bytes) - zebrafish (danio rerio) chromosome 1 sequence:
- Header: CM002885.2 GRCz11 ref assembly
- Sequence: 80bp sample for quick testing

**output*.txt** (38KB each) - example runs showing matched patterns

## Performance Characteristics

| Operation | Time |
|-----------|------|
| Parse FASTA | 1.5s |
| Parse + encode to cache | 2s |
| Decode from cache | 500ms |
| Full pipeline (smart cache) | 500ms-1.5s |

Workers: `runtime.NumCPU() / 2` (parallelizes CPU-bound pattern search)

## Development Status

**Current work** (as of main branch head):
- Type naming improvements (NewPattern* → better naming)
- Results aggregation/processing pipeline
- Validation flow refactoring
- Config compatibility updates

**Notable refactors**:
- Switched from `bytes` to `Text` for headers (fixed buffer issues)
- Introduced chunking for predictable memory usage
- Object pooling (FastaRecord, SlidingWindow, MatchedPattern)
- Bit encoding for compact segment representation

**Legacy code**:
- pattern_legacy.go: old SegmentConfig/PatternConfig (pre-New* refactor)
- Final "legacy commit" at 02260c7 (functional baseline)

## Known TODOs

1. **processors.go:12** - `TODO: Fetch the guide sequence here`
   - GUIDE_SEQUENCE currently hardcoded to "CTAATAGGAGAGTATGCTGATGG"
   - Should be sourced from command-line args or config

2. **internal_pattern_config.go:28** - `TODO: Add the reverse complement of the variants`
   - AllowedVariants only store forward strand
   - Reverse complement variants not pre-encoded
   - May affect bidirectional search accuracy

---

## File Inventory

| File | Size | Purpose |
|------|------|---------|
| **Entry Points** | | |
| main.go | 163B | Stub directing to cmd/cli/main.go |
| cmd/cli/main.go | 2.0K | CLI entry point, orchestrates parser + workers + results |
| **Models** | | |
| pkg/models/pattern_config.go | 4.5K | Public API: NewPatternSearchConfig + validation + converter |
| pkg/models/internal_pattern_config.go | 2.3K | Internal: pre-encoded configs, reverse guides, Print() |
| pkg/models/constants.go | 1.5K | Enums: Genomes, PresetVariants, BenchmarkAlgorithms |
| pkg/models/results.go | 400B | Hierarchical result structure |
| pkg/models/pattern_legacy.go | 300B | Old config types (reference only) |
| **Pools** | | |
| pkg/models/pools/fasta_record.go | 800B | sync.Pool for FastaRecord + FastaRecordChunk |
| pkg/models/pools/matched_pattern.go | 600B | MatchedPattern + pool + output formatter |
| pkg/models/pools/sliding_window.go | 2.3K | Bit-encoded sliding window, AddNucleotide, GetSequence |
| **Processing** | | |
| pkg/processors/processors.go | 3.8K | ComparePatterns (bit-level matching), ProcessFastaRecordChunks, ProcessResults |
| pkg/encoders/encoders.go | 1.3K | SegmentAndEncodePattern (validates + encodes pattern) |
| **Scoring** | | |
| pkg/scoring/scoring.go | 1.0K | ScoredMatch type, ScoreOffTarget orchestrator |
| pkg/scoring/mit_score.go | 1.8K | MIT off-target scoring (Hsu 2013), pure Go |
| pkg/scoring/cfd_score.go | 8.5K | CFD off-target scoring (Doench 2016), 240-entry matrix, pure Go |
| pkg/scoring/doench_client.go | 1.2K | Doench on-target scoring Go HTTP client |
| **Scoring Server (Python)** | | |
| scoring_server/server.py | 2.5K | Flask HTTP server wrapping Azimuth model |
| scoring_server/requirements.txt | 60B | Python dependencies |
| scoring_server/setup.sh | 400B | Setup script (venv + clone CrisprScoringHub) |
| **I/O & Parsing** | | |
| internal/parsers/parsers.go | 5.7K | ParseFastaFile, parseAndEncodeFastaFile, decodeFastaCacheFile, smart selector |
| **Utils** | | |
| internal/utils/utils.go | 964B | TimeFunction, PrintPattern, DebugPrintln/DebugRun, Prepend |
| internal/utils/file.go | 711B | GetCacheFileName, EnsureDir, DoesFileExists, IsFileModifiedAfterCaching |
| internal/bitmaps/nucleotide_map.go | 1.1K | NucleotideToBitMap, BitMapToNucleotide, NucleotideComplementMap |
| internal/stubs/stubs.go | 237B | FastaChanConsumerStub (test utility) |
| **Scripts** | | |
| scripts/reverse_complement_sequence.go | 523B | Standalone: reverse complement demo (build:ignore) |
| scripts/sequence_converter.go | 755B | Standalone: nucleotide → hex/binary demo (build:ignore) |
| **Config & Docs** | | |
| mprocs.yaml | 1.1K | Dev process runner (test, benchmarks, resource monitor) |
| .gitignore | 260B | Exclude large genomes, cache, outputs |
| go.mod | 34B | Module: vivalchemy/cris, Go 1.26.2 |
| Makefile | 111B | Template recipes (unused) |
| ARCHITECTURE.md | 7.5K | This file |
| **Test Data** | | |
| input.fna | 153B | Zebrafish chromosome 1 sample |
| output1.txt, output2.txt | 38K each | Example match results |
| **Genomic Data** (genomes/ncbi_dataset/data/) | | |
| GCA_022985175.1_*.fna | 1.08GB | Labeo rohita (rohu) GenBank assembly |
| GCF_022985175.1_*.fna | 1.08GB | Labeo rohita (rohu) RefSeq assembly |
| *.gtf | 365-428MB | Annotations (GCA/GCF) |
| sequence_report.jsonl | ~800KB | Metadata per assembly |
| data_summary.tsv | 610B | NCBI Dataset manifest |
| assembly_data_report.jsonl | 7.0K | NCBI assembly metadata |
| dataset_catalog.json | 1.3K | NCBI package metadata |

**Total source code**: ~50KB (Go) + ~3KB (Python scoring server)
**Total with genomes**: ~4.6GB (FASTA + GTF)
