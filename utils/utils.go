package utils

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/url"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/docker/libcompose/docker"
	"github.com/docker/libcompose/project"
	dockerclient "github.com/fsouza/go-dockerclient"
	mutils "github.com/mefellows/mirror/filesystem/utils"
	"github.com/mefellows/parity/log"
	"golang.org/x/crypto/ssh"
)

const bootlocalTemplateFile = "templates/bootlocal.sh"
const daemonTemplateFile = "templates/mirror-daemon.sh"

// CreateTemplateTempFile resurrects a go-bindata asset and execute the template (if any) it contains.
// return a reference to the temporary file
func CreateTemplateTempFile(data func() ([]byte, error), perms os.FileMode, templateData interface{}) *os.File {
	daemon, err := data()
	if err != nil {
		log.Fatalf("Template failed:", err.Error())
	}

	tmpl, err := template.New("template file").Parse(string(daemon))
	if err != nil {
		log.Fatalf("Template failed:", err.Error())
	}

	file, _ := ioutil.TempFile("/tmp", "parity")
	file.Chmod(perms)

	err = tmpl.Execute(file, templateData)
	if err != nil {
		panic(err)
	}

	return file
}

// TODO: Don't rely on environment variables. If they are set, great, otherwise default
// to

// Get the hostname (sans protocol/port) of the current Docker Host
func dockerHost() string {
	host, _ := url.Parse(os.Getenv("DOCKER_HOST"))
	return strings.Split(host.Host, ":")[0]
}

// Get the user required to connect to the remote Docker host
func dockerUser() string {
	return "docker"
}

// Get the name of the current running Docker Machine
func dockerMachineName() string {
	return os.Getenv("DOCKER_MACHINE_NAME")
}

// Get the path to the current Docker Machine's client certificate
func dockerCertPath() string {
	return fmt.Sprintf("%s/id_rsa", os.Getenv("DOCKER_CERT_PATH"))
}

// DefaultParityConfigurationFile gets the default parity configuration file
func DefaultParityConfigurationFile() string {
	dir, _ := os.Getwd()
	return fmt.Sprintf("%s/parity.yml", dir)
}

// DefaultComposeFile gets the default docker-compose.yml file
func DefaultComposeFile() string {
	dir, _ := os.Getwd()
	return fmt.Sprintf("%s/docker-compose.yml", dir)
}

// FindNetwork will return the IP and Network interface
// given an IP address.
func FindNetwork(ip string) (net.IP, net.Addr, error) {
	addr := net.ParseIP(ip)

	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, nil, fmt.Errorf("Unable to lookup local network (%s)", err.Error())
	}

	for _, i := range ifaces {
		if addrs, err := i.Addrs(); err == nil {
			for _, a := range addrs {
				switch a.(type) {
				case *net.IPNet:
					ip, network, _ := net.ParseCIDR(a.String())
					if network.Contains(addr) {
						return ip, network, nil
					}
				}
			}
		}
	}
	return nil, nil, fmt.Errorf("Unable to find network for ip %s", ip)
}

// DockerClient creates a docker client from environment
func DockerClient() *dockerclient.Client {
	client, err := dockerclient.NewClientFromEnv()
	if err != nil {
		log.Fatalf("Unabled to create a Docker Client: Is Docker Machine installed and running?")
	}
	client.SkipServerVersionCheck = true
	return client
}

// dockerPort returns the SSH port for Docker
func dockerPort() string {
	return "22"
}

// mirrorPort returns the Mirror daemon port
func mirrorPort() string {
	return "8123"
}

// DockerHost gets the ip:port of the current active Docker Machine
func DockerHost() string {
	return fmt.Sprintf("%s:%s", dockerHost(), dockerPort())
}

// DockerVMHost gets the IP address of the underlying VM for the current active Docker Machine
func DockerVMHost() (string, error) {
	ip, _, err := FindNetwork(dockerHost())
	if err == nil {
		return ip.String(), nil
	}
	return "", err
}

// MirrorHost gets the ip:port of the current active Docker Machine
func MirrorHost() string {
	return fmt.Sprintf("%s:%s", dockerHost(), mirrorPort())
}

// SSHSession creates an SSH Session for a remote Docker host
func SSHSession(host string) (session *ssh.Session, err error) {
	config := SSHConfig()
	connection, err := ssh.Dial("tcp", host, config)
	if err != nil {
		return nil, fmt.Errorf("Failed to dial: %s", err)
	}
	session, err = connection.NewSession()
	if err != nil {
		return nil, fmt.Errorf("Failed to create session: %s", err)
	}
	return session, nil
}

// RunCommandWithDefaults runs a command on a host using the default configuration for I/O redirection
func RunCommandWithDefaults(host string, command string) error {
	return RunCommand(host, command, os.Stdin, os.Stdout, os.Stderr)
}

// RunCommandAndReturn runs a command on a given SSH Connection and return the output of Stdout
func RunCommandAndReturn(host string, command string) (string, error) {
	var output bytes.Buffer
	RunCommand(host, command, os.Stdin, &output, os.Stderr)
	return output.String(), nil
}

// RunCommand runs command on a given SSH connection
func RunCommand(host string, command string, reader io.Reader, stdOut io.Writer, stdErr io.Writer) error {
	session, err := SSHSession(host)
	if err != nil {
		return err
	}
	defer session.Close()

	modes := ssh.TerminalModes{
		ssh.ECHO:          0,     // disable echoing
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}

	if err := session.RequestPty("xterm", 80, 40, modes); err != nil {
		session.Close()
		return fmt.Errorf("request for pseudo terminal failed: %s", err)
	}

	session.Stdout = stdOut
	session.Stdin = reader
	session.Stderr = stdErr
	err = session.Run(command)

	return nil
}

// SSHConfig gets an SSH configuration to the Docker Host
func SSHConfig() *ssh.ClientConfig {
	return &ssh.ClientConfig{
		User: dockerUser(),
		Auth: []ssh.AuthMethod{
			PublicKeyFile(dockerCertPath()),
		},
	}
}

// PublicKeyFile reads a public key and returns an SSH auth method
func PublicKeyFile(file string) ssh.AuthMethod {
	buffer, err := ioutil.ReadFile(file)
	if err != nil {
		return nil
	}

	key, err := ssh.ParsePrivateKey(buffer)
	if err != nil {
		return nil
	}
	return ssh.PublicKeys(key)
}

// WaitForNetwork waits for a network connection to become available within a timeout
func WaitForNetwork(name string, host string) {
	WaitForNetworkWithTimeout(name, host, 120*time.Second)
}

// WaitForNetworkWithTimeout waits for a network connection to become available within a timeout
func WaitForNetworkWithTimeout(name string, host string, timeout time.Duration) {
	waitDone := make(chan bool, 1)
	go func() {
		log.Info("Waiting for %s to become available (%s)", name, host)
		for {
			select {
			case <-time.After(5 * time.Second):
			}

			_, err := net.DialTimeout("tcp", host, 10*time.Second)
			if err != nil {
				continue
			}
			waitDone <- true
		}
	}()

WaitLoop:
	for {
		select {
		case <-waitDone:
			log.Info("Connected to %s", name)
			break WaitLoop
		case <-time.After(timeout):
			log.Fatalf("Unable to connect to %s %s", name, "Is Docker running?")
		}
	}

	return
}

// FindSharedFolders gets the list of shared folders on the remote Docker Host
func FindSharedFolders() []string {
	sharesRes, err := RunCommandAndReturn(DockerHost(), "mount | grep 'type vboxsf' | awk '{print $3}'")
	if err != nil {
		log.Warn("Unable to determine Virtualbox shared folders, please manually ensure shared folders are removed to ensure proper operation of Parity")
	}
	var shares []string
	for _, s := range strings.Split(sharesRes, "\n") {
		if s != "" {
			shares = append(shares, s)
		}
	}
	return shares
}

// UnmountSharedFolders unmounts all shared folders on the remote Docker host
func UnmountSharedFolders() {
	shares := FindSharedFolders()
	for _, s := range shares {
		share := strings.TrimSpace(s)
		RunCommandWithDefaults(DockerHost(), fmt.Sprintf(`sudo umount "%s"`, share))
	}
}

// CheckSharedFolders seturn true if shared folders exist and the user agrees to removing them
func CheckSharedFolders() bool {
	shares := FindSharedFolders()
	if len(shares) > 0 {
		log.Warn("For Parity to operate properly, Virtualbox shares must be removed. Parity will automatically do this for you")
		return true
	}
	return false
}

// FindDockerComposeFiles returns the list of Docker Compose files
// in the current project. Currently just defaults to a single ['docker-compose.yml']
func FindDockerComposeFiles() []string {
	return []string{"docker-compose.yml"}
}

// ReadComposeVolumes reads a docker-compose.yml and return a slice of
// directories to sync into the Docker Host
//
// "." and "./." is converted to the current directory parity is running from.
// Any volume starting with "/" will be treated as an absolute path.
// All other volumes (e.g. starting with "./" or without a prefix "/") will be treated as
// relative paths.
func ReadComposeVolumes() []string {
	var volumes []string

	files := FindDockerComposeFiles()
	for i, file := range files {
		if _, err := os.Stat(file); err == nil {
			project, err := docker.NewProject(&docker.Context{
				Context: project.Context{
					ComposeFiles: []string{file},
					ProjectName:  fmt.Sprintf("parity-%d", i),
				},
			})

			if err != nil {
				log.Info("Could not parse compose file")
			}

			for _, c := range project.Configs {
				for _, v := range c.Volumes {
					v = strings.SplitN(v, ":", 2)[0]

					if v == "." || v == "./." {
						v, _ = os.Getwd()
					} else if strings.Index(v, "/") != 0 {
						cwd, _ := os.Getwd()
						v = fmt.Sprintf("%s/%s", cwd, v)
					}
					volumes = append(volumes, mutils.LinuxPath(v))
				}
			}
		}
	}

	return volumes
}

// ProjectNameSafe creates a Docker Compose compatible (safe) name given a string
func ProjectNameSafe(name string) string {
	return strings.Replace(strings.ToLower(name), " ", "", -1)
}

// CleanupDockerContainersList returns a list of containers to be removed
func CleanupDockerContainersList() ([]dockerclient.APIContainers, error) {
	client := DockerClient()
	listOpts := dockerclient.ListContainersOptions{
		Filters: map[string][]string{
			"status": []string{"exited"},
		},
	}

	return client.ListContainers(listOpts)
}

// CleanupDockerImageList returns a list of images to be removed
func CleanupDockerImageList() ([]dockerclient.APIImages, error) {
	client := DockerClient()
	listOpts := dockerclient.ListImagesOptions{
		Filters: map[string][]string{
			"dangling": []string{"true"},
		},
	}

	return client.ListImages(listOpts)
}

// Cleanup removes dangling images and exited containers.
func Cleanup() {
	client := DockerClient()
	if list, err := CleanupDockerImageList(); err == nil {
		for _, i := range list {
			client.RemoveImage(i.ID)
		}
	}
	if list, err := CleanupDockerContainersList(); err == nil {
		for _, c := range list {
			opts := dockerclient.RemoveContainerOptions{ID: c.ID}
			client.RemoveContainer(opts)
		}
	}
}
