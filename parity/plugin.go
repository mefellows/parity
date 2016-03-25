package parity

import "github.com/mitchellh/cli"

type PluginConfig struct {
	Ui              cli.Ui
	ProjectName     string
	ProjectNameSafe string
}

type Plugin interface {
	Configure(*PluginConfig)
	Teardown() error
	Name() string
}
