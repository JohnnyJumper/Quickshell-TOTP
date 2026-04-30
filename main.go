package main

import (
	"fmt"
	"os"

	"totp/internal/cli"
)

func main() {

	if err := cli.Run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "%s", err.Error())
		os.Exit(1)
	}
}
