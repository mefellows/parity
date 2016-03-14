package command

import (
	"flag"
	"strings"

	"github.com/mefellows/parity/config"
	"github.com/mefellows/parity/install"
)

type InitCommand struct {
	Meta config.Meta
}

func (c *InitCommand) Run(args []string) int {
	cmdFlags := flag.NewFlagSet("install", flag.ContinueOnError)
	cmdFlags.Usage = func() { c.Meta.Ui.Output(c.Help()) }

	if err := cmdFlags.Parse(args); err != nil {
		return 1
	}

	c.Meta.Ui.Output("Initialising default Parity environment")
	install.Init()

	return 0
}

func (c *InitCommand) Help() string {
	helpText := `
Usage: parity init

  Creates a default parity.yml file in the current dir.
`

	return strings.TrimSpace(helpText)
}

func (c *InitCommand) Synopsis() string {
	return "Initialise Parity"
}
