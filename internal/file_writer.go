package internal

import (
	"fmt"
	"os"
)

type FileWriterInterface interface {
	Open(filepath string) error
	Write(b []byte) (n int, err error)
	Close() error
}

type FileWriter struct {
	f *os.File
}

func (fw *FileWriter) Open(filepath string) error {
	var err error
	fw.f, err = os.OpenFile(filepath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755) //nolint:gofumpt
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	err = fw.f.Truncate(0)
	if err != nil {
		return fmt.Errorf("failed to truncate coverage.html file whilst overwriting it: %w", err)
	}
	_, err = fw.f.Seek(0, 0)
	if err != nil {
		return fmt.Errorf("failed to find beginning of coverage.html file: %w", err)
	}

	return nil
}

func (fw *FileWriter) Write(b []byte) (int, error) {
	return fw.f.Write(b)
}

func (fw *FileWriter) Close() error {
	return fw.f.Close()
}
