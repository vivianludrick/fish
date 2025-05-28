package stubs

import (
	"fmt"
	"vivalchemy/cris/pkg/models/pools"
)

func FastaChanConsumerStub(fastaChan <-chan *pools.FastaRecord) {
	for fastaRecord := range fastaChan {
		fmt.Println(fastaRecord.Header)
		fastaRecord.Release()
	}
}
