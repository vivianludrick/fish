package stubs

import (
	"fmt"
	"vivalchemy/cris/pkg/models"
)

func FastaChanConsumerStub(fastaChan <-chan *models.FastaRecord) {
	for fastaRecord := range fastaChan {
		fmt.Println(fastaRecord.Header)
		fastaRecord.Release()
	}
}
