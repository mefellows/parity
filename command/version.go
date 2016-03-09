package command

import (
	"fmt"
	"strings"

	"github.com/mefellows/parity/version"
)

type VersionCommand struct {
}

func (c *VersionCommand) Run(args []string) int {
	fmt.Println(version.Version)
	return 0
}

func (c *VersionCommand) Help() string {
	helpText := `Usage: parity version`

	return strings.TrimSpace(helpText)
}

func (c *VersionCommand) Synopsis() string {
	return "Parity version"
}
