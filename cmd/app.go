// Package cmd provides the pcopy CLI application
package cmd

import (
	"github.com/urfave/cli/v2"
	"os"
)

// New creates a new CLI application
func New() *cli.App {
	return &cli.App{
		Name:                   "submonkey",
		Usage:                  "create videos from your favorite Reddit subs",
		UsageText:              "submonkey COMMAND [OPTION..] [ARG..]",
		HideHelp:               true,
		HideVersion:            true,
		UseShortOptionHandling: true,
		Reader:                 os.Stdin,
		Writer:                 os.Stdout,
		ErrWriter:              os.Stderr,
		Commands: []*cli.Command{
			cmdCreate,
		},
	}
}
