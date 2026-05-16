package main

import "testing"

func TestCopy(t *testing.T) {
	err := Copy("testdata/input.txt", "test_res", 0, 3000)
	if err != nil {
		return
	}

	// Place your code here.
}
