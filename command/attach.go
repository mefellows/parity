package command

import (
	"flag"
	"fmt"
	"strings"

	"github.com/mefellows/parity/config"
	app "github.com/mefellows/parity/parity"
	"github.com/mefellows/parity/utils"
)

// AttachCommand contains parameters required to configure the Parity runtime
type AttachCommand struct {
	Meta       config.Meta
	Service    string
	ParityFile string
}

// Run Parity
func (c *AttachCommand) Run(args []string) int {
	cmdFlags := flag.NewFlagSet("sync", flag.ContinueOnError)
	cmdFlags.Usage = func() { c.Meta.Ui.Output(c.Help()) }

	cmdFlags.StringVar(&c.Service, "service", "web", "Service to shell into. Defaults to 'web'")
	cmdFlags.StringVar(&c.ParityFile, "config", utils.DefaultParityConfigurationFile(), "Parity configuration file")

	if err := cmdFlags.Parse(args); err != nil {
		return 1
	}

	c.Meta.Ui.Output("Attaching to running container")
	parity := app.New(&config.Config{ConfigFile: c.ParityFile})
	parity.LoadPlugins()

	if shell, err := parity.GetShellPlugin("compose"); err == nil {
		err := shell.Attach(app.ShellConfig{
			Service: c.Service,
		})
		if err != nil {
			Ui.Error(fmt.Sprintf("Unable to attach to container: %s", err.Error()))
		}
	} else {
		Ui.Error(fmt.Sprintf("Unable to attach to container: %s", err.Error()))
	}

	return 0
}

// Help text for the command
func (c *AttachCommand) Help() string {
	helpText := `
Usage: parity attach [options]

	Attaches to a running Docker container as PID: 1 (access to logs, debugger etc.)

Options:

  --service                   The service in your compose file to shell into. Defaults to 'web'.
  --config                    Path to the configuration file. Defaults to parity.yml.
`

	return strings.TrimSpace(helpText)
}

// Synopsis for the command
func (c *AttachCommand) Synopsis() string {
	return "Attach to a running Docker process"
}
