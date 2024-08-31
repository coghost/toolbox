package epub

import (
	"bytes"
	"io"

	"github.com/ungerik/go-dry"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

func MustGbkToUtf8(s []byte) []byte {
	b, err := GbkToUtf8(s)
	dry.PanicIfErr(err)
	return b
}

func GbkToUtf8(src []byte) ([]byte, error) {
	reader := transform.NewReader(bytes.NewReader(src), simplifiedchinese.GBK.NewDecoder())
	return io.ReadAll(reader)
}
