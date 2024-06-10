package main

import (
	"bufio"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
)

func main() {
	in, err := argOrStdin()
	if err != nil {
		log.Fatal(err)
	}

	if err := run(in); err != nil {
		if !errors.Is(err, io.EOF) {
			log.Fatal(err)
		}
	}
}

func argOrStdin() (io.Reader, error) {
	if len(os.Args) < 2 {
		return os.Stdin, nil
	}

	arg := os.Args[1]
	if arg == "-" {
		return os.Stdin, nil
	}

	return os.Open(arg)
}

func run(in io.Reader) error {
	br := bufio.NewReader(in)
	w := io.Discard
	tee := &teeByteReader{r: br, w: w}

	var (
		zr  *gzip.Reader
		err error
	)

	for {
		start := tee.n

		if zr == nil {
			zr, err = gzip.NewReader(tee)
		} else {
			err = zr.Reset(tee)
		}
		if err != nil {
			return err
		}

		fmt.Printf("%d %s\n", start, zr.Header.Name)

		zr.Multistream(false)

		if _, err := io.Copy(io.Discard, zr); err != nil {
			return err
		}
	}
}

// like io.TeeReader but also implements io.ByteReader for gzip.
//
// From gzip.Reader.Multistream:
//
// > If the underlying reader implements io.ByteReader,
// > it will be left positioned just after the gzip stream.
type teeByteReader struct {
	n int64

	r interface {
		io.Reader
		io.ByteReader
	}

	w io.Writer
}

func (t *teeByteReader) ReadByte() (byte, error) {
	c, err := t.r.ReadByte()
	n, err := t.w.Write([]byte{c})
	t.n += int64(n)
	if err != nil {
		return c, err
	}
	return c, err
}

func (t *teeByteReader) Read(p []byte) (int, error) {
	n, err := t.r.Read(p)
	t.n += int64(n)
	if n > 0 {
		if n, err := t.w.Write(p[:n]); err != nil {
			return n, err
		}
	}
	return n, err
}
