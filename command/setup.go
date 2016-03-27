package command

import (
	"flag"
	"strings"

	"github.com/mefellows/parity/config"
	"github.com/mefellows/parity/setup"
	"github.com/mefellows/parity/utils"
)

type SetupCommand struct {
	Meta             config.Meta
	Template         string
	TemplateLocation string
	Hostname         string
	ConfigFile       string
}

func (c *SetupCommand) Run(args []string) int {
	cmdFlags := flag.NewFlagSet("install", flag.ContinueOnError)
	cmdFlags.Usage = func() { c.Meta.Ui.Output(c.Help()) }

	cmdFlags.StringVar(&c.Hostname, "template", "", "Specify a pre-built template to use")
	cmdFlags.StringVar(&c.Hostname, "templateLocation", "", "Specify a URL pointing to a Parity template")
	cmdFlags.StringVar(&c.ConfigFile, "config", utils.DefaultParityConfigurationFile(), "Enable verbose output")

	if err := cmdFlags.Parse(args); err != nil {
		return 1
	}

	c.Meta.Ui.Output("Setting up a new Parity project")
	setup.SetupParityProject(setup.SetupConfig{
		ImageName:         "test",
		Ci:                "test-ci",
		Base:              "base",
		Version:           "1.0.0",
		Overwrite:         true,
		TemplateIndex:     "https://raw.githubusercontent.com/mefellows/parity-rails/master/index.txt",
		TemplateSourceUrl: "https://raw.githubusercontent.com/mefellows/parity-rails/master",
	})

	return 0
}

func (c *SetupCommand) Help() string {
	helpText := `
Usage: parity setup [options]

  Install Parity as a local daemon and into the running Docker Machine

Options:

	--template                 Specify a template name e.g. rails
	--templateLocation         Specify a template location (e.g. https://github.com/mefellows/parity) to use as a starting point.
  --hostname                 Specify the host entry for ''--dns'.
`

	return strings.TrimSpace(helpText)
}

func (c *SetupCommand) Synopsis() string {
	return "Setup a new application based on a prototype"
}
