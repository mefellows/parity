package command

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/mefellows/parity/config"
	"github.com/mefellows/parity/run"
)

// XCommand contains parameters required to configure the X Server Proxy
type XCommand struct {
	Meta       config.Meta
	ParityFile string
}

// Run Parity
func (c *XCommand) Run(args []string) int {
	cmdFlags := flag.NewFlagSet("x", flag.ContinueOnError)
	cmdFlags.Usage = func() { c.Meta.Ui.Output(c.Help()) }

	dir, _ := os.Getwd()
	var parityFile = fmt.Sprintf("%s/parity.yml", dir)
	cmdFlags.StringVar(&c.ParityFile, "config", parityFile, "Parity configuration file")

	if err := cmdFlags.Parse(args); err != nil {
		return 1
	}

	run.XServerProxy()

	return 0
}

// Help text for the command
func (c *XCommand) Help() string {
	helpText := `
Usage: parity x [options]

	Creates a X Quartz window proxy

Options:

  --config                    Path to the configuration file. Defaults to parity.yml.
`

	return strings.TrimSpace(helpText)
}

// Synopsis for the command
func (c *XCommand) Synopsis() string {
	return "Attach to a running Docker process"
}
