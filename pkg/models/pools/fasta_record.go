package pools

import "sync"

type FastaRecord struct {
	Header   string
	Sequence []byte
}

var FastaRecordPool = sync.Pool{
	New: func() any {
		return &FastaRecord{
			Sequence: make([]byte, 0, 1024*8),
		}
	},
}

// Release the FastaRecord back to the pool
func (fe *FastaRecord) Release() {
	fe.Header = ""
	fe.Sequence = fe.Sequence[:0]
	FastaRecordPool.Put(fe)
}

// Get a new FastaRecord from the pool
func NewFastaRecord() *FastaRecord {
	return FastaRecordPool.Get().(*FastaRecord)
}

type FastaRecordChunk struct {
	Header   string
	Sequence []byte
	Offset   int
}

var FastaRecordChunkPool = sync.Pool{
	New: func() any {
		return &FastaRecordChunk{
			Sequence: make([]byte, 0, 1024*8),
		}
	},
}

// Release the FastaRecord back to the pool
func (fe *FastaRecordChunk) Release() {
	fe.Header = ""
	fe.Sequence = fe.Sequence[:0]
	FastaRecordChunkPool.Put(fe)
}

// Get a new FastaRecord from the pool
func NewFastaRecordChunk() *FastaRecordChunk {
	return FastaRecordChunkPool.Get().(*FastaRecordChunk)
}
