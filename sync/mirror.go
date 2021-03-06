package sync

import (
	"fmt"
	"os"
	"os/signal"
	"regexp"

	"github.com/mefellows/parity/log"
	"github.com/mefellows/parity/parity"
	"github.com/mefellows/plugo/plugo"

	// Need these blank imports. Ideally we fix mirror to
	// auto-export these plugins?
	_ "github.com/mefellows/mirror/filesystem/fs"
	_ "github.com/mefellows/mirror/filesystem/remote"
	mutils "github.com/mefellows/mirror/filesystem/utils"
	pki "github.com/mefellows/mirror/pki"
	sync "github.com/mefellows/mirror/sync"
	"github.com/mefellows/parity/utils"
)

type Mirror struct {
	Dest         string
	Src          string
	Filters      []string
	Exclude      []string
	Verbose      bool
	pluginConfig *parity.PluginConfig
}

func init() {
	plugo.PluginFactories.Register(func() (interface{}, error) {
		return &Mirror{}, nil
	}, "mirror")
}

func (p *Mirror) Name() string {
	return "mirror"
}

func (p *Mirror) Sync() error {
	log.Stage("Synchronising source/dest folders")
	pkiMgr, err := pki.New()
	pkiMgr.Config.Insecure = true

	if err != nil {
		p.pluginConfig.Ui.Error(fmt.Sprintf("Unable to setup public key infrastructure: %s", err.Error()))
	}

	Config, err := pkiMgr.GetClientTLSConfig()
	if err != nil {
		p.pluginConfig.Ui.Error(fmt.Sprintf("%v", err))
	}

	// Removing shared folders
	if utils.CheckSharedFolders() {
		utils.UnmountSharedFolders()
	}

	// Read volumes for share/watching
	var volumes []string

	// Exclude non-local volumes (e.g. might want to mount a dir on the VM guest)
	for _, v := range utils.ReadComposeVolumes() {
		if _, err := os.Stat(v); err == nil {
			volumes = append(volumes, v)
		}
	}
	// Add PWD if nothing in compose
	dir, _ := os.Getwd()
	if len(volumes) == 0 {
		volumes = append(volumes, mutils.LinuxPath(dir))
	}

	pki.MirrorConfig.ClientTlsConfig = Config
	excludes := make([]regexp.Regexp, len(p.Exclude))
	for i, v := range p.Exclude {
		r, err := regexp.CompilePOSIX(v)
		if err == nil {
			excludes[i] = *r
		} else {
			log.Error("Error parsing Regex:", err.Error())
		}
	}

	options := &sync.Options{Exclude: excludes, Verbose: p.Verbose}

	// Sync and watch all volumes
	for _, v := range volumes {
		log.Step("Syncing contents of '%s' -> '%s'", v, fmt.Sprintf("mirror://%s%s", utils.MirrorHost(), v))
		err = sync.Sync(v, fmt.Sprintf("mirror://%s%s", utils.MirrorHost(), v), options)
		if err != nil {
			log.Error("Error during initial file sync: %v", err)
		}

		log.Step("Monitoring '%s' for changes", v)
		go sync.Watch(v, fmt.Sprintf("mirror://%s%s", utils.MirrorHost(), v), options)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, os.Kill)

	<-sigChan
	log.Debug("Interrupt received, shutting down")

	return nil
}

func (m *Mirror) Configure(c *parity.PluginConfig) {
	log.Debug("Configuring mirror sync plugin")
	m.pluginConfig = c
}

func (m *Mirror) Teardown() error {
	log.Debug("Tearing down mirror sync plugin")
	return nil
}
