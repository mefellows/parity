package command

import (
	"flag"
	"strings"

	"github.com/mefellows/parity/config"
	"github.com/mefellows/parity/setup"
	"github.com/mefellows/parity/utils"
)

type SetupCommand struct {
	Meta              config.Meta
	Base              string
	Version           string
	Template          string
	TemplateSourceURL string
	ConfigFile        string
	Overwrite         bool
}

func (c *SetupCommand) Run(args []string) int {
	cmdFlags := flag.NewFlagSet("install", flag.ContinueOnError)
	cmdFlags.Usage = func() { c.Meta.Ui.Output(c.Help()) }

	cmdFlags.StringVar(&c.Base, "base", "", "Base image name e.g. 'my-awesome-project'")
	cmdFlags.StringVar(&c.Version, "version", "1.0.0", "Initial Docker image version")
	cmdFlags.StringVar(&c.Template, "template", "", "Specify a pre-built template to use")
	cmdFlags.StringVar(&c.TemplateSourceURL, "templateSourceUrl", "", "Specify a URL pointing to a Parity template")
	cmdFlags.StringVar(&c.ConfigFile, "config", utils.DefaultParityConfigurationFile(), "Enable verbose output")
	cmdFlags.BoolVar(&c.Overwrite, "force", false, "Overwrites any existing Parity files")

	if err := cmdFlags.Parse(args); err != nil {
		return 1
	}

	c.Meta.Ui.Output("Setting up a new Parity project")
	if err := setup.SetupParityProject(&setup.Config{
		TemplateSourceURL:  c.TemplateSourceURL,
		TemplateSourceName: c.Template,
		Version:            c.Version,
		Overwrite:          c.Overwrite,
		Base:               c.Base,
	}); err != nil {
		c.Meta.Ui.Error(err.Error())
		return 1
	}

	return 0
}

func (c *SetupCommand) Help() string {
	helpText := `
Usage: parity setup [options]

  Setup a new Parity project in the current working directory.

Options:

  --template                 Specify a template name e.g. "rails", "node".
  --templateSourceUrl        Specify a template location (e.g. https://github.com/mefellows/parity) to use as a starting point.
  --base                     Base docker image name e.g. my-awesome-project.
  --version                  Docker container version.
  --force                    Overwrite any existing files when expanding the template.
  --config                   Path to the configuration file. Defaults to parity.yml.
`

	return strings.TrimSpace(helpText)
}

func (c *SetupCommand) Synopsis() string {
	return "Setup a new application based on a prototype"
}
