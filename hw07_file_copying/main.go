package main

import (
	"errors"
	"fmt"
	"log"

	"github.com/spf13/pflag"
)

type CopySetting struct {
	from   string
	to     string
	offset int64
	limit  int64
}

func (c *CopySetting) checkPaths() error {
	if c.from == "" || c.to == "" {
		return errors.New("path from or to is empty")
	}

	return nil
}

func main() {
	settings, err := parseCopySettings()
	if err != nil {
		log.Fatal("Parse settings failed: ", err)
	}

	err = Copy(settings.from, settings.to, settings.offset, settings.limit)
	if err != nil {
		log.Fatal("Copy failed: ", err)
	}
}

func parseCopySettings() (*CopySetting, error) {
	setting := &CopySetting{}

	pflag.StringVarP(&setting.from, "from", "f", "", "file to read from")
	pflag.StringVarP(&setting.to, "to", "t", "", "file to write to")
	pflag.Int64VarP(&setting.offset, "offset", "o", 0, "limit of bytes to copy")
	pflag.Int64VarP(&setting.limit, "limit", "l", 0, "offset in input file")

	pflag.Parse()

	if err := setting.checkPaths(); err != nil {
		return nil, fmt.Errorf("parse setting: %w", err)
	}

	return setting, nil
}
