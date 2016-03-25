package install

import (
	"os"

	"github.com/mefellows/parity/log"
	"github.com/mefellows/parity/utils"
)

// Init creates a default parity.yml file in the current dir
func Init() {
	log.Stage("Initialising Parity")
	log.Step("Creating 'parity.yml'")
	type FileTemplate struct {
		Name string
	}
	templateData := FileTemplate{Name: "somefile"}

	// Create the install mirror daemon template
	file := utils.CreateTemplateTempFile(templatesParityYmlBytes, 0655, templateData)
	os.Rename(file.Name(), "parity.yml")

	log.Stage("Initialising Parity : Complete")
}
