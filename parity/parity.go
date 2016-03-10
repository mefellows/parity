package parity

import (
	"github.com/mefellows/muxy/log"
	"github.com/mefellows/parity/config"
	"github.com/mefellows/plugo/plugo"
)

type Parity struct {
	config *config.Config
	// DeploymentStrategies []DeploymentStrategy
}

func (g *Parity) LoadPlugins() {
	// Load Configuration
	var err error
	var confLoader *plugo.ConfigLoader
	c := &config.PluginConfig{}
	if g.config.ConfigFile != "" {
		confLoader = &plugo.ConfigLoader{}
		err = confLoader.LoadFromFile(g.config.ConfigFile, &c)
		if err != nil {
			log.Fatalf("Unable to read configuration file: %s", err.Error())
		}
	} else {
		log.Fatal("No config file provided")
	}

	log.SetLevel(log.LogLevel(c.LogLevel))

	// Load all plugins
	// g.DeploymentStrategies = make([]DeploymentStrategy, len(c.Deployment))
	// plugins := plugo.LoadPluginsWithConfig(confLoader, c.Deployment)
	// for i, p := range plugins {
	// 	log.Debug("Loading plugin\t" + log.Colorize(log.YELLOW, c.Deployment[i].Name))
	// 	g.DeploymentStrategies[i] = p.(DeploymentStrategy)
	// }
}

// MergeConfig merges ~/.parityrc with any ./parity.yml files
func (g *Parity) MergeConfig() {

}
