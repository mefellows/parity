package install

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/mefellows/parity/log"
	"github.com/mefellows/parity/utils"
	"github.com/mefellows/parity/version"
	"github.com/mitchellh/cli"
	"github.com/mitchellh/multistep"
	"github.com/tmc/scp"
)

type stepAdd struct{}

func (s *stepAdd) Run(state multistep.StateBag) multistep.StepAction {
	// Read our value and assert that it is they type we want
	value := state.Get("value").(int)
	fmt.Printf("Value is %d\n", value)

	// Store some state back
	state.Put("value", value+1)
	return multistep.ActionContinue
}

func (s *stepAdd) Cleanup(multistep.StateBag) {
	// This is called after all the steps have run or if the runner is
	// cancelled so that cleanup can be performed.
	log.Info("Cleaning up some step...")
}

func InstallParity(ui cli.Ui) {
	done := make(chan bool)
	// Interrupt handler
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, os.Kill)

	// Our "bag of state" that we read the value from
	state := new(multistep.BasicStateBag)
	state.Put("value", 0)

	steps := []multistep.Step{
		&stepAdd{},
		&stepAdd{},
		&stepAdd{},
	}

	runner := &multistep.BasicRunner{Steps: steps}

	go func() {
		// Executes the steps
		runner.Run(state)
		<-done
	}()

	select {
	case <-done:
		log.Info("Done stepping stuff")
	case <-sigChan:
		log.Info("Cancel signal arrived...")

		// Wrap this in an interruptable mechanism
		runner.Cancel()
	}

	log.Info("chan completed!")
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
	session, err := utils.SshSession(utils.DockerHost())
	if err != nil {
		log.Fatalf("Unable to connect to Docker utils.DockerHost(). Is Docker running? (%v)", err.Error())
	}

	log.Info("Installing bootlocal.sh on Docker Host")
	remoteTmpFile := fmt.Sprintf("/tmp/%s", filepath.Base(file.Name()))
	err = scp.CopyPath(file.Name(), remoteTmpFile, session)
	utils.RunCommandWithDefaults(utils.DockerHost(), fmt.Sprintf("sudo cp %s %s", remoteTmpFile, "/var/lib/boot2docker/bootlocal.sh"))
	session.Close()

	file = utils.CreateTemplateTempFile(templatesMirrorDaemonShBytes, 0655, templateData)
	session, err = utils.SshSession(utils.DockerHost())
	if err != nil {
		log.Fatalf("Unable to connect to Docker utils.DockerHost(). Is Docker running? (%v)", err.Error())
	}

	log.Info("Installing mirror-daemon.sh on Docker Host")
	remoteTmpFile = fmt.Sprintf("/tmp/%s", filepath.Base(file.Name()))
	err = scp.CopyPath(file.Name(), remoteTmpFile, session)
	utils.RunCommandWithDefaults(utils.DockerHost(), fmt.Sprintf("sudo cp %s %s", remoteTmpFile, "/var/lib/boot2docker/mirror-daemon.sh"))
	session.Close()

	log.Info("Downloading file sync utility (mirror)")
	utils.RunCommandWithDefaults(utils.DockerHost(), fmt.Sprintf("sudo /var/lib/boot2docker/bootlocal.sh start"))

	log.Info("Restarting Docker")
	utils.RunCommandWithDefaults(utils.DockerHost(), "sudo shutdown -r now")
	utils.WaitForNetwork("docker", utils.DockerHost())
	utils.WaitForNetwork("mirror", utils.MirrorHost())

	// Removing shared folders
	if utils.CheckSharedFolders(ui) {
		log.Info("Unmounting Virtualbox shared folders")
		utils.UnmountSharedFolders()
	}

	log.Info("Parity installed. Run 'parity run' to to get started!")
}
