package command

import (
	"flag"
	"io/ioutil"
	"log"

	"strings"

	"github.com/mefellows/parity/config"
	app "github.com/mefellows/parity/parity"
	"github.com/mefellows/parity/utils"
)

// BuildCommand contains parameters required to configure the Parity runtime
type BuildCommand struct {
	Meta       config.Meta
	ConfigFile string
	Verbose    bool
}

// Run Parity
func (c *BuildCommand) Run(args []string) int {
	cmdFlags := flag.NewFlagSet("sync", flag.ContinueOnError)
	cmdFlags.Usage = func() { c.Meta.Ui.Output(c.Help()) }

	cmdFlags.BoolVar(&c.Verbose, "verbose", true, "Enable verbose output")
	cmdFlags.StringVar(&c.ConfigFile, "config", utils.DefaultParityConfigurationFile(), "Specifies the Parity configuration file path")

	if err := cmdFlags.Parse(args); err != nil {
		return 1
	}

	if !c.Verbose {
		log.SetOutput(ioutil.Discard)
	}

	c.Meta.Ui.Output("Building containers")
	parity := app.New(&config.Config{Ui: c.Meta.Ui, ConfigFile: c.ConfigFile})
	if err := parity.Build(); err != nil {
		c.Meta.Ui.Error(err.Error())
	}

	return 0
}

// Help text for the command
func (c *BuildCommand) Help() string {
	helpText := `
Usage: parity build [options]

  Builds and publishes your applications' Docker containers to a registry.

Options:

  --config                    Path to the configuration file. Defaults to ./parity.yml.
  --verbose                   Enable verbose logging.
`

	return strings.TrimSpace(helpText)
}

// Synopsis for the command
func (c *BuildCommand) Synopsis() string {
	return "Build and publish your applications' Docker images."
}
