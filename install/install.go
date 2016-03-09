package install

import (
	"github.com/mefellows/parity/utils"
	"github.com/mitchellh/cli"
	"github.com/tmc/scp"
	"log"
	"fmt"
	"path/filepath"
)

func InstallParity(ui cli.Ui) {
	// Check - is there a Docker Machine created?

	//    -> If so, use the currently selected machine

	//    -> If not, create another machine

	//    -> Persist these settings in ~/.parityrc?

	// Wrap the local Docker command so that we don't have to use Docker Machine all of the time!

	// Create the install mirror daemon template
	// Create the install mirror daemon template
	file := utils.CreateTemplateTempFile(templatesBootlocalShBytes, 0666)
	session, err := utils.SshSession(utils.DockerHost())
	if err != nil {
		log.Fatalf("Unable to connect to Docker utils.DockerHost(). Is Docker running? (%v)", err.Error())
	}

	log.Printf("Installing bootlocal.sh on Docker Host")
	remoteTmpFile := fmt.Sprintf("/tmp/%s", filepath.Base(file.Name()))
	err = scp.CopyPath(file.Name(), remoteTmpFile, session)
	utils.RunCommandWithDefaults(utils.DockerHost(), fmt.Sprintf("sudo cp %s %s", remoteTmpFile, "/var/lib/boot2docker/bootlocal.sh"))
	session.Close()

	file = utils.CreateTemplateTempFile(templatesMirrorDaemonShBytes, 0666)
	session, err = utils.SshSession(utils.DockerHost())
	if err != nil {
		log.Fatalf("Unable to connect to Docker utils.DockerHost(). Is Docker running? (%v)", err.Error())
	}

	log.Printf("Installing mirror-daemon.sh on Docker Host")
	remoteTmpFile = fmt.Sprintf("/tmp/%s", filepath.Base(file.Name()))
	err = scp.CopyPath(file.Name(), remoteTmpFile, session)
	utils.RunCommandWithDefaults(utils.DockerHost(), fmt.Sprintf("sudo cp %s %s", remoteTmpFile, "/var/lib/boot2docker/mirror-daemon.sh"))
	session.Close()

	log.Println("Downloading file sync utility (mirror)")
	utils.RunCommandWithDefaults(utils.DockerHost(), fmt.Sprintf("sudo /var/lib/boot2docker/bootlocal.sh start"))

	log.Println("Restarting Docker")
	utils.RunCommandWithDefaults(utils.DockerHost(), "sudo shutdown -r now")
	utils.WaitForNetwork("docker", utils.DockerHost())
	utils.WaitForNetwork("mirror", utils.MirrorHost())

	// Removing shared folders
	if utils.CheckSharedFolders(ui) {
		log.Println("Unmounting Virtualbox shared folders")
		utils.UnmountSharedFolders()
	}

	log.Println("Parity installed. Run 'parity run' to to get started!")
}
