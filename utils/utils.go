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

	"github.com/mefellows/parity/log"

	"github.com/docker/libcompose/docker"
	"github.com/docker/libcompose/project"
	// "github.com/docker/machine/libmachine"
	dockerclient "github.com/fsouza/go-dockerclient"
	"github.com/mitchellh/cli"
	"golang.org/x/crypto/ssh"
)

const bootlocalTemplateFile = "templates/bootlocal.sh"
const daemonTemplateFile = "templates/mirror-daemon.sh"
const TMP_FILE = "/tmp/bootlocal.sh"

// Resurrect a go-bindata asset and execute the template (if any) it contains.
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

// Create a docker client from environment
func dockerClient() *dockerclient.Client {
	client, err := dockerclient.NewClientFromEnv()
	if err != nil {
		log.Fatalf("Unabled to create a Docker Client: Is Docker Machine installed and running?")
	}
	return client
}

// Docker Host port
func dockerPort() string {
	return "22"
}

// Mirror Daemon port
func mirrorPort() string {
	return "8123"
}

// Get the IP address of the current active Docker Machine
func DockerHost() string {
	return fmt.Sprintf("%s:%s", dockerHost(), dockerPort())
}

// Get the IP address of the current active Docker Machine
func MirrorHost() string {
	return fmt.Sprintf("%s:%s", dockerHost(), mirrorPort())
}

// Create an SSH Session for a remote Docker host
func SshSession(host string) (session *ssh.Session, err error) {
	config := SshConfig()
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

// Run a command on a host using the default configuration for I/O redirection
func RunCommandWithDefaults(host string, command string) error {
	return RunCommand(host, command, os.Stdin, os.Stdout, os.Stderr)
}

// Run a command on a given SSH Connection and return the output of Stdout
func RunCommandAndReturn(host string, command string) (string, error) {
	var output bytes.Buffer
	RunCommand(host, command, os.Stdin, &output, os.Stderr)
	return output.String(), nil
}

// Run command on a given SSH connection
func RunCommand(host string, command string, reader io.Reader, stdOut io.Writer, stdErr io.Writer) error {
	session, err := SshSession(host)
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

// Get an SSH configuration to the Docker Host
func SshConfig() *ssh.ClientConfig {
	return &ssh.ClientConfig{
		User: dockerUser(),
		Auth: []ssh.AuthMethod{
			PublicKeyFile(dockerCertPath()),
		},
	}
}

// Read a public key and
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

// Wait for a network connection to become available within a timeout
func WaitForNetwork(name string, host string) {
	WaitForNetworkWithTimeout(name, host, 60*time.Second)
}

// Wait for a network connection to become available within a timeout
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
			log.Info("Connected to", name)
			break WaitLoop
		case <-time.After(timeout):
			log.Fatalf("Unable to connect to %s %s", name, "Is Docker running?")
		}
	}

	return
}

// Get the list of shared folders on the remote Docker Host
func FindSharedFolders() []string {
	sharesRes, err := RunCommandAndReturn(DockerHost(), "mount | grep 'type vboxsf' | awk '{print $3}'")
	if err != nil {
		log.Warn("Unable to determine Virtualbox shared folders, please manually ensure shared folders are removed to ensure proper operation of Parity")
	}
	shares := make([]string, 0)
	for _, s := range strings.Split(sharesRes, "\n") {
		if s != "" {
			shares = append(shares, s)
		}
	}
	return shares
}

// Unmount all shared folders on the remote Docker host
func UnmountSharedFolders() {
	shares := FindSharedFolders()
	for _, s := range shares {
		share := strings.TrimSpace(s)
		RunCommandWithDefaults(DockerHost(), fmt.Sprintf(`sudo umount "%s"`, share))
	}
}

// Return true if shared folders exist and the user agrees to removing them
func CheckSharedFolders(ui cli.Ui) bool {
	shares := FindSharedFolders()
	if len(shares) > 0 {
		fmt.Printf("%v", shares)
		res, err := ui.Ask("For Parity to operate properly, Virtualbox shares must be removed. Would you like us to automatically do this for you? (yes/no)")
		return err == nil && res == "yes"
	}
	return false
}

func interactiveDocker() {
	//client := dockerClient()
	//container, err := client.CreateContainer(createContainerOptions)

}

func FindDockerComposeFiles() []string {
	return []string{"docker-compose.yml"}
}

// Read a docker-compose.yml and return a slice of
// directories to sync into the Docker Host
//
// "." is converted to the current directory parity is running from
func ReadComposeVolumes() []string {
	volumes := make([]string, 0)

	files := FindDockerComposeFiles()
	for i, file := range files {
		if _, err := os.Stat(file); err == nil {
			project, err := docker.NewProject(&docker.Context{
				Context: project.Context{
					ComposeFiles: []string{file},
					ProjectName:  fmt.Sprintf("parity-", i),
				},
			})

			if err != nil {
				log.Info("Could not parse compose file")
			}

			for _, c := range project.Configs {
				for _, v := range c.Volumes {
					v = strings.SplitN(v, ":", 2)[0]
					if v == "." {
						v, _ = os.Getwd()
					}
					volumes = append(volumes, v)
				}
			}
		}
	}

	return volumes
}
