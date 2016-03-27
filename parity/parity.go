package parity

import (
	"errors"
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

type Parity struct {
	config       *config.Config
	SyncPlugins  []Sync
	RunPlugins   []Run
	ShellPlugins []Shell
	pluginConfig *PluginConfig
	errorChan    chan error
	plugins      []Plugin
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

func (p *Parity) GetPlugin(name string) (pl interface{}, err error) {
	for _, pl := range p.plugins {
		if pl.Name() == name {
			return pl, nil
		}
	}
	return nil, errors.New(fmt.Sprintf("Plugin '%s' not found", name))
}

func (p *Parity) GetShellPlugin(plugin string) (Shell, error) {
	if pl, err := p.GetPlugin(plugin); err == nil {
		return pl.(Shell), nil
	} else {
		return nil, err
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
	log.Banner(banner)

	log.Debug("Loading plugins...")
	p.LoadPlugins()

	// Execute all plugins in parallel?
	// TODO: Initial Sync may need to be blocking so that Run
	//       can work?
	for _, pl := range p.SyncPlugins {
		p.runAsync(pl.Sync)
	}

	//
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
