package main

import (
	"fmt"
	"os"

	"github.com/ExquisiteCore/cnki-search/internal/cli"
)

var version = "dev"

func main() {
	if err := cli.NewRoot(version).Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(exitCodeFor(err))
	}
}

func exitCodeFor(err error) int {
	if coded, ok := err.(cli.CodedError); ok {
		return coded.ExitCode()
	}
	return 1
}
