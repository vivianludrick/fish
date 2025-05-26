package pools

import "sync"

type FastaRecord struct {
	Header   []byte
	Sequence []byte
}

var FastaRecordPool = sync.Pool{
	New: func() any {
		return &FastaRecord{
			Header:   make([]byte, 0, 120),
			Sequence: make([]byte, 0, 1024*8),
		}
	},
}

// Release the FastaRecord back to the pool
func (fe *FastaRecord) Release() {
	fe.Header = fe.Header[:0] // reset but keep capacity
	fe.Sequence = fe.Sequence[:0]
	FastaRecordPool.Put(fe)
}

// Get a new FastaRecord from the pool
func NewFastaRecord() *FastaRecord {
	return FastaRecordPool.Get().(*FastaRecord)
}
