package command

import (
	"flag"

	"strings"

	"github.com/mefellows/parity/config"
	"github.com/mefellows/parity/utils"
)

// CleanupCommand contains parameters required to cleanup Docker images/containers
type CleanupCommand struct {
	Meta       config.Meta
	ConfigFile string
}

// Run Parity
func (c *CleanupCommand) Run(args []string) int {
	cmdFlags := flag.NewFlagSet("sync", flag.ContinueOnError)
	cmdFlags.Usage = func() { c.Meta.Ui.Output(c.Help()) }

	cmdFlags.StringVar(&c.ConfigFile, "config", utils.DefaultParityConfigurationFile(), "Specifies the Parity configuration file path")

	if err := cmdFlags.Parse(args); err != nil {
		return 1
	}

	c.Meta.Ui.Output("Cleaning up Docker images and containers")
	utils.Cleanup()

	return 0
}

// Help text for the command
func (c *CleanupCommand) Help() string {
	helpText := `
Usage: parity cleanup [options]

  Removes dangling docker images and stopped containers.

Options:

  --config                    Path to the configuration file. Defaults to ./parity.yml.
`

	return strings.TrimSpace(helpText)
}

// Synopsis for the command
func (c *CleanupCommand) Synopsis() string {
	return "Cleans up Docker images and containers"
}
