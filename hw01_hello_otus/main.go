package main

import (
	"fmt"

	"golang.org/x/example/hello/reverse"
)

func main() {
	text := "Hello, OTUS!"
	fmt.Println(reverse.String(text))
}
