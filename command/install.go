package command

import (
	"flag"
	"fmt"
	"github.com/mefellows/parity/install"
	"strings"
)

type InstallCommand struct {
	Meta Meta
	Port int // Which port to listen on
}

func (c *InstallCommand) Run(args []string) int {
	cmdFlags := flag.NewFlagSet("install", flag.ContinueOnError)
	cmdFlags.Usage = func() { c.Meta.Ui.Output(c.Help()) }

	cmdFlags.IntVar(&c.Port, "port", 8123, "The http port to listen on")

	// Validate
	if err := cmdFlags.Parse(args); err != nil {
		return 1
	}

	c.Meta.Ui.Output(fmt.Sprintf("Installing parity on port %d", c.Port))
	install.InstallParity()
	// install.CreateBoot2DockerDaemon()
	// _, err := install.RunCommand("docker:22", "ls -larth")
	// fmt.Printf("Error: %v", err)

	return 0
}

func (c *InstallCommand) Help() string {
	helpText := `
Usage: parity install [options]

  Install Parity as a local daemon and into the running Docker Machine

Options:

  --port                      The http(s) port to listen on
`

	return strings.TrimSpace(helpText)
}

func (c *InstallCommand) Synopsis() string {
	return "Run the parity install"
}
