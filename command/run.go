package command

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"

	"os"
	"strings"

	"github.com/mefellows/parity/config"
	app "github.com/mefellows/parity/parity"
)

// RunCommand contains parameters required to configure the Parity runtime
type RunCommand struct {
	Meta    config.Meta
	Verbose bool
}

// Run Parity
func (c *RunCommand) Run(args []string) int {
	cmdFlags := flag.NewFlagSet("sync", flag.ContinueOnError)
	cmdFlags.Usage = func() { c.Meta.Ui.Output(c.Help()) }

	var verbose bool
	dir, _ := os.Getwd()
	var configFile = fmt.Sprintf("%s/parity.yml", dir)
	cmdFlags.BoolVar(&verbose, "verbose", true, "Enable verbose output")
	cmdFlags.StringVar(&configFile, "config", configFile, "Enable verbose output")

	if !verbose {
		log.SetOutput(ioutil.Discard)
	}

	// TODO: Run should do the following

	// - Parse parity.yml file
	// - Automatically sync files
	// - Automatically run Docker/Compose
	parity := app.New(&config.Config{Ui: c.Meta.Ui, ConfigFile: configFile})
	parity.Run()

	return 0
}

// Help text for the command
func (c *RunCommand) Help() string {
	helpText := `
Usage: parity run [options]

  Runs Parity.

	By default, Parity will parse any local docker-compose.yml file, automatically sync the appropriate volumes
	and run your application.

Options:

  --config                    Path to the configuration file. Defaults to ./parity.yml.
  --verbose                   Enable verbose logging.
`

	return strings.TrimSpace(helpText)
}

// Synopsis for the command
func (c *RunCommand) Synopsis() string {
	return "Run Parity file watcher"
}
