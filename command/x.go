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
	Port       int
}

// Run Parity
func (c *XCommand) Run(args []string) int {
	cmdFlags := flag.NewFlagSet("x", flag.ContinueOnError)
	cmdFlags.Usage = func() { c.Meta.Ui.Output(c.Help()) }

	dir, _ := os.Getwd()
	var parityFile = fmt.Sprintf("%s/parity.yml", dir)
	cmdFlags.StringVar(&c.ParityFile, "config", parityFile, "Parity configuration file")
	cmdFlags.IntVar(&c.Port, "port", 6000, "Proxy port")

	if err := cmdFlags.Parse(args); err != nil {
		return 1
	}

	c.Meta.Ui.Output("Starting X Proxy")
	run.XServerProxy(c.Port)

	return 0
}

// Help text for the command
func (c *XCommand) Help() string {
	helpText := `
Usage: parity x [options]

	Creates a X Quartz window proxy for any docker container.

Options:

  --port		The X Server Proxy listener port. Defaults to 6000 (XQuartz default).
`

	return strings.TrimSpace(helpText)
}

// Synopsis for the command
func (c *XCommand) Synopsis() string {
	return "Attach to a running Docker process"
}
