package parity

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"

	"github.com/mefellows/parity/config"
	"github.com/mefellows/parity/log"
	"github.com/mefellows/plugo/plugo"
)

var banner = `
██████╗  █████╗ ██████╗ ██╗████████╗██╗   ██╗
██╔══██╗██╔══██╗██╔══██╗██║╚══██╔══╝╚██╗ ██╔╝
██████╔╝███████║██████╔╝██║   ██║    ╚████╔╝
██╔═══╝ ██╔══██║██╔══██╗██║   ██║     ╚██╔╝
██║     ██║  ██║██║  ██║██║   ██║      ██║
╚═╝     ╚═╝  ╚═╝╚═╝  ╚═╝╚═╝   ╚═╝      ╚═╝


`

// Parity contains the top level configuration for Parity (plugins etc.)
type Parity struct {
	config       *config.Config
	SyncPlugins  []Sync
	RunPlugins   []Run
	BuildPlugins []Builder
	ShellPlugins []Shell
	pluginConfig *PluginConfig
	errorChan    chan error
	plugins      []Plugin
}

// LoadPlugins loads all plugins referenced in the parity.yml file
// from those registered at runtime
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
	p.pluginConfig = &PluginConfig{Ui: p.config.Ui}

	// Set project name
	p.pluginConfig.ProjectName = c.Name
	p.pluginConfig.ProjectNameSafe = strings.Replace(strings.ToLower(c.Name), " ", "", -1)

	// Sync plugins
	p.SyncPlugins = make([]Sync, len(c.Sync))
	syncPlugins := plugo.LoadPluginsWithConfig(confLoader, c.Sync)

	for i, pl := range syncPlugins {
		log.Debug("Loading Sync Plugin\t" + log.Colorize(log.YELLOW, c.Sync[i].Name))
		p.SyncPlugins[i] = pl.(Sync)
		p.SyncPlugins[i].Configure(p.pluginConfig)
		p.plugins = append(p.plugins, p.SyncPlugins[i])
	}

	// Run plugins
	p.RunPlugins = make([]Run, len(c.Run))
	runPlugins := plugo.LoadPluginsWithConfig(confLoader, c.Run)

	for i, pl := range runPlugins {
		log.Debug("Loading Run Plugin\t" + log.Colorize(log.YELLOW, c.Run[i].Name))
		p.RunPlugins[i] = pl.(Run)
		p.RunPlugins[i].Configure(p.pluginConfig)
		p.plugins = append(p.plugins, p.RunPlugins[i])
	}

	// Build plugins
	p.BuildPlugins = make([]Builder, len(c.Build))
	buildPlugins := plugo.LoadPluginsWithConfig(confLoader, c.Build)

	for i, pl := range buildPlugins {
		log.Debug("Loading Build Plugin\t" + log.Colorize(log.YELLOW, c.Build[i].Name))
		p.BuildPlugins[i] = pl.(Builder)
		p.BuildPlugins[i].Configure(p.pluginConfig)
		p.plugins = append(p.plugins, p.BuildPlugins[i])
	}

	// Shell plugins
	p.ShellPlugins = make([]Shell, len(c.Shell))
	shellPlugins := plugo.LoadPluginsWithConfig(confLoader, c.Shell)

	for i, pl := range shellPlugins {
		log.Debug("Loading Shell Plugin\t" + log.Colorize(log.YELLOW, c.Shell[i].Name))
		p.ShellPlugins[i] = pl.(Shell)
		p.ShellPlugins[i].Configure(p.pluginConfig)
		p.plugins = append(p.plugins, p.ShellPlugins[i])
	}
}

// GetPlugin gets a plugin by name (no type)
func (p *Parity) GetPlugin(name string) (pl interface{}, err error) {
	for _, pl := range p.plugins {
		if pl.Name() == name {
			return pl, nil
		}
	}
	return nil, fmt.Errorf("Plugin '%s' not found", name)
}

// GetShellPlugin gets a plugin by name and converts to a Shell
func (p *Parity) GetShellPlugin(plugin string) (Shell, error) {
	if pl, err := p.GetPlugin(plugin); err == nil {
		return pl.(Shell), nil
	}
	return nil, nil
}

// MergeConfig merges ~/.parityrc with any ./parity.yml files
func (p *Parity) mergeConfig() {
	// https://github.com/imdario/mergo -> MergeWithOverride
}

// New creates a default instance of Parity, using the provided config
func New(config *config.Config) *Parity {
	return &Parity{config: config}
}

// NewWithDefault creates a new instance of Parity with default settings
func NewWithDefault() *Parity {
	c := &config.Config{}
	return &Parity{config: c}
}

// Build runs all builders on the project, e.g. Docker build
func (p *Parity) Build() error {
	log.Debug("Loading plugins...")
	p.LoadPlugins()

	for _, pl := range p.BuildPlugins {
		if err := pl.Build(); err != nil {
			return err
		}
	}
	return nil
}

// Run Parity - the main application entrypoint
func (p *Parity) Run() {
	log.Banner(banner)

	log.Debug("Loading plugins...")
	p.LoadPlugins()

	// Execute all plugins in parallel?
	// TODO: Initial Sync may need to be blocking so that Run
	//       can work?
	for _, pl := range p.SyncPlugins {
		p.runAsync(pl.Sync)
	}

	// Run all Runners
	for _, pl := range p.RunPlugins {
		p.runAsync(pl.Run)
	}

	// Interrupt handler
	sigChan := make(chan os.Signal, 1)
	p.errorChan = make(chan error)
	signal.Notify(sigChan, os.Interrupt, os.Kill)

	select {
	case e := <-p.errorChan:
		log.Error(e.Error())
	case <-sigChan:
		log.Debug("Received interrupt, shutting down.")
		p.Teardown()
	}
}

func (p *Parity) runAsync(f func() error) {
	go func() {
		if err := f(); err != nil {
			p.errorChan <- err
		}
	}()
}

// Teardown safely shuts down all registered plugins
func (p *Parity) Teardown() {
	group := &sync.WaitGroup{}

	for _, pl := range p.SyncPlugins {
		p.runGroupAsync(group, pl.Teardown)
	}

	for _, pl := range p.RunPlugins {
		p.runGroupAsync(group, pl.Teardown)
	}
	group.Wait()
}

func (p *Parity) runGroupAsync(group *sync.WaitGroup, f func() error) {
	group.Add(1)
	go func() {
		if err := f(); err != nil {
			p.errorChan <- err
		}
		group.Done()
	}()
}
