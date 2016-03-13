package parity

import (
	"os"
	"os/signal"

	"strings"
	"sync"

	"github.com/mefellows/parity/log"

	"github.com/mefellows/parity/config"
	"github.com/mefellows/plugo/plugo"
)

type Parity struct {
	config      *config.Config
	SyncPlugins []Sync
	RunPlugins  []Run
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

	// Set project name
	pluginConfig.ProjectName = c.Name
	pluginConfig.ProjectNameSafe = strings.Replace(strings.ToLower(c.Name), " ", "", -1)

	// Sync plugins
	p.SyncPlugins = make([]Sync, len(c.Sync))
	syncPlugins := plugo.LoadPluginsWithConfig(confLoader, c.Sync)

	for i, pl := range syncPlugins {
		log.Debug("Loading Sync Plugin\t" + log.Colorize(log.YELLOW, c.Sync[i].Name))
		p.SyncPlugins[i] = pl.(Sync)
		p.SyncPlugins[i].Configure(pluginConfig)
	}

	// Run plugins
	p.RunPlugins = make([]Run, len(c.Run))
	runPlugins := plugo.LoadPluginsWithConfig(confLoader, c.Run)

	for i, pl := range runPlugins {
		log.Debug("Loading Run Plugin\t" + log.Colorize(log.YELLOW, c.Run[i].Name))
		p.RunPlugins[i] = pl.(Run)
		p.RunPlugins[i].Configure(pluginConfig)
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
	for _, pl := range p.SyncPlugins {
		go pl.Sync()
	}

	// Execute all plugins in parallel?
	for _, pl := range p.RunPlugins {
		go pl.Run()
	}

	// Interrupt handler
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, os.Kill)

	select {
	case <-sigChan:
		log.Debug("Received interrupt, shutting down.")
		p.Teardown()
	}
}

func (p *Parity) Teardown() {
	group := &sync.WaitGroup{}

	for _, pl := range p.SyncPlugins {
		runAsync(group, pl.Teardown)
	}

	for _, pl := range p.RunPlugins {
		runAsync(group, pl.Teardown)
	}
	group.Wait()
}

func runAsync(group *sync.WaitGroup, f func()) {
	group.Add(1)
	go func() {
		f()
		group.Done()
	}()
}
