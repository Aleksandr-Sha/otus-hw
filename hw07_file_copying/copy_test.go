package main

import (
	"errors"
	"io"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCopy(t *testing.T) {
	temp, err := os.MkdirTemp(".", "my_temp_dir")
	require.NoError(t, err)
	defer os.RemoveAll(temp)

	tests := []struct {
		offset             int64
		limit              int64
		pathToExpectedFile string
	}{
		{offset: 0, limit: 0, pathToExpectedFile: "testdata/out_offset0_limit0.txt"},
		{offset: 0, limit: 10, pathToExpectedFile: "testdata/out_offset0_limit10.txt"},
		{offset: 0, limit: 1000, pathToExpectedFile: "testdata/out_offset0_limit1000.txt"},
		{offset: 100, limit: 1000, pathToExpectedFile: "testdata/out_offset100_limit1000.txt"},
	}

	for _, test := range tests {
		t.Run(test.pathToExpectedFile, func(t *testing.T) {
			err = Copy("testdata/input.txt", temp, test.offset, test.limit)

			pathToResult, err := getResultFilePath(temp)
			require.NoError(t, err)

			resultFile, err := os.Open(pathToResult)
			require.NoError(t, err)

			defer closeFile(resultFile)
			defer removeTempFile(pathToResult)

			resultBuffer := getBytesFromFile(resultFile, t)

			expectedFile, err := os.Open(test.pathToExpectedFile)
			require.NoError(t, err)
			defer closeFile(expectedFile)

			expectedBuffer := getBytesFromFile(expectedFile, t)

			require.Equal(t, expectedBuffer, resultBuffer)
		})
	}
}

func TestCopyWhenEOF(t *testing.T) {
	temp, err := os.MkdirTemp(".", "my_temp_dir")
	require.NoError(t, err)
	defer os.RemoveAll(temp)

	tests := []struct {
		offset             int64
		limit              int64
		pathToExpectedFile string
	}{
		{offset: 0, limit: 10000, pathToExpectedFile: "testdata/out_offset0_limit10000.txt"},
		{offset: 6000, limit: 1000, pathToExpectedFile: "testdata/out_offset6000_limit1000.txt"},
	}

	for _, test := range tests {
		t.Run(test.pathToExpectedFile, func(t *testing.T) {
			err = Copy("testdata/input.txt", temp, test.offset, test.limit)
			require.Truef(t, errors.Is(err, io.EOF), "actual error %q", err)

			pathToResult, err := getResultFilePath(temp)
			require.NoError(t, err)

			resultFile, err := os.Open(pathToResult)
			require.NoError(t, err)

			defer closeFile(resultFile)
			defer removeTempFile(pathToResult)

			resultBuffer := getBytesFromFile(resultFile, t)

			expectedFile, err := os.Open(test.pathToExpectedFile)
			require.NoError(t, err)
			defer closeFile(expectedFile)

			expectedBuffer := getBytesFromFile(expectedFile, t)

			require.Equal(t, expectedBuffer, resultBuffer)
		})
	}
}

func TestCopyWhenOffsetGreaterFileSize(t *testing.T) {
	err := Copy("testdata/input.txt", "", 1_000_000_000_000_000, 0)
	require.Truef(t, errors.Is(err, ErrOffsetExceedsFileSize), "actual error %q", err)
}

func removeTempFile(name string) {
	err := os.Remove(name)
	if err != nil {
		log.Printf("Error removing temp file with result: %v", err)
	}
}

func closeFile(expectedFile *os.File) {
	err := expectedFile.Close()
	if err != nil {
		log.Printf("Error closing file: %v", err)
	}
}

func getBytesFromFile(file *os.File, t *testing.T) []byte {
	info, err := file.Stat()
	require.NoError(t, err)

	resultBuffer := make([]byte, info.Size())

	n, err := file.Read(resultBuffer)
	require.NoError(t, err)

	require.Equal(t, info.Size(), int64(n), "Incorrect number of bytes read")

	return resultBuffer
}

func getResultFilePath(temp string) (string, error) {
	glob, err := filepath.Glob(temp + "/*")
	if err != nil {
		return "", err
	}

	return glob[0], nil
}
