package command

import (
	"flag"
	"strings"

	"github.com/mefellows/parity/config"
	"github.com/mefellows/parity/install"
)

type InstallCommand struct {
	Meta     config.Meta
	Dns      bool
	Hostname string
}

func (c *InstallCommand) Run(args []string) int {
	cmdFlags := flag.NewFlagSet("install", flag.ContinueOnError)
	cmdFlags.Usage = func() { c.Meta.Ui.Output(c.Help()) }

	cmdFlags.BoolVar(&c.Dns, "dns", false, "Create a host entry to your Docker environment at 'parity.dev'")
	cmdFlags.StringVar(&c.Hostname, "hostname", "parity.dev", "Create a host entry to your Docker environment at 'parity.dev'")

	if err := cmdFlags.Parse(args); err != nil {
		return 1
	}

	c.Meta.Ui.Output("Installing Parity")
	install.InstallParity(install.InstallConfig{Dns: c.Dns, DevHost: c.Hostname})

	return 0
}

func (c *InstallCommand) Help() string {
	helpText := `
Usage: parity install [options]

  Install Parity as a local daemon and into the running Docker Machine

Options:

  --dns                      Create a host entry to your Docker environment at 'parity.dev'
  --hostname                 Specify the host entry for ''--dns'.
`

	return strings.TrimSpace(helpText)
}

func (c *InstallCommand) Synopsis() string {
	return "Install Parity"
}
