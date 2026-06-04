package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/cheggaaa/pb/v3"
)

const TmpFileNamePattern = "tmp_for_copy"

var ErrOffsetExceedsFileSize = errors.New("offset exceeds file size")

func Copy(fromPath, toPath string, offset, limit int64) error {
	fileForCopyFrom, fileForCopyInfo, err := preparingFileForCopyFrom(fromPath, offset)
	if err != nil {
		return fmt.Errorf("prepare file for copy from: %w", err)
	}
	defer closeFileWithErrorHandle(fileForCopyFrom)

	tempFileForCopyTo, err := preparingTempFileForCopyTo(toPath)
	if err != nil {
		return fmt.Errorf("prepare temp file for copy to : %w", err)
	}
	defer closeFileWithErrorHandle(tempFileForCopyTo)

	bar, proxyReader := getReaderProxyWithProgressBar(fileForCopyFrom, fileForCopyInfo, offset, limit)
	defer bar.Finish()

	err = copyData(proxyReader, tempFileForCopyTo, limit)
	if err != nil {
		return fmt.Errorf("copy data : %w", err)
	}

	return nil
}

func preparingFileForCopyFrom(fromPath string, offset int64) (*os.File, os.FileInfo, error) {
	absPath, err := filepath.Abs(fromPath)
	if err != nil {
		return nil, nil, fmt.Errorf("get file path: %w", err)
	}

	file, err := os.Open(absPath)
	if err != nil {
		return nil, nil, fmt.Errorf("file file: %w", err)
	}

	fileInfo, err := file.Stat()
	if err != nil {
		return nil, nil, fmt.Errorf("get file fileInfo: %w", err)
	}

	if offset > 0 {
		if fileInfo.Size() < offset {
			closeFileWithErrorHandle(file)
			return nil, nil, ErrOffsetExceedsFileSize
		}

		_, err := file.Seek(offset, io.SeekStart)
		if err != nil {
			closeFileWithErrorHandle(file)
			return nil, nil, fmt.Errorf("seek by offset : %w", err)
		}
	}

	return file, fileInfo, nil
}

func preparingTempFileForCopyTo(toPath string) (*os.File, error) {
	absPath, err := filepath.Abs(toPath)
	if err != nil {
		return nil, fmt.Errorf("get file path: %w", err)
	}

	file, err := os.CreateTemp(absPath, TmpFileNamePattern)
	if err != nil {
		return nil, fmt.Errorf("create file: %w", err)
	}

	return file, nil
}

func getReaderProxyWithProgressBar(fileForCopyFrom *os.File, stat os.FileInfo, offset, limit int64) (*pb.ProgressBar, io.Reader) {
	bar := pb.Full.Start64(getBarSize(stat, offset, limit))
	return bar, bar.NewProxyReader(fileForCopyFrom)
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

func closeFileWithErrorHandle(file *os.File) {
	err := file.Close()
	if err != nil {
		log.Printf("close file for copy from: %s", err.Error())
	}
}

func getBarSize(stat os.FileInfo, offset, limit int64) int64 {
	sizeReadablePart := stat.Size() - offset

	if limit == 0 || sizeReadablePart == limit || sizeReadablePart < limit {
		return sizeReadablePart
	}

	return limit
}
