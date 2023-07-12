package internal_test

import (
	"bytes"
	"io"
)

type exiterMock struct {
	code int
}

func (m *exiterMock) Exit(code int) {
	m.code = code
}

type fileWriterMock struct {
	f io.Writer
}

func (fw *fileWriterMock) Open(filepath string) error {
	fw.f = &bytes.Buffer{}
	return nil
}

func (fw *fileWriterMock) Write(b []byte) (int, error) {
	return fw.f.Write(b)
}

func (fw *fileWriterMock) Close() error {
	return nil
}
