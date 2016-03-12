package command

import (
	"os"

	"github.com/mitchellh/cli"
	"github.com/mefellows/parity/config"
)

var Commands map[string]cli.CommandFactory
var Ui cli.Ui

func init() {

	Ui = &cli.ColoredUi{
		Ui:          &cli.BasicUi{Writer: os.Stdout, Reader: os.Stdin, ErrorWriter: os.Stderr},
		OutputColor: cli.UiColorYellow,
		InfoColor:   cli.UiColorNone,
		ErrorColor:  cli.UiColorRed,
	}

	meta := config.Meta{
		Ui: Ui,
	}

	Commands = map[string]cli.CommandFactory{
		"install": func() (cli.Command, error) {
			return &InstallCommand{
				Meta: meta,
			}, nil
		},
		"run": func() (cli.Command, error) {
			return &RunCommand{
				Meta: meta,
			}, nil
		},
		"version": func() (cli.Command, error) {
			return &VersionCommand{}, nil
		},
	}
}
