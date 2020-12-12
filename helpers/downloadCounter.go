package helpers

import (
	"fmt"
	"strings"
)

type WriteCounter struct {
	Total int
}

func (wc *WriteCounter) Write(p []byte) (int, error) {
	n := len(p)
	wc.Total += n
	wc.PrintProgress()
	return n, nil
}
func (wc WriteCounter) PrintProgress() {
	fmt.Printf("\r%s", strings.Repeat(" ", 35))
	fmt.Printf("\rDownloading... %s complete", ByteHuman(wc.Total, 2))
}

func (wc *WriteCounter) GetTotal() int {
	return wc.Total
}
