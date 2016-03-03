package command

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"regexp"
	"strings"

	_ "github.com/mefellows/mirror/filesystem/fs"
	_ "github.com/mefellows/mirror/filesystem/remote"
	pki "github.com/mefellows/mirror/pki"
	sync "github.com/mefellows/mirror/sync"
	"github.com/mefellows/parity/install"
)

type excludes []regexp.Regexp

func (e *excludes) String() string {
	return fmt.Sprintf("%s", *e)
}

func (e *excludes) Set(value string) error {
	r, err := regexp.CompilePOSIX(value)
	if err != nil {
		*e = append(*e, *r)
	}
	return nil
}

type RunCommand struct {
	Meta    Meta
	Dest    string
	Src     string
	Filters []string
	Exclude excludes
	Verbose bool
}

func (c *RunCommand) Run(args []string) int {
	cmdFlags := flag.NewFlagSet("sync", flag.ContinueOnError)
	cmdFlags.Usage = func() { c.Meta.Ui.Output(c.Help()) }

	dir, _ := os.Getwd()
	cmdFlags.StringVar(&c.Src, "src", dir, "The src location to copy from")
	cmdFlags.StringVar(&c.Dest, "dest", fmt.Sprintf("%s%s", install.DockerHost(), dir), "The destination location to copy the contents of 'src' to.")
	cmdFlags.Var(&c.Exclude, "exclude", "Set of exclusions as POSIX regular expressions to exclude from the transfer")
	cmdFlags.BoolVar(&c.Verbose, "verbose", false, "Enable verbose output")

	// Validate
	if err := cmdFlags.Parse(args); err != nil {
		return 1
	}

	if !c.Verbose {
		log.SetOutput(ioutil.Discard)
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

	// Removing shared folders
	if install.CheckSharedFolders(c.Meta.Ui) {
		install.UnmountSharedFolders()
	}

	// Read volumes for share/watching
	volumes := make([]string, 0)

	// Exclude non-local volumes (e.g. might want to mount a dir on the VM guest)
	for _, v := range install.ReadComposeVolumes() {
		if _, err := os.Stat(v); err == nil {
			volumes = append(volumes, v)
		}
	}
	// Add PWD if nothing in compose
	if len(volumes) == 0 {
		volumes = append(volumes, dir)
	}

	pki.MirrorConfig.ClientTlsConfig = config

	options := &sync.Options{Exclude: c.Exclude}
	for _, v := range volumes {
		c.Meta.Ui.Output(fmt.Sprintf("Syncing contents of '%s' -> '%s'", v, fmt.Sprintf("mirror://%s/%s", install.MirrorHost(), v)))
		err = sync.Sync(v, fmt.Sprintf("mirror://%s/%s", install.MirrorHost(), v), options)
		if err != nil {
			c.Meta.Ui.Error(fmt.Sprintf("Error during initial file sync: %v", err))
			return 1
		}
	}

	for _, v := range volumes {
		c.Meta.Ui.Output(fmt.Sprintf("Monitoring '%s' for changes", v))
		go sync.Watch(v, fmt.Sprintf("mirror://%s/%s", install.MirrorHost(), v), options)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, os.Kill)

	<-sigChan
	c.Meta.Ui.Output("Received interrupt. Shutting down")

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
                              This is a special option that may be specified multiple times.
  --verbose                   Enable verbose logging.
`

	return strings.TrimSpace(helpText)
}

func (c *RunCommand) Synopsis() string {
	return "Run Parity file watcher"
}
