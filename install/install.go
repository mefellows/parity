package install

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/mefellows/parity/version"
	"github.com/mitchellh/cli"
	"github.com/tmc/scp"
	"golang.org/x/crypto/ssh"
)

const bootlocalTemplateFile = "templates/bootlocal.sh"
const daemonTemplateFile = "templates/mirror-daemon.sh"
const TMP_FILE = "/tmp/bootlocal.sh"

func CreateBoot2DockerDaemon() *os.File {

	type Foo struct {
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
	someStruct := Foo{Version: version.Version}
	file, _ := ioutil.TempFile("/tmp", "parity")
	file.Chmod(0655)
	err = tmpl.Execute(file, someStruct)
	if err != nil {
		panic(err)
	}
	return file
}

func dockerHost() string {
	return "docker"
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
	/*
		stdin, err := session.StdinPipe()
		if err != nil {
			return fmt.Errorf("Unable to setup stdin for session: %v", err)
		}
		go io.Copy(stdin, reader)

		stdout, err := session.StdoutPipe()
		if err != nil {
			return fmt.Errorf("Unable to setup stdout for session: %v", err)
		}
		go io.Copy(stdOut, stdout)

		stderr, err := session.StderrPipe()
		if err != nil {
			return fmt.Errorf("Unable to setup stderr for session: %v", err)
		}
		go io.Copy(stdErr, stderr)
	*/

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
			log.Fatalf("Unable to connect to", name, "Is Docker running?")
		}
	}

	return
}

func FindSharedFolders() []string {
	sharesRes, err := RunCommandAndReturn(DockerHost(), "mount | grep 'type vboxsf' | awk '{print $3}'")
	if err != nil {
		log.Println("Unable to determine Virtualbox shared folders, please manually ensure shared folders are removed to ensure proper operation of Parity")
	}
	return strings.Split(sharesRes, "\n")

}

func UnmountSharedFolders() {
	shares := FindSharedFolders()
	for _, s := range shares {
		share := strings.TrimSpace(s)
		if share != "" {
			RunCommandWithDefaults(DockerHost(), fmt.Sprintf(`sudo umount "%s"`, share))
		}
	}
}
func CheckSharedFolders(ui cli.Ui) bool {
	res, err := ui.Ask("For Parity to operate properly, it requires the default Virtualbox shared folders to be removed,\nwould you like us to automatically unmount them for you? (yes/no)\n")
	return err == nil || res == "yes"
}

func InstallParity(ui cli.Ui) {
	os.Exit(1)

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
