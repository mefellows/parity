package parity

import (
	"os"
	"os/signal"

	"github.com/mefellows/parity/log"

	"github.com/mefellows/parity/config"
	"github.com/mefellows/plugo/plugo"
)

type Parity struct {
	config *config.Config
	Sync   []Sync
}

func (p *Parity) LoadPlugins() {
	log.Debug("loading plugins")
	var err error
	var confLoader *plugo.ConfigLoader
	c := &config.RootConfig{}
	if p.config.ConfigFile != "" {
		confLoader = &plugo.ConfigLoader{}
		err = confLoader.LoadFromFile(p.config.ConfigFile, &c)
		if err != nil {
			log.Fatalf("Unable to read configuration file: %s", err.Error())
		}
	} else {
		log.Fatalf("No configuration file provided. Please create a 'parity.yml' file.")
	}
	log.SetLevel(log.LogLevel(c.LogLevel))

	// Load all plugins
	pluginConfig := &PluginConfig{Ui: p.config.Ui}
	p.Sync = make([]Sync, len(c.Sync))
	plugins := plugo.LoadPluginsWithConfig(confLoader, c.Sync)

	for i, pl := range plugins {
		log.Debug("Loading plugin\t" + log.Colorize(log.YELLOW, c.Sync[i].Name))
		p.Sync[i] = pl.(Sync)
		p.Sync[i].Configure(pluginConfig)
	}
}

// MergeConfig merges ~/.parityrc with any ./parity.yml files
func (p *Parity) mergeConfig() {
	// https://github.com/imdario/mergo -> MergeWithOverride
}

func New(config *config.Config) *Parity {
	return &Parity{config: config}
}

func NewWithDefault() *Parity {
	c := &config.Config{}
	return &Parity{config: c}
}

// Run Parity - the main application entrypoint
func (p *Parity) Run() {
	log.Info("Running Parity")

	log.Debug("Loading plugins...")
	p.LoadPlugins()

	// Execute all plugins in parallel?
	for _, pl := range p.Sync {
		go pl.Sync()
	}

	// Interrupt handler
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, os.Kill)

	select {
	case <-sigChan:
		log.Debug("Received interrupt, shutting down.")

		// Cancel stufff?
	}
}
