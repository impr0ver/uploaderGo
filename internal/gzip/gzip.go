package gzip

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
)

func CompressJSON(w io.Writer, i interface{}) error {
	gz := gzip.NewWriter(w)
	if err := json.NewEncoder(gz).Encode(i); err != nil {
		return err
	}
	return gz.Close()
}

func CompressMultipart(w io.Writer, b *bytes.Buffer) error {
	gzipWriter := gzip.NewWriter(w)
	if _, err := gzipWriter.Write(b.Bytes()); err != nil {
		return err
	}
	return gzipWriter.Close()

}
