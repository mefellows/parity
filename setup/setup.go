package setup

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"text/template"

	"github.com/mefellows/parity/log"
)

// SetupConfig is the text/template structure used to expand out
// template values in the Parity Template file
type SetupConfig struct {
	ImageName          string
	Base               string
	Ci                 string
	Dist               string
	Version            string
	Overwrite          bool
	TemplateSourceName string
	TemplateSourceURL  string
	templateIndex      string
}

var defaultTemplates = map[string]bool{
	"node":  true,
	"rails": true,
}
var defaultTemplateUrlTemplate = "https://github.com/mefellows/parity-%s/raw/master"

func getDefaultTemplateUrl(templateName string) (string, error) {
	if defaultTemplates[templateName] {
		return fmt.Sprintf(defaultTemplateUrlTemplate, templateName), nil
	}
	return "", fmt.Errorf("Default template '%s' not found", templateName)
}

func tempFile(reader io.Reader, templateData interface{}) (*os.File, error) {
	buffer := make([]byte, 8096)
	i, err := reader.Read(buffer)
	tmpl, err := template.New("").Parse(string(buffer[:i]))
	if err != nil {
		return nil, fmt.Errorf("Template parsing failed:", err.Error())
	}
	file, _ := ioutil.TempFile("/tmp", "parity")
	file.Chmod(0655)

	err = tmpl.Execute(file, templateData)
	if err != nil {
		return nil, fmt.Errorf("Template failed:", err.Error())
	}

	return file, nil
}

// expandAndValidateConfig takes the initial SetupConfig and expands the other variables
// sets defaults etc.
func expandAndValidateConfig(config *SetupConfig) error {
	if config.Base == "" {
		return fmt.Errorf("Missing required 'Base' configuration item to the Parity Template")
	}
	if config.Version == "" {
		log.Warn("Missing required 'Version' configuration item to the Parity Template. Defaults to 'latest'")
		config.Version = "latest"
	}
	if config.TemplateSourceName == "" && config.TemplateSourceURL == "" {
		return fmt.Errorf("Must provide one of 'TemplateSourceName' or 'TemplateSourceURL' to the Parity Template")
	}
	if config.TemplateSourceName != "" && config.TemplateSourceURL != "" {
		return fmt.Errorf("Must provide only one of 'TemplateSourceName' or 'TemplateSourceURL' to the Parity Template")
	}
	if config.TemplateSourceURL != "" {
		config.templateIndex = fmt.Sprintf("%s/index.txt", config.TemplateSourceURL)
	}
	if config.TemplateSourceName != "" {
		url, err := getDefaultTemplateUrl(config.TemplateSourceName)
		if err == nil {
			config.TemplateSourceURL = url
			config.templateIndex = fmt.Sprintf("%s/index.txt", url)
		} else {
			return err
		}
	}
	if config.Ci == "" {
		config.Ci = fmt.Sprintf("%s-ci", config.Base)
	}
	if config.Dist == "" {
		config.Dist = fmt.Sprintf("%s-dist", config.Base)
	}
	return nil
}

func SetupParityProject(config *SetupConfig) error {
	log.Stage("Setup Parity Project")

	// 1. Merge SetupConfig with Defaults -> need to create Base, Ci and Production image names
	if err := expandAndValidateConfig(config); err != nil {
		return err
	}

	var parityTemplate []string

	// Scan template index file
	if config.templateIndex != "" {
		log.Step("Downloading template index: %s", config.templateIndex)

		resp, err := http.Get(config.templateIndex)
		if err != nil {
			return err
		}

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			parityTemplate = append(parityTemplate, scanner.Text())
		}
	}

	// Make .parity dir if not exists
	if _, err := os.Stat(".parity"); err == nil {
		os.Mkdir(".parity", 0755)
	}

	errors := make(chan error)
	done := make(chan bool)
	defer close(done)

	// Download and install templates in parallel
	go func() {
		wg := sync.WaitGroup{}
		wg.Add(len(parityTemplate))
		dir, _ := os.Getwd()

		for _, f := range parityTemplate {
			go func(f string, errorChan chan error) {
				// 1. Check presence of local File - overwrite? TODO: config
				targetFile := filepath.Join(dir, f)
				if _, err := os.Stat(targetFile); err == nil {
					if !config.Overwrite {
						errorChan <- fmt.Errorf("File '%s' already exists. Please specify --force to overwrite files.", targetFile)
						wg.Done()
						return
					}
				}

				// 2. Download resources
				// 2a. TODO: Check/store local cache?
				// 2b. Pull from remote
				url := fmt.Sprintf(`%s/%s`, config.TemplateSourceURL, f)
				log.Step("Downloading template file: %s", url)

				resp, err := http.Get(url)
				var file *os.File
				if err == nil {
					// 3. Interpolate template with Setup data
					if file, err = tempFile(resp.Body, config); err != nil {
						errorChan <- err
						wg.Done()
						return
					}
				} else {
					log.Error("Error downloading template: %s", err.Error())
					errorChan <- fmt.Errorf("Error downloading template: %s", err.Error())
					wg.Done()
					return
				}

				// 4. Move to local folder.
				log.Debug("Moving %s -> %s", file.Name(), targetFile)
				log.Debug("Ensuring parent dir exists: %s", filepath.Dir(targetFile))
				os.MkdirAll(filepath.Dir(targetFile), 0755)
				os.Rename(file.Name(), targetFile)
				wg.Done()
			}(f, errors)
		}

		wg.Wait()
		done <- true
	}()

	select {
	case e := <-errors:
		log.Error(e.Error())
		return e
	case <-done:
		log.Debug("Finished installing template")
	}

	log.Stage("Setup Parity Project : Complete")
	return nil
}
