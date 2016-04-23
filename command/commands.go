package command

import (
	"os"
	"runtime"

	"github.com/going/toolkit/log"
	"github.com/mefellows/parity/config"
	"github.com/mitchellh/cli"
)

var Commands map[string]cli.CommandFactory
var Ui cli.Ui

func init() {

	if runtime.GOOS == "darwin" {
		Ui = &cli.ColoredUi{
			Ui:          &cli.BasicUi{Writer: os.Stdout, Reader: os.Stdin, ErrorWriter: os.Stderr},
			OutputColor: cli.UiColorYellow,
			InfoColor:   cli.UiColorNone,
			ErrorColor:  cli.UiColorRed,
		}
	} else {
		log.Debug("Not running an OSX environment, removing colour from output")
		Ui = &cli.BasicUi{Writer: os.Stdout, Reader: os.Stdin, ErrorWriter: os.Stderr}
	}

	meta := config.Meta{
		Ui: Ui,
	}

	Commands = map[string]cli.CommandFactory{
		"build": func() (cli.Command, error) {
			return &BuildCommand{
				Meta: meta,
			}, nil
		},
		"cleanup": func() (cli.Command, error) {
			return &CleanupCommand{
				Meta: meta,
			}, nil
		},
		"init": func() (cli.Command, error) {
			return &InitCommand{
				Meta: meta,
			}, nil
		},
		"x": func() (cli.Command, error) {
			return &XCommand{
				Meta: meta,
			}, nil
		},
		"install": func() (cli.Command, error) {
			return &InstallCommand{
				Meta: meta,
			}, nil
		},
		"interactive": func() (cli.Command, error) {
			return &InteractiveCommand{
				Meta: meta,
			}, nil
		},
		"attach": func() (cli.Command, error) {
			return &AttachCommand{
				Meta: meta,
			}, nil
		},
		"run": func() (cli.Command, error) {
			return &RunCommand{
				Meta: meta,
			}, nil
		},
		"setup": func() (cli.Command, error) {
			return &SetupCommand{
				Meta: meta,
			}, nil
		},
		"version": func() (cli.Command, error) {
			return &VersionCommand{}, nil
		},
	}
}
