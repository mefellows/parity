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

type SetupConfig struct {
	ImageName         string
	Base              string
	CiImageName       string
	Ci                string
	Version           string
	Overwrite         bool
	TemplateSourceUrl string
	TemplateIndex     string
}

var defaultParityTemplate = []string{
	"Dockerfile",
	"Dockerfile.ci",
	"Dockerfile.dist",
	"docker-compose.yml",
	"docker-compose.yml.dev",
	"parity.yml",
}

func tempFile(reader io.Reader, templateData interface{}) (*os.File, error) {
	buffer := make([]byte, 8096)
	_, err := reader.Read(buffer)
	tmpl, err := template.New("").Parse(string(buffer))
	if err != nil {
		return nil, fmt.Errorf("Template failed:", err.Error())
	}
	file, _ := ioutil.TempFile("/tmp", "parity")
	file.Chmod(0655)

	err = tmpl.Execute(file, templateData)
	if err != nil {
		return nil, fmt.Errorf("Template failed:", err.Error())
	}

	return file, nil
}

func SetupParityProject(config SetupConfig) error {
	log.Stage("Setup Parity Project")

	// 1. Merge SetupConfig with Defaults -> need to create Base, Ci and Production image names
	dir, _ := os.Getwd()
	parityTemplate := defaultParityTemplate

	// Scan template index file
	if config.TemplateIndex != "" {
		parityTemplate = make([]string, 0)
		log.Step("Downloading template index: %s", config.TemplateIndex)

		resp, err := http.Get(config.TemplateIndex)
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
				url := fmt.Sprintf(`%s/%s`, config.TemplateSourceUrl, f)
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
