package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/statikowsky/mar/internal/homebrew"
)

func main() {
	var tag string
	var checksumsPath string
	var formulaPath string
	flag.StringVar(&tag, "tag", "", "release tag, e.g. v0.2.0")
	flag.StringVar(&checksumsPath, "checksums", "", "path to release checksums.txt")
	flag.StringVar(&formulaPath, "formula", "", "path to Formula/mar.rb in the tap checkout")
	flag.Parse()

	if err := run(tag, checksumsPath, formulaPath); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(tag, checksumsPath, formulaPath string) error {
	if checksumsPath == "" {
		return fmt.Errorf("checksums path is required")
	}
	f, err := os.Open(checksumsPath)
	if err != nil {
		return fmt.Errorf("open checksums: %w", err)
	}
	defer f.Close()

	formula, err := homebrew.RenderFormula(tag, f)
	if err != nil {
		return err
	}
	if err := homebrew.WriteFormula(formulaPath, formula); err != nil {
		return err
	}
	return nil
}
