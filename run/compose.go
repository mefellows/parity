package run

import (
	"fmt"
	"io"
	"net"
	"os"
	"runtime"

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

// XServerProxy creates a TCP proxy on port 6000 to a the Unix
// socket that XQuartz is listening on.
//
// NOTE: this function does not start/install the XQuartz service
func XServerProxy() {
	if runtime.GOOS != "darwin" {
		log.Debug("Not running an OSX environment, skip run X Server Proxy")
		return
	}

	l, err := net.Listen("tcp", ":6000")
	if err != nil {
		log.Fatal(err)
	}
	defer l.Close()

	// Send all traffic back to unix $DISPLAY socket on a running XQuartz server
	addr, err := net.ResolveUnixAddr("unix", os.Getenv("DISPLAY"))
	if err != nil {
		log.Error("Error: ", err.Error())
	}

	// Proxy loop
	for {
		xServerClient, err := net.DialUnix("unix", nil, addr)
		if err != nil {
			log.Error("Error: ", err.Error())
		}
		defer xServerClient.Close()

		conn, err := l.Accept()
		log.Debug("X Service Proxy listening on: %s (remote: %s)", conn.LocalAddr(), conn.RemoteAddr())
		if err != nil {
			log.Fatal(err)
		}
		go func(c net.Conn, s *net.UnixConn) {
			buf := make([]byte, 8092)
			io.CopyBuffer(s, c, buf)
			s.CloseWrite()
		}(conn, xServerClient)

		go func(c net.Conn, s *net.UnixConn) {
			buf := make([]byte, 8092)
			io.CopyBuffer(c, s, buf)
			c.Close()
		}(conn, xServerClient)
	}
}

// Run the Docker Compose Run Plugin
//
// Detects docker-compose.yml files, builds and runs.
func (c *DockerCompose) Run() (err error) {
	log.Stage("Run Docker")
	log.Step("Building compose project")

	if c.project, err = c.GetProject(); err == nil {
		log.Debug("Compose - starting docker compose services")

		go XServerProxy()

		env := []string{
			"DISPLAY=192.168.99.1:0",
		}

		for _, conf := range c.project.Configs {
			existing := conf.Environment.Slice()
			env = append(env, existing...)
			conf.Environment = project.NewMaporEqualSlice(env)
		}
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
	} else {
		log.Error("Could not parse compose file: %s", err.Error())
		return p, err
	}

	return p, nil
}

func (c *DockerCompose) Attach(config parity.ShellConfig) (err error) {
	log.Stage("Interactive Shell")
	log.Debug("Compose - starting docker compose services")

	mergedConfig := *parity.DEFAULT_INTERACTIVE_SHELL_OPTIONS
	mergo.MergeWithOverwrite(&mergedConfig, &config)

	client := utils.DockerClient()
	client.SkipServerVersionCheck = true
	container := fmt.Sprintf("parity-%s_%s_1", c.pluginConfig.ProjectNameSafe, mergedConfig.Service)

	opts := dockerclient.AttachToContainerOptions{
		Stdin:        true,
		Stdout:       true,
		Stderr:       true,
		InputStream:  os.Stdin,
		OutputStream: os.Stdout,
		ErrorStream:  os.Stderr,
		RawTerminal:  true,
		Container:    container,
		Stream:       true,
		Logs:         true,
	}

	log.Step("Attaching to container '%s'", container)
	if err := client.AttachToContainer(opts); err == nil {
		c.project.Up(config.Service)
	} else {
		return err
	}

	_, err = client.WaitContainer(container)

	log.Debug("Docker Compose Run() finished")
	return err
}

func (c *DockerCompose) mergeEnvironmentArrays(new, existing []string) *project.MaporEqualSlice {
	return nil
}

// Shell creates an interactive Docker session to the specified service
// starting it if not currently running
func (c *DockerCompose) Shell(config parity.ShellConfig) (err error) {
	log.Stage("Interactive Shell")
	log.Debug("Compose - starting docker compose services")

	mergedConfig := *parity.DEFAULT_INTERACTIVE_SHELL_OPTIONS
	mergo.MergeWithOverwrite(&mergedConfig, &config)
	// Check if services running

	// if running, attach to running container
	// if NOT running, start services and attach
	if c.project, err = c.GetProject(); err == nil {
		log.Step("Starting compose services")

		env := []string{
			"DISPLAY=192.168.99.1:0",
		}

		// tmpConfig.Environment = project.
		for _, conf := range c.project.Configs {
			existing := conf.Environment.Slice()
			env = append(env, existing...)
			conf.Environment = project.NewMaporEqualSlice(env)
		}

		// Do I cal this? Restarts the services. Maybe check to see if they're up?
		c.project.Up()
	}

	client := utils.DockerClient()
	client.SkipServerVersionCheck = true
	container := fmt.Sprintf("parity-%s_%s_1", c.pluginConfig.ProjectNameSafe, mergedConfig.Service)

	createExecOptions := dockerclient.CreateExecOptions{
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          true,
		Cmd:          mergedConfig.Command,
		User:         mergedConfig.User,
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
