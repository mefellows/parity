package command

import (
	"flag"
	"fmt"
	"strings"

	"github.com/mefellows/parity/config"
	app "github.com/mefellows/parity/parity"
	"github.com/mefellows/parity/utils"
)

// InteractiveCommand contains parameters required to configure the Parity runtime
type InteractiveCommand struct {
	Meta       config.Meta
	Service    string
	ParityFile string
}

// Run Parity
func (c *InteractiveCommand) Run(args []string) int {
	cmdFlags := flag.NewFlagSet("sync", flag.ContinueOnError)
	cmdFlags.Usage = func() { c.Meta.Ui.Output(c.Help()) }

	cmdFlags.StringVar(&c.Service, "service", "web", "Service to shell into. Defaults to 'web'")
	cmdFlags.StringVar(&c.ParityFile, "config", utils.DefaultParityConfigurationFile(), "Parity configuration file")

	if err := cmdFlags.Parse(args); err != nil {
		return 1
	}

	c.Meta.Ui.Output("Starting interactive session")
	parity := app.New(&config.Config{ConfigFile: c.ParityFile})
	parity.LoadPlugins()

	if shell, err := parity.GetShellPlugin("compose"); err == nil {
		shell.Shell(app.ShellConfig{
			Service: c.Service,
		})
	} else {
		Ui.Error(fmt.Sprintf("Unable to shell into container: %s", err.Error()))
	}

	return 0
}

// Help text for the command
func (c *InteractiveCommand) Help() string {
	helpText := `
Usage: parity interactive [options]

	Shells into an interactive Docker

Options:

  --service                   The service in your compose file to shell into. Required.
  --config                    Path to the configuration file. Defaults to parity.yml.
  --composefile               Path to the compose file. Defaults to docker-compose.yml.
`

	return strings.TrimSpace(helpText)
}

// Synopsis for the command
func (c *InteractiveCommand) Synopsis() string {
	return "Shell into an interactive Docker terminal"
}
