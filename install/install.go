package install

import (
	"fmt"
	"github.com/mefellows/parity/version"
	"github.com/tmc/scp"
	"golang.org/x/crypto/ssh"
	"io"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"text/template"
)

const TMP_FILE = "/tmp/bootlocal.sh"

func CreateBoot2DockerDaemon() *os.File {

	type Foo struct {
		Version string
	}
	tmpl, err := template.ParseFiles("templates/bootlocal.sh")
	if err != nil {
		panic(err)
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

// Get the IP address of the current active Docker Machine
func DockerIp() net.IP {
	return nil
}

func SshSession(host string) (session *ssh.Session, err error) {
	config := SshConfig()
	fmt.Printf("Connecting?")
	connection, err := ssh.Dial("tcp", host, config)
	if err != nil {
		fmt.Printf("Error: %v", err)
		return nil, fmt.Errorf("Failed to dial: %s", err)
	}
	session, err = connection.NewSession()
	if err != nil {
		return nil, fmt.Errorf("Failed to create session: %s", err)
	}
	return session, nil
}

//q Run command on a given SSH connection
func RunCommand(host string, command string) (output string, err error) {
	session, err := SshSession(host)
	if err != nil {
		return "", err
	}

	modes := ssh.TerminalModes{
		ssh.ECHO:          0,     // disable echoing
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}

	if err := session.RequestPty("xterm", 80, 40, modes); err != nil {
		session.Close()
		return "", fmt.Errorf("request for pseudo terminal failed: %s", err)
	}

	stdin, err := session.StdinPipe()
	if err != nil {
		return "", fmt.Errorf("Unable to setup stdin for session: %v", err)
	}
	go io.Copy(stdin, os.Stdin)

	stdout, err := session.StdoutPipe()
	if err != nil {
		return "", fmt.Errorf("Unable to setup stdout for session: %v", err)
	}
	go io.Copy(os.Stdout, stdout)

	stderr, err := session.StderrPipe()
	if err != nil {
		return "", fmt.Errorf("Unable to setup stderr for session: %v", err)
	}
	go io.Copy(os.Stderr, stderr)
	err = session.Run(command)

	return output, nil
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

func InstallParity() {

	// Create the install mirror daemon template
	host := "docker:22"
	file := CreateBoot2DockerDaemon()
	fmt.Printf("File: %v", file.Name())
	session, err := SshSession(host)
	fmt.Printf("Error: %v", err)

	// Upload the boot script and give perms
	remoteTmpFile := fmt.Sprintf("/tmp/%s", filepath.Base(file.Name()))
	err = scp.CopyPath(file.Name(), remoteTmpFile, session)
	RunCommand(host, fmt.Sprintf("sudo cp %s %s", remoteTmpFile, "/var/lib/boot2docker/bootlocal.sh"))
	fmt.Printf("Error: %v", err)
	session.Close()

	session, err = SshSession(host)
	err = scp.CopyPath("./templates/mirror-daemon.sh", remoteTmpFile, session)
	fmt.Printf("Error: %v", err)
	RunCommand(host, fmt.Sprintf("sudo cp %s %s", remoteTmpFile, "/var/lib/boot2docker/mirror-daemon.sh"))

	// Execute that script ()

	//

}
