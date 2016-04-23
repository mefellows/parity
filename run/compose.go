package run

import (
	"crypto/md5"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	dockerclient2 "github.com/docker/engine-api/client"
	// dockertypes "github.com/docker/engine-api/types"
	"github.com/Sirupsen/logrus"
	"github.com/docker/docker/builder"
	"github.com/docker/docker/builder/dockerignore"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/docker/pkg/fileutils"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/docker/docker/pkg/progress"
	"github.com/docker/docker/pkg/streamformatter"
	"github.com/docker/docker/pkg/term"
	"github.com/docker/engine-api/types"
	"github.com/docker/libcompose/docker"
	"github.com/docker/libcompose/project"
	dockerclient "github.com/fsouza/go-dockerclient"
	"github.com/imdario/mergo"
	"github.com/mefellows/parity/log"
	"github.com/mefellows/parity/parity"
	"github.com/mefellows/parity/utils"
	"github.com/mefellows/plugo/plugo"
	"golang.org/x/net/context"
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
	}

	container := fmt.Sprintf("parity-%s_%s_1", c.pluginConfig.ProjectNameSafe, mergedConfig.Service)

	client := utils.DockerClient()

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

	/*
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		client, _ := dockerclient2.NewEnvClient()
		execConfig := dockertypes.ExecConfig{
			User: mergedConfig.User, // User that will run the command
			// Privileged   bool     // Is the container in privileged mode
			Tty:          true,      // Attach standard streams to a tty.
			Container:    container, // Name of the container (to execute in)
			AttachStdin:  true,      // Attach the standard input, makes possible user interaction
			AttachStderr: true,      // Attach the standard output
			AttachStdout: true,      // Attach the standard error
			Detach:       false,     // Execute in detach mode
			// DetachKeys:   "ctrl-c",             // Escape keys for detach
			Cmd: mergedConfig.Command, // Execution commands and args
		}
		execStartConfig := dockertypes.ExecStartCheck{
			Tty:    true, // Attach standard streams to a tty.
			Detach: false,
		}

		execId, err := client.ContainerExecCreate(ctx, execConfig)
		if err != nil {
			log.Error(err.Error())
			return err
		}
		err = client.ContainerExecStart(ctx, execId.ID, execStartConfig)
		if err != nil {
			log.Error(err.Error())
			return err
		}
		res, err := client.ContainerExecAttach(ctx, execId.ID, execConfig)
		if err != nil {
			log.Error(err.Error())
			return err
		}
		go res.Reader.WriteTo(os.Stdin)
		go res.Reader.Reset(os.Stdout)
		log.Step("started exec attach?")

		// done := make(chan bool)
		// <-done
		// res.Close()
		_, err = client.ContainerWait(ctx, execId.ID)
		if err != nil {
			log.Fatal(err)
		}
	*/

	log.Debug("Docker Compose Run() finished")
	return err
}

// generateContainerVersion creates a unique hash for a given Dockerfile.
// It uses the contents of the Dockerfile and any package lock file (package.json, Gemfile etc.)
// Replaces this shell: `echo $(md5Files $(find -L $1 -maxdepth 1 | egrep "(Gemfile.lock|package\.json|Dockerfile)"))`
func (c *DockerCompose) generateContainerVersion(dirName string, dockerfile string) string {
	log.Debug("Looking for %s and related package files in: %s", dockerfile, dirName)
	dir, _ := os.Open(dirName)
	files, _ := dir.Readdir(-1)

	var data []byte
	regex, _ := regexp.CompilePOSIX(fmt.Sprintf("(Gemfile.lock|package\\.json|^%s$)", dockerfile))

	for _, f := range files {
		if regex.MatchString(f.Name()) {
			log.Debug("Found file: %s", f.Name())
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
	log.Debug("Configuring 'Docker Machine' 'Run\\Build\\Shell' plugin")
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

// CreateTar create a build context tar for the specified project and service name.
func (c *DockerCompose) CreateTar(root string, dockerfile string) (io.ReadCloser, error) {
	// This code was mostly ripped off from docker/api/client/build.go

	dockerfileName := filepath.Join(root, dockerfile)

	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}

	filename := dockerfileName

	if dockerfileName == "" {
		// No -f/--file was specified so use the default
		dockerfileName = "Dockerfile"
		filename = filepath.Join(absRoot, dockerfileName)

		// Just to be nice ;-) look for 'dockerfile' too but only
		// use it if we found it, otherwise ignore this check
		if _, err = os.Lstat(filename); os.IsNotExist(err) {
			tmpFN := path.Join(absRoot, strings.ToLower(dockerfileName))
			if _, err = os.Lstat(tmpFN); err == nil {
				dockerfileName = strings.ToLower(dockerfileName)
				filename = tmpFN
			}
		}
	}

	origDockerfile := dockerfileName // used for error msg
	if filename, err = filepath.Abs(filename); err != nil {
		return nil, err
	}

	// Now reset the dockerfileName to be relative to the build context
	dockerfileName, err = filepath.Rel(absRoot, filename)
	if err != nil {
		return nil, err
	}

	// And canonicalize dockerfile name to a platform-independent one
	dockerfileName, err = archive.CanonicalTarNameForPath(dockerfileName)
	if err != nil {
		return nil, fmt.Errorf("Cannot canonicalize dockerfile path %s: %v", dockerfileName, err)
	}

	if _, err = os.Lstat(filename); os.IsNotExist(err) {
		return nil, fmt.Errorf("Cannot locate Dockerfile: %s", origDockerfile)
	}
	var includes = []string{"."}
	var excludes []string

	dockerIgnorePath := path.Join(root, ".dockerignore")
	dockerIgnore, err := os.Open(dockerIgnorePath)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
		logrus.Warnf("Error while reading .dockerignore (%s) : %s", dockerIgnorePath, err.Error())
		excludes = make([]string, 0)
	} else {
		excludes, err = dockerignore.ReadAll(dockerIgnore)
		if err != nil {
			return nil, err
		}
	}

	// If .dockerignore mentions .dockerignore or the Dockerfile
	// then make sure we send both files over to the daemon
	// because Dockerfile is, obviously, needed no matter what, and
	// .dockerignore is needed to know if either one needs to be
	// removed.  The deamon will remove them for us, if needed, after it
	// parses the Dockerfile.
	keepThem1, _ := fileutils.Matches(".dockerignore", excludes)
	keepThem2, _ := fileutils.Matches(dockerfileName, excludes)
	if keepThem1 || keepThem2 {
		includes = append(includes, ".dockerignore", dockerfileName)
	}

	if err := builder.ValidateContextDirectory(root, excludes); err != nil {
		return nil, fmt.Errorf("Error checking context is accessible: '%s'. Please check permissions and try again.", err)
	}

	options := &archive.TarOptions{
		Compression:     archive.Uncompressed,
		ExcludePatterns: excludes,
		IncludeFiles:    includes,
	}

	return archive.TarWithOptions(root, options)
}

// Build will build all images in the Parity setup
func (c *DockerCompose) Build(config parity.BuilderConfig) error {
	log.Stage("Bulding containers")
	base := "Dockerfile"
	cwd, _ := os.Getwd()
	baseVersion := c.generateContainerVersion(cwd, base)
	imageName := fmt.Sprintf("%s:%s", "web", baseVersion)
	// imageName := fmt.Sprintf("%s:%s", config.ImageName, baseVersion)
	client, _ := dockerclient2.NewEnvClient()

	log.Step("Checking if image %s exists locally", imageName)
	if images, err := client.ImageList(context.Background(), types.ImageListOptions{MatchName: imageName}); err == nil {
		for _, i := range images {
			log.Info("Found image: %s", i.ID)
			return nil
		}
	}

	log.Step("Image %s not found locally, pulling", imageName)
	client.ImagePull(context.Background(), types.ImagePullOptions{ImageID: imageName}, nil)

	log.Step("Image %s not found anywhere, building", imageName)

	ctx, err := c.CreateTar(".", "Dockerfile")
	if err != nil {
		return err
	}
	defer ctx.Close()

	var progBuff io.Writer = os.Stdout
	var buildBuff io.Writer = os.Stdout

	// Setup an upload progress bar
	progressOutput := streamformatter.NewStreamFormatter().NewProgressOutput(progBuff, true)

	var body io.Reader = progress.NewProgressReader(ctx, progressOutput, 0, "", "Sending build context to Docker daemon")

	logrus.Infof("Building %s...", imageName)

	outFd, isTerminalOut := term.GetFdInfo(os.Stdout)

	response, err := client.ImageBuild(context.Background(), types.ImageBuildOptions{
		Context:    body,
		Tags:       []string{imageName},
		NoCache:    false,
		Remove:     true,
		Dockerfile: "Dockerfile",
		// AuthConfigs: d.context.ConfigFile.AuthConfigs,
	})

	if err != nil {
		log.Error(err.Error())
		return err
	}

	err = jsonmessage.DisplayJSONMessagesStream(response.Body, buildBuff, outFd, isTerminalOut, nil)
	if err != nil {
		if jerr, ok := err.(*jsonmessage.JSONError); ok {
			// If no error code is set, default to 1
			if jerr.Code == 0 {
				jerr.Code = 1
			}
			fmt.Fprintf(os.Stderr, "%s%s", progBuff, buildBuff)
			return fmt.Errorf("Status: %s, Code: %d", jerr.Message, jerr.Code)
		}
	}
	return err
}
