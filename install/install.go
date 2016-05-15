package install

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/lextoumbourou/goodhosts"
	"github.com/mefellows/parity/log"
	"github.com/mefellows/parity/utils"
	"github.com/mefellows/parity/version"
	"github.com/tmc/scp"
)

// InstallConfig is the configuration for the Installer
type InstallConfig struct {
	Dns     bool
	DevHost string
}

// InstallParity installs Parity into the running Docker Machine
func InstallParity(config InstallConfig) {
	log.Stage("Install Parity")

	if config.DevHost == "" {
		config.DevHost = "parity.local"
	}

	// Create DNS entry
	if config.Dns {
		hostname := strings.Split(utils.DockerHost(), ":")[0]
		log.Step("Creating host entry: %s -> %s", hostname, config.DevHost)
		var hosts goodhosts.Hosts
		var err error
		if hosts, err = goodhosts.NewHosts(); err == nil {
			hosts.Add(hostname, config.DevHost)
		} else {
			log.Error("Unable to create DNS Entry: %s", err.Error())
		}
		if err = hosts.Flush(); err != nil {
			log.Error("Unable to create DNS Entry: %s", err.Error())
		}
	}

	// Check - is there a Docker Machine created?

	//    -> If so, use the currently selected machine

	//    -> If not, create another machine

	//    -> Persist these settings in ~/.parityrc?

	// Wrap the local Docker command so that we don't have to use Docker Machine all of the time!

	type FileTemplate struct {
		Version string
	}
	templateData := FileTemplate{Version: version.Version}

	// Create the install mirror daemon template
	file := utils.CreateTemplateTempFile(templatesBootlocalShBytes, 0655, templateData)
	session, err := utils.SSHSession(utils.DockerHost())
	if err != nil {
		log.Fatalf("Unable to connect to Docker utils.DockerHost(). Is Docker running? (%v)", err.Error())
	}

	log.Step("Installing bootlocal.sh on Docker Host")
	remoteTmpFile := fmt.Sprintf("/tmp/%s", filepath.Base(file.Name()))
	err = scp.CopyPath(file.Name(), remoteTmpFile, session)
	utils.RunCommandWithDefaults(utils.DockerHost(), fmt.Sprintf("sudo cp %s %s", remoteTmpFile, "/var/lib/boot2docker/bootlocal.sh"))
	session.Close()

	file = utils.CreateTemplateTempFile(templatesMirrorDaemonShBytes, 0655, templateData)
	session, err = utils.SSHSession(utils.DockerHost())
	if err != nil {
		log.Fatalf("Unable to connect to Docker utils.DockerHost(). Is Docker running? (%v)", err.Error())
	}

	log.Step("Installing mirror-daemon.sh on Docker Host")
	remoteTmpFile = fmt.Sprintf("/tmp/%s", filepath.Base(file.Name()))
	err = scp.CopyPath(file.Name(), remoteTmpFile, session)
	utils.RunCommandWithDefaults(utils.DockerHost(), fmt.Sprintf("sudo cp %s %s", remoteTmpFile, "/var/lib/boot2docker/mirror-daemon.sh"))
	session.Close()

	log.Step("Downloading file sync utility (mirror)")
	utils.RunCommandWithDefaults(utils.DockerHost(), fmt.Sprintf("sudo chmod +x /var/lib/boot2docker/*.sh"))
	utils.RunCommandWithDefaults(utils.DockerHost(), fmt.Sprintf(" /var/lib/boot2docker/*.sh"))
	utils.RunCommandWithDefaults(utils.DockerHost(), fmt.Sprintf("sudo /var/lib/boot2docker/bootlocal.sh start"))

	log.Step("Restarting Docker")
	utils.RunCommandWithDefaults(utils.DockerHost(), "sudo shutdown -r now")
	utils.WaitForNetwork("docker", utils.DockerHost())
	utils.WaitForNetwork("mirror", utils.MirrorHost())

	// Removing shared folders
	if utils.CheckSharedFolders() {
		log.Step("Unmounting Virtualbox shared folders")
		utils.UnmountSharedFolders()
	}

	log.Stage("Install Parity : Complete")
}
