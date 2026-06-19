package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/statikowsky/mar/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		if !errors.Is(err, cli.ErrHandled) {
			fmt.Fprintln(os.Stderr, "mar:", err)
		}
		os.Exit(1)
	}
}
