package install

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/docker/libcompose/docker"
	"github.com/docker/libcompose/project"
	dockerclient "github.com/fsouza/go-dockerclient"
	"github.com/mefellows/parity/version"
	"github.com/mitchellh/cli"
	"github.com/tmc/scp"
	"golang.org/x/crypto/ssh"
)

const bootlocalTemplateFile = "templates/bootlocal.sh"
const daemonTemplateFile = "templates/mirror-daemon.sh"
const TMP_FILE = "/tmp/bootlocal.sh"

func CreateBoot2DockerDaemon() *os.File {

	type Boot2DockerTemplate struct {
		Version string
	}
	daemon, err := templatesBootlocalShBytes()
	if err != nil {
		log.Fatalf("CreateBoot2DockerDaemon template failed:", err.Error())
	}
	tmpl, err := template.New("boot2docker daemon").Parse(string(daemon))
	if err != nil {
		log.Fatalf("CreateBoot2DockerDaemon template failed:", err.Error())
	}
	someStruct := Boot2DockerTemplate{Version: version.Version}
	file, _ := ioutil.TempFile("/tmp", "parity")
	file.Chmod(0655)
	err = tmpl.Execute(file, someStruct)
	if err != nil {
		panic(err)
	}
	return file
}

func dockerHost() string {
	host, _ := url.Parse(os.Getenv("DOCKER_HOST"))
	return strings.Split(host.Host, ":")[0]
}
func dockerMachineName() string {
	return os.Getenv("DOCKER_MACHINE_NAME")
}

func dockerCertPath() string {
	return fmt.Sprintf("%s/id_rsa", os.Getenv("DOCKER_CERT_PATH"))
}

func dockerClient() *dockerclient.Client {
	client, err := dockerclient.NewClientFromEnv()
	if err != nil {
		log.Fatalf("Unabled to create a Docker Client: Is Docker Machine installed and running?")
	}
	return client
}

func dockerPort() string {
	return "22"
}
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

func RunCommandWithDefaults(host string, command string) error {
	return RunCommand(host, command, os.Stdin, os.Stdout, os.Stderr)
}

func RunCommandAndReturn(host string, command string) (string, error) {
	var output bytes.Buffer
	RunCommand(DockerHost(), command, os.Stdin, &output, os.Stderr)
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

func SshConfig() *ssh.ClientConfig {
	return &ssh.ClientConfig{
		User: "docker",
		Auth: []ssh.AuthMethod{
			PublicKeyFile(fmt.Sprintf("%s/.docker/machine/machines/%s/id_rsa", os.Getenv("HOME"), "dev")),
		},
	}
}

func Scp(file string, dest, session *ssh.Session) {
}

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

func WaitForNetwork(name string, host string) {
	waitDone := make(chan bool, 1)
	go func() {
		log.Println("Waiting for", name, "to become available (", host, ")")
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
			log.Println("Connected to", name)
			break WaitLoop
		case <-time.After(60 * time.Second):
			log.Fatalf("Unable to connect to %s %s", name, "Is Docker running?")
		}
	}

	return
}

func FindSharedFolders() []string {
	sharesRes, err := RunCommandAndReturn(DockerHost(), "mount | grep 'type vboxsf' | awk '{print $3}'")
	if err != nil {
		log.Println("Unable to determine Virtualbox shared folders, please manually ensure shared folders are removed to ensure proper operation of Parity")
	}
	shares := make([]string, 0)
	for _, s := range strings.Split(sharesRes, "\n") {
		if s != "" {
			shares = append(shares, s)
		}
	}
	return shares

}

func UnmountSharedFolders() {
	shares := FindSharedFolders()
	for _, s := range shares {
		share := strings.TrimSpace(s)
		RunCommandWithDefaults(DockerHost(), fmt.Sprintf(`sudo umount "%s"`, share))
	}
}
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

func ReadComposeVolumes() []string {
	volumes := make([]string, 0)
	if _, err := os.Stat("docker-compose.yml"); err == nil {
		project, err := docker.NewProject(&docker.Context{
			Context: project.Context{
				ComposeFiles: []string{"docker-compose.yml"},
				ProjectName:  "my-compose",
			},
		})
		if err != nil {
			log.Println("Could not parse compose file")
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

	return volumes
}

func InstallParity(ui cli.Ui) {
	// Create the install mirror daemon template
	file := CreateBoot2DockerDaemon()
	session, err := SshSession(DockerHost())
	if err != nil {
		log.Fatalf("Unable to connect to Docker DockerHost(). Is Docker running? (%v)", err.Error())
	}

	log.Printf("Installing files on Docker Host")
	remoteTmpFile := fmt.Sprintf("/tmp/%s", filepath.Base(file.Name()))
	err = scp.CopyPath(file.Name(), remoteTmpFile, session)
	RunCommandWithDefaults(DockerHost(), fmt.Sprintf("sudo cp %s %s", remoteTmpFile, "/var/lib/boot2docker/bootlocal.sh"))
	session.Close()
	session, err = SshSession(DockerHost())
	err = scp.CopyPath("./templates/mirror-daemon.sh", remoteTmpFile, session)
	RunCommandWithDefaults(DockerHost(), fmt.Sprintf("sudo cp %s %s", remoteTmpFile, "/var/lib/boot2docker/mirror-daemon.sh"))

	log.Println("Downloading file sync utility (mirror)")
	RunCommandWithDefaults(DockerHost(), fmt.Sprintf("sudo /var/lib/boot2docker/bootlocal.sh start"))

	log.Println("Restarting Docker")
	RunCommandWithDefaults(DockerHost(), "sudo shutdown -r now")
	WaitForNetwork("docker", DockerHost())
	WaitForNetwork("mirror", MirrorHost())

	// Removing shared folders
	if CheckSharedFolders(ui) {
		log.Println("Unmounting Virtualbox shared folders")
		UnmountSharedFolders()
	}

	log.Println("Parity installed. Run 'parity up' to to get started!")
}
