package command

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/mefellows/parity/config"
	app "github.com/mefellows/parity/parity"
)

// InteractiveCommand contains parameters required to configure the Parity runtime
type InteractiveCommand struct {
	Meta        config.Meta
	Service     string
	ComposeFile string
	ParityFile  string
}

// Run Parity
func (c *InteractiveCommand) Run(args []string) int {
	cmdFlags := flag.NewFlagSet("sync", flag.ContinueOnError)
	cmdFlags.Usage = func() { c.Meta.Ui.Output(c.Help()) }

	dir, _ := os.Getwd()
	var parityFile = fmt.Sprintf("%s/parity.yml", dir)
	var configFile = fmt.Sprintf("%s/docker-compose.yml", dir)

	cmdFlags.StringVar(&c.Service, "service", "web", "Service to shell into. Defaults to 'web'")
	cmdFlags.StringVar(&c.ComposeFile, "composefile", configFile, "Compose file")
	cmdFlags.StringVar(&c.ParityFile, "config", parityFile, "Parity configuration file")

	parity := app.New(&config.Config{ConfigFile: parityFile})
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
