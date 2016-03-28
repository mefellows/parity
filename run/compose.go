package run

import (
	"crypto/md5"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"regexp"
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
	XProxyPort   int    `default:"6000" required:"true" mapstructure:"x_proxy_port"`
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
func XServerProxy(port int) {
	if runtime.GOOS != "darwin" {
		log.Debug("Not running an OSX environment, skip run X Server Proxy")
		return
	}

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatal(err)
	}
	defer l.Close()

	// Send all traffic back to unix $DISPLAY socket on a running XQuartz server
	addr, err := net.ResolveUnixAddr("unix", os.Getenv("DISPLAY"))
	if err != nil {
		log.Error("Error: ", err.Error())
	}
	log.Info("X Service Proxy available on all network interfaces on port %d", port)
	if host, err := utils.DockerVMHost(); err == nil {
		log.Info("Parity has detected your Docker environment and recommends running 'export DISPLAY=%s:0' in your container to forward the X display", host)
	}

	for {
		xServerClient, err := net.DialUnix("unix", nil, addr)
		if err != nil {
			log.Error("Error: ", err.Error())
		}
		defer xServerClient.Close()

		conn, err := l.Accept()
		log.Debug("X Service Proxy connected to client on: %s (remote: %s)", conn.LocalAddr(), conn.RemoteAddr())
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

func injectDisplayEnvironmentVariables(p *project.Project) {
	if host, err := utils.DockerVMHost(); err == nil {
		injectEnvironmentVariable([]string{fmt.Sprintf("DISPLAY=%s:0", host)}, p)
	}
}

func injectEnvironmentVariable(envVars []string, p *project.Project) {
	for _, conf := range p.Configs {
		existing := conf.Environment.Slice()
		envVars = append(envVars, existing...)
		conf.Environment = project.NewMaporEqualSlice(envVars)
	}
}

// runXServerProxy runs the X Server, including setting any Environment
// variables (e.g. DISPLAY)
func (c *DockerCompose) runXServerProxy() {
	XServerProxy(c.XProxyPort)
	injectDisplayEnvironmentVariables(c.project)
}

// Run the Docker Compose Run Plugin
//
// Detects docker-compose.yml files, builds and runs.
func (c *DockerCompose) Run() (err error) {
	log.Stage("Run Docker")
	log.Step("Building compose project")

	if c.project != nil {
		log.Debug("Compose - starting docker compose services")

		go c.runXServerProxy()

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

// Attach attaches to the specified service in the running container
func (c *DockerCompose) Attach(config parity.ShellConfig) (err error) {
	log.Stage("Interactive Shell")
	log.Debug("Compose - starting docker compose services")

	mergedConfig := *parity.DEFAULT_INTERACTIVE_SHELL_OPTIONS
	mergo.MergeWithOverwrite(&mergedConfig, &config)

	client := utils.DockerClient()
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

	if c.project.Configs[config.Service] == nil {
		return fmt.Errorf("Service %s does not exist", config.Service)
	}

	log.Step("Attaching to container '%s'", container)
	if err := client.AttachToContainer(opts); err == nil {
		err = c.project.Up(config.Service)
		if err != nil {
			log.Error("error: %s", err.Error())
			return err
		}
	} else {
		return err
	}

	_, err = client.WaitContainer(container)
	log.Error("wc error: %s", err.Error())

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

	// TODO: if running, attach to running container
	// if NOT running, start services and attach
	if c.project != nil {
		log.Step("Starting compose services")

		injectDisplayEnvironmentVariables(c.project)
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

// Build will build all images in the Parity setup
func (c *DockerCompose) Build(parity.BuilderConfig) error {
	// client := utils.DockerClient()

	// Check base up to date - md5 hash version

	return nil
}

// generateContainerVersion creates a unique hash for a given Dockerfile.
// It uses the contents of the Dockerfile and any package lock file (package.json, Gemfile etc.)
// Replaces this shell: `echo $(md5Files $(find -L $1 -maxdepth 1 | egrep "(Gemfile.lock|package\.json|Dockerfile)"))`
func (c *DockerCompose) generateContainerVersion(dirName string) string {
	dir, _ := os.Open(dirName)
	files, _ := dir.Readdir(-1)

	var data []byte
	regex, _ := regexp.CompilePOSIX(`(Gemfile.lock|package\.json|Dockerfile)`)

	for _, f := range files {
		if regex.MatchString(f.Name()) {
			if d, err := ioutil.ReadFile(filepath.Join(dirName, f.Name())); err == nil {
				data = append(data, d...)
			}
		}
	}
	if len(data) == 0 {
		return ""
	}

	return fmt.Sprintf("%x", md5.Sum(data))
}

// Publish pushes all images to the registries
func (c *DockerCompose) Publish(parity.BuilderConfig) error {
	return nil
}

// Configure sets up this plugin with initial state
func (c *DockerCompose) Configure(pc *parity.PluginConfig) {
	log.Debug("Configuring 'Docker Machine' 'Run' plugin")
	c.pluginConfig = pc
	var err error
	if c.project, err = c.GetProject(); err != nil {
		log.Fatalf("Unable to create Compose Project: %s", err.Error())
	}
}

// Teardown stops any running projects before Parity exits
func (c *DockerCompose) Teardown() error {
	log.Debug("Tearing down 'Docker Machine' 'Run' plugin")

	if c.project != nil {
		c.project.Down()
	}
	return nil
}
