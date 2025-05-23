package cli

import (
	"context"
	"encoding/json"

	"fmt"
	"io"
	"net/mail"

	"github.com/alnah/fla/deps"
	"github.com/urfave/cli/v3"
)

type Flux struct {
	Reader    io.Reader
	Writer    io.Writer
	ErrWriter io.Writer
}

func NewFlaCmd(d deps.CLIDeps, f Flux) *cli.Command {
	cmd := &cli.Command{
		Name:        "fla",
		Usage:       "entry point for the fla-cli",
		UsageText:   "",
		Version:     "0.0.1",
		Description: "discover over commands from here",
		Commands: []*cli.Command{
			{
				Name:  "dump-config",
				Usage: "print the effective configuration as JSON",
				Action: func(ctx context.Context, c *cli.Command) error {
					b, _ := json.MarshalIndent(d.ConfigManager.Config, "", "  ")
					fmt.Println(string(b))
					return nil

				},
			},
		},
		EnableShellCompletion: true,
		Authors: []any{
			mail.Address{Name: "Alexis Nahan", Address: "alexis.nahan@gmail.com"},
		},
		Copyright: "(c) 2025 Alexis Nahan",
		Reader:    f.Reader,
		Writer:    f.Writer,
		ErrWriter: f.ErrWriter,
		Suggest:   true,
	}
	return cmd
}
