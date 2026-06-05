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

const (
	TempDirPattern     = "test_temp_dir"
	TmpFileNamePattern = "tmp_for_copy*.txt"
	InputFilePath      = "testdata/input.txt"
)

func TestCopy(t *testing.T) {
	temp, err := os.MkdirTemp(".", TempDirPattern)
	require.NoError(t, err)
	defer removeTempDir(temp)

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
			tempFileForResult, err := os.CreateTemp(temp, TmpFileNamePattern)
			require.NoError(t, err)

			err = Copy(InputFilePath, tempFileForResult.Name(), test.offset, test.limit)
			require.NoError(t, err)

			getResultFileAnsCompareWithExpected(t, tempFileForResult, test.pathToExpectedFile)
		})
	}
}

func TestCopyWhenFileNotExist(t *testing.T) {
	temp, err := os.MkdirTemp(".", TempDirPattern)
	require.NoError(t, err)
	defer removeTempDir(temp)

	pathForNewFile := filepath.Join(temp, "new_file.txt")

	err = Copy(InputFilePath, pathForNewFile, 0, 0)
	require.NoError(t, err)

	open, err := os.Open(pathForNewFile)
	require.NoError(t, err)

	getResultFileAnsCompareWithExpected(t, open, "testdata/out_offset0_limit0.txt")
}

func TestCopyWhenEOF(t *testing.T) {
	temp, err := os.MkdirTemp(".", TempDirPattern)
	require.NoError(t, err)
	defer removeTempDir(temp)

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
			tempFileForResult, err := os.CreateTemp(temp, TmpFileNamePattern)
			require.NoError(t, err)

			err = Copy(InputFilePath, tempFileForResult.Name(), test.offset, test.limit)
			require.Truef(t, errors.Is(err, io.EOF), "actual error %q", err)

			getResultFileAnsCompareWithExpected(t, tempFileForResult, test.pathToExpectedFile)
		})
	}
}

func TestCopyWhenOffsetGreaterFileSize(t *testing.T) {
	err := Copy(InputFilePath, "", 1_000_000_000_000_000, 0)
	require.Truef(t, errors.Is(err, ErrOffsetExceedsFileSize), "actual error %q", err)
}

func TestCopyWhenNotExistentFile(t *testing.T) {
	err := Copy("testdata/not_existent.txt", "", 0, 0)
	require.Truef(t, errors.Is(err, os.ErrNotExist), "actual error %q", err)
}

func getResultFileAnsCompareWithExpected(t *testing.T, resultFile *os.File, pathToExpectedFile string) {
	t.Helper()

	defer func() {
		closeFile(resultFile)
	}()

	resultBuffer := getBytesFromFile(t, resultFile)

	expectedFile, err := os.Open(pathToExpectedFile)
	require.NoError(t, err)
	defer closeFile(expectedFile)

	expectedBuffer := getBytesFromFile(t, expectedFile)

	require.Equal(t, expectedBuffer, resultBuffer)
}

func removeTempDir(name string) {
	err := os.RemoveAll(name)
	if err != nil {
		log.Printf("Error removing temp dir %q: %v", name, err)
	}
}

func closeFile(expectedFile *os.File) {
	err := expectedFile.Close()
	if err != nil {
		log.Printf("Error closing file: %v", err)
	}
}

func getBytesFromFile(t *testing.T, file *os.File) []byte {
	t.Helper()

	info, err := file.Stat()
	require.NoError(t, err)

	resultBuffer := make([]byte, info.Size())

	n, err := file.Read(resultBuffer)
	require.NoError(t, err)

	require.Equal(t, info.Size(), int64(n), "Incorrect number of bytes read")

	return resultBuffer
}
