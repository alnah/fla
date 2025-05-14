package cli

import (
	"net/mail"

	"github.com/urfave/cli/v3"
)

func NewFlaCmd() *cli.Command {
	cmd := &cli.Command{
		Name:                  "fla",
		Usage:                 "entry point for the fla-cli",
		UsageText:             "",
		Version:               "0.0.1",
		Description:           "discover over commands from here",
		DefaultCommand:        "help", // TODO: validate
		Commands:              []*cli.Command{},
		Flags:                 []cli.Flag{},
		EnableShellCompletion: true,
		Authors: []any{
			mail.Address{Name: "Alexis Nahan", Address: "alexis.nahan@gmail.com"},
		},
		Copyright:      "(c) 2025 Alexis Nahan",
		Reader:         nil, // TODO: for testing
		Writer:         nil,
		ErrWriter:      nil,
		ExitErrHandler: nil,
		ExtraInfo: func() map[string]string {
			panic("TODO")
		},
		Suggest: true,
	}
	return cmd
}
