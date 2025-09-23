package util

import (
	"bytes"
	"encoding/ascii85"
	"io"

	"github.com/klauspost/compress/gzip"
)

func CompressGzipString(s string) ([]byte, error) {
	buf := bytes.NewBuffer(make([]byte, 0, len(s)/2))
	g, err := gzip.NewWriterLevel(buf, gzip.BestCompression)
	if err != nil {
		return nil, err
	}
	_, err = g.Write([]byte(s))
	if err != nil {
		return nil, err
	}
	err = g.Close()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func DecompressGzipString(b []byte) (string, error) {
	g, err := gzip.NewReader(bytes.NewReader(b))
	if err != nil {
		return "", err
	}
	defer g.Close()

	buf := bytes.NewBuffer(make([]byte, 0, len(b)*2))
	_, err = io.Copy(buf, g)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

func DecompressGzipAscii85(s string) ([]byte, error) {
	dec := ascii85.NewDecoder(bytes.NewReader([]byte(s)))
	r, err := gzip.NewReader(dec)
	if err != nil {
		return nil, err
	}
	return io.ReadAll(r)
}
