package command

import (
	"flag"
	"fmt"
	_ "github.com/mefellows/mirror/filesystem/fs"
	_ "github.com/mefellows/mirror/filesystem/remote"
	pki "github.com/mefellows/mirror/pki"
	sync "github.com/mefellows/mirror/sync"
	"os"
	"strings"
)

type excludes []string

func (e *excludes) String() string {
	return fmt.Sprintf("%s", *e)
}

func (e *excludes) Set(value string) error {
	*e = append(*e, value)
	return nil
}

type RunCommand struct {
	Meta    Meta
	Dest    string
	Src     string
	Filters []string
	Exclude excludes
}

func (c *RunCommand) Run(args []string) int {
	cmdFlags := flag.NewFlagSet("sync", flag.ContinueOnError)
	cmdFlags.Usage = func() { c.Meta.Ui.Output(c.Help()) }

	dir, _ := os.Getwd()
	cmdFlags.StringVar(&c.Src, "src", dir, "The src location to copy from")
	cmdFlags.StringVar(&c.Dest, "dest", fmt.Sprintf("mirror://docker:8123%s", dir), "The destination location to copy the contents of 'src' to.")

	// Validate
	if err := cmdFlags.Parse(args); err != nil {
		return 1
	}

	pkiMgr, err := pki.New()
	pkiMgr.Config.Insecure = true

	if err != nil {
		c.Meta.Ui.Error(fmt.Sprintf("Unable to setup public key infrastructure: %s", err.Error()))
		return 1
	}

	config, err := pkiMgr.GetClientTLSConfig()
	if err != nil {
		c.Meta.Ui.Error(fmt.Sprintf("%v", err))
		return 1
	}
	pki.MirrorConfig.ClientTlsConfig = config
	c.Meta.Ui.Output(fmt.Sprintf("Syncing contents of '%s' -> '%s'", c.Src, c.Dest))

	err = sync.Sync(c.Src, c.Dest)
	if err != nil {
		c.Meta.Ui.Error(fmt.Sprintf("Error during initial file sync: %v", err))
		return 1
	}

	c.Meta.Ui.Output(fmt.Sprintf("Monitoring %s for changes and syncing to %s.", c.Src, c.Dest))
	sync.Watch(c.Src, c.Dest)

	return 0
}

func (c *RunCommand) Help() string {
	helpText := `
Usage: parity run [options] 

  Runs Parity file watcher
  
Options:

  --src                       The source directory from which to copy from. Defaults to current dir.
  --dest                      The destination directory from which to copy to. Defaults to mirror://docker:8123/<curdir>.
  --exclude                   A regular expression used to exclude files and directories that match. 
                              This is a special option that may be specified multiple times
  --watch                     Watch for changes in source directory and continuously sync to dest
`

	return strings.TrimSpace(helpText)
}

func (c *RunCommand) Synopsis() string {
	return "Run Parity file watcher"
}
