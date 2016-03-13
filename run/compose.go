package run

import (
	"fmt"
	"os"

	"github.com/docker/libcompose/docker"
	"github.com/docker/libcompose/project"
	"github.com/mefellows/parity/log"
	"github.com/mefellows/parity/parity"
	"github.com/mefellows/plugo/plugo"
)

// DockerCompose is a type of Run Plugin, that uses Docker Compose
// to run a local development environment
type DockerCompose struct {
	// sync.Mutex
	Dest         string
	ComposeFile  string `default:"docker-compose.yml" required:"true" mapstructure:"composefile"`
	pluginConfig *parity.PluginConfig
	project      *project.Project
}

func init() {
	log.Info("Initing docker machine")
	plugo.PluginFactories.Register(func() (interface{}, error) {
		log.Info("Initing docker machine")
		return &DockerCompose{}, nil
	}, "machine")
}

func (m *DockerCompose) Run() error {
	log.Debug("Running docker compose")

	if _, err := os.Stat(m.ComposeFile); err == nil {
		m.project, err = docker.NewProject(&docker.Context{
			Context: project.Context{
				ComposeFiles: []string{m.ComposeFile},
				ProjectName:  fmt.Sprintf("parity-%s", m.pluginConfig.ProjectNameSafe),
			},
		})

		if err != nil {
			log.Error("Could not parse compose file")
		}

		log.Debug("Compose - running docker up")
		m.project.Delete()
		m.project.Build()
		m.project.Up()
	}

	log.Info("Done docker up ing")
	return nil
}

func (m *DockerCompose) Configure(c *parity.PluginConfig) {
	log.Debug("Configuring 'Docker Machine' 'Run' plugin")
	m.pluginConfig = c
}

func (m *DockerCompose) Teardown() {
	log.Debug("Tearing down 'Docker Machine' 'Run' plugin")
	if m.project != nil {
		m.project.Down()
	}
}
