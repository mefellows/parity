package command

import (
	"flag"
	"fmt"
	"github.com/mefellows/mirror/sync"
	"os"
	"strings"
)

type RunCommand struct {
	Meta Meta
	Port int // Which port to listen on
}

func (c *RunCommand) Run(args []string) int {
	cmdFlags := flag.NewFlagSet("run", flag.ContinueOnError)
	cmdFlags.Usage = func() { c.Meta.Ui.Output(c.Help()) }

	cmdFlags.IntVar(&c.Port, "port", 8123, "The http port to listen on")

	if err := cmdFlags.Parse(args); err != nil {
		return 1
	}

	// TODO: Interrupt signal handling
	c.Meta.Ui.Output("Runing Parity")
	dir, _ := os.Getwd()
	fmt.Printf("Dir: %s", dir)
	sync.Watch(dir, fmt.Sprintf("mirror://docker:8123%s", dir))

	return 0
}

func (c *RunCommand) Help() string {
	helpText := `
Usage: parity run [options]

  Run Parity!

Options:

  --port                      The http(s) port to listen on
`

	return strings.TrimSpace(helpText)
}

func (c *RunCommand) Synopsis() string {
	return "Run Parity"
}
