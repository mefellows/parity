package parity

import "github.com/mitchellh/cli"

type PluginConfig struct {
	Ui cli.Ui
}

type Plugin interface {
	Configure(*PluginConfig)
	Teardown()
}
