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

const TempDirPattern = "test_temp_dir"

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
			err = Copy("testdata/input.txt", temp, test.offset, test.limit)
			require.NoError(t, err)

			// getResultFileAnsCompareWithExpected(t, temp, test.pathToExpectedFile)
		})
	}
}

func TestCopy2(t *testing.T) {
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
			file, err := os.CreateTemp(temp, TmpFileNamePattern)

			err = Copy("testdata/input.txt", file.Name(), test.offset, test.limit)
			require.NoError(t, err)

			getResultFileAnsCompareWithExpected(t, file, test.pathToExpectedFile)
		})
	}
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
			err = Copy("testdata/input.txt", temp, test.offset, test.limit)
			require.Truef(t, errors.Is(err, io.EOF), "actual error %q", err)

			// getResultFileAnsCompareWithExpected(t, temp, test.pathToExpectedFile)
		})
	}
}

func TestCopyWhenOffsetGreaterFileSize(t *testing.T) {
	err := Copy("testdata/input.txt", "", 1_000_000_000_000_000, 0)
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

func getResultFilePath(temp string) (string, error) {
	glob, err := filepath.Glob(temp + "/*")
	if err != nil {
		return "", err
	}

	return glob[0], nil
}
