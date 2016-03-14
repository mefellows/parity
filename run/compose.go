package run

import (
	"fmt"
	"os"

	"github.com/docker/libcompose/docker"
	"github.com/docker/libcompose/project"
	dockerclient "github.com/fsouza/go-dockerclient"
	"github.com/imdario/mergo"
	"github.com/mefellows/parity/log"
	"github.com/mefellows/parity/parity"
	"github.com/mefellows/parity/utils"
	"github.com/mefellows/plugo/plugo"
)

// DockerCompose is a type of Run Plugin, that uses Docker Compose
// to run a local development environment
type DockerCompose struct {
	Dest         string
	ComposeFile  string `default:"docker-compose.yml" required:"true" mapstructure:"composefile"`
	pluginConfig *parity.PluginConfig
	project      *project.Project
}

func init() {
	plugo.PluginFactories.Register(func() (interface{}, error) {
		return &DockerCompose{}, nil
	}, "compose")
}

// Name of this Plugin
func (c *DockerCompose) Name() string {
	return "compose"
}

// Run the Docker Compose Run Plugin
//
// Detects docker-compose.yml files, builds and runs.
func (c *DockerCompose) Run() (err error) {
	log.Stage("Run Docker")
	log.Step("Building compose project")

	if c.project, err = c.GetProject(); err == nil {
		log.Debug("Compose - starting docker compose services")

		c.project.Delete()
		c.project.Build()
		c.project.Up()
	}

	log.Debug("Docker Compose Run() finished")
	return err
}

// GetProject returns the Docker project from the configuration
func (c *DockerCompose) GetProject() (p *project.Project, err error) {
	if _, err = os.Stat(c.ComposeFile); err == nil {
		p, err = docker.NewProject(&docker.Context{
			Context: project.Context{
				ComposeFiles: []string{c.ComposeFile},
				ProjectName:  fmt.Sprintf("parity-%s", c.pluginConfig.ProjectNameSafe),
			},
		})

		if err != nil {
			log.Error("Could not create Compose project %s", err.Error())
			return p, err
		}

		log.Debug("Compose - running docker up")
	} else {
		log.Error("Could not parse compose file: %s", err.Error())
		return p, err
	}

	return p, nil
}

// Shell creates an interactive Docker session to the specified service
// starting it if not currently running
func (c *DockerCompose) Shell(config parity.ShellConfig) (err error) {
	log.Stage("Interactive Shell")
	log.Debug("Compose - starting docker compose services")

	defaultOptions := parity.DEFAULT_INTERACTIVE_SHELL_OPTIONS
	mergo.Merge(&config, defaultOptions)
	// Start / check running services
	// Check if services running

	// if running, attach to running container
	// if NOT running, start services and attach
	if c.project, err = c.GetProject(); err == nil {
		log.Step("Starting compose services")
		c.project.Up()
	}

	client := utils.DockerClient()
	client.SkipServerVersionCheck = true
	container := fmt.Sprintf("parity-%s_%s_1", c.pluginConfig.ProjectNameSafe, config.Service)

	createExecOptions := dockerclient.CreateExecOptions{
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          true,
		Cmd:          config.Command,
		User:         config.User,
		Container:    container,
	}

	startExecOptions := dockerclient.StartExecOptions{
		Detach:       false,
		Tty:          true,
		InputStream:  os.Stdin,
		OutputStream: os.Stdout,
		ErrorStream:  os.Stderr,
		RawTerminal:  true,
	}

	log.Step("Attaching to container '%s'", container)
	if id, err := client.CreateExec(createExecOptions); err == nil {
		client.StartExec(id.ID, startExecOptions)
	} else {
		log.Error("error: %v", err.Error())
	}

	log.Debug("Docker Compose Run() finished")
	return err
}

func (c *DockerCompose) Configure(pc *parity.PluginConfig) {
	log.Debug("Configuring 'Docker Machine' 'Run' plugin")

	c.pluginConfig = pc
}

func (c *DockerCompose) Teardown() error {
	log.Debug("Tearing down 'Docker Machine' 'Run' plugin")

	if c.project != nil {
		c.project.Down()
	}
	return nil
}
