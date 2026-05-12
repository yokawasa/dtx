package main

import (
	"fmt"
	"os"

	"github.com/yokawasa/dtx/internal/apperr"
	"github.com/yokawasa/dtx/internal/cli"
)

func main() {
	err := cli.Run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr)
	if err == nil {
		return
	}

	if !apperr.IsSilent(err) {
		fmt.Fprintf(os.Stderr, "dtx: %v\n", err)
	}
	os.Exit(apperr.ExitCode(err))
}
