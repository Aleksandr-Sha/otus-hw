package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/cheggaaa/pb/v3"
)

var (
	ErrUnsupportedFile       = errors.New("unsupported file")
	ErrOffsetExceedsFileSize = errors.New("offset exceeds file size")
)

func Copy(fromPath, toPath string, offset, limit int64) error {
	fileForCopyFrom, stat, err := preparingFileForCopyFrom(fromPath, offset)
	if err != nil {
		return fmt.Errorf("prepeare file for copy from: %w", err)
	}

	tempFileForCopyTo, err := preparingTempFileForCopyTo(toPath)
	if err != nil {
		return fmt.Errorf("prepeare temp file for copy to : %w", err)
	}

	bar, proxyReader := getReaderWithProgressBar(fileForCopyFrom, stat, offset, limit)

	err = copyData(proxyReader, tempFileForCopyTo, limit)
	if err != nil {
		return fmt.Errorf("copy data : %w", err)
	}

	bar.Finish()
	// Писать будем во временный файл, важно не забыть удалить его через Remove

	return nil
}

func copyData(reader io.Reader, writer io.Writer, limit int64) error {
	if limit == 0 {
		_, err := io.Copy(writer, reader)
		if err != nil {
			return fmt.Errorf("full copy : %w", err)
		}
	} else {
		_, err := io.CopyN(writer, reader, limit)
		if err != nil {
			return fmt.Errorf("limit copy : %w", err)
		}
	}

	return nil
}

func getReaderWithProgressBar(fileForCopyFrom *os.File, stat os.FileInfo, offset, limit int64) (*pb.ProgressBar, io.Reader) {
	bar := pb.Full.Start64(getBarSize(stat, offset, limit))
	return bar, bar.NewProxyReader(fileForCopyFrom)
}

func preparingTempFileForCopyTo(toPath string) (*os.File, error) {
	abs, err := filepath.Abs(toPath)
	if err != nil {
		return nil, fmt.Errorf("get file path: %w", err)
	}

	temp, err := os.CreateTemp(abs, "tmp_for_copy")
	if err != nil {
		return nil, fmt.Errorf("create file: %w", err)
	}

	return temp, nil
}

func preparingFileForCopyFrom(fromPath string, offset int64) (*os.File, os.FileInfo, error) {
	abs, err := filepath.Abs(fromPath)
	if err != nil {
		return nil, nil, fmt.Errorf("get file path: %w", err)
	}

	open, err := os.Open(abs)
	if err != nil {
		return nil, nil, fmt.Errorf("open file: %w", err)
	}

	stat, err := open.Stat()
	if err != nil {
		return nil, nil, fmt.Errorf("get file stat: %w", err)
	}

	if offset > 0 {
		if stat.Size() < offset {
			return nil, nil, ErrOffsetExceedsFileSize
		}

		_, err := open.Seek(offset, io.SeekStart)
		if err != nil {
			return nil, nil, fmt.Errorf("seek by offset : %w", err)
		}
	}

	return open, stat, nil
}

func getBarSize(stat os.FileInfo, offset, limit int64) int64 {
	sizeReadablePart := stat.Size() - offset

	if limit == 0 || sizeReadablePart == limit || sizeReadablePart < limit {
		return sizeReadablePart
	}

	return limit
}
