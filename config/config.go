package config

import "github.com/mefellows/plugo/plugo"

type Config struct {
	RawConfig  *plugo.RawConfig
	ConfigFile string
}

type PluginConfig struct {
	Name        string
	Description string
	LogLevel    int                  `default:"2" required:"true" mapstructure:"loglevel"`
	Deployment  []plugo.PluginConfig `mapstructure:"deployment"`
}
