package config

import (
	"fmt"
	"regexp"

	"github.com/mefellows/parity/log"

	"github.com/mefellows/plugo/plugo"
	"github.com/mitchellh/cli"
)

type Config struct {
	RawConfig  *plugo.RawConfig
	ConfigFile string
	Ui         cli.Ui
}

type Excludes []regexp.Regexp

func (e *Excludes) String() string {
	return fmt.Sprintf("%s", *e)
}

func (e *Excludes) Set(value string) error {
	r, err := regexp.CompilePOSIX(value)
	if err == nil {
		*e = append(*e, *r)
	} else {
		log.Error("Error:", err.Error())
	}

	return nil
}

type RootConfig struct {
	Name        string
	Description string
	LogLevel    int                  `default:"2" required:"true" mapstructure:"loglevel"`
	Run         []plugo.PluginConfig `mapstructure:"run"`
	Sync        []plugo.PluginConfig `mapstructure:"sync"`
	Shell       []plugo.PluginConfig `mapstructure:"shell"`
}
