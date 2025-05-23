package main

import (
	"context"
	"fmt"
	"os"

	"github.com/alnah/fla/cli"
	"github.com/alnah/fla/deps"
)

func main() {
	dependencies, err := deps.New(os.Stdin, os.Stdout, os.Stderr)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	cliFla := cli.NewFlaCmd(*dependencies, cli.Flux{
		Reader:    os.Stdin,
		Writer:    os.Stdout,
		ErrWriter: os.Stderr,
	})

	if err := cliFla.Run(context.Background(), os.Args); err != nil {
		dependencies.Logger.Error("fla running", "error", err.Error())
		return
	}
}
