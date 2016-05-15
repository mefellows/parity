package run

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestGenerateContainerVersion_NoFiles(t *testing.T) {
	c := &DockerCompose{}
	dir, _ := os.Getwd()
	res := c.generateContainerVersion(dir, "Dockerfile")
	if res != "" {
		t.Fatalf("Expected an empty MD5 hash, got '%s'", res)
	}
}

func TestGenerateContainerVersion_WithFiles(t *testing.T) {

	packageJsonContents := `{
    "name": "run",
    "version": "1.0.0",
    "description": "",
    "main": "index.js",
    "scripts": {
      "test": "echo \"Error: no test specified\" && exit 1"
    },
    "author": "",
    "license": "ISC",
    "dependencies": {
      "express": "^4.13.4"
    }
  }`

	dockerfileContents := `FROM node:5.2.0

  RUN mkdir -p /var/app/current
  WORKDIR /var/app/current
  COPY . /var/app/current/
  RUN NODE_ENV=development npm install
  RUN npm run package:dist
  RUN npm prune --production
  RUN rm -rf /var/app/current/lib
  ENV NODE_ENV production

  EXPOSE 80
  ENV PORT 80

  ENTRYPOINT ["node", "./dist/index.js"]`

	tempDir := "/tmp/parity-test"

	if _, err := os.Stat(tempDir); err != nil {
		os.Mkdir(tempDir, 0755)
	}
	ioutil.WriteFile("/tmp/parity-test/Dockerfile", []byte(dockerfileContents), 0655)
	ioutil.WriteFile("/tmp/parity-test/package.json", []byte(packageJsonContents), 0655)
	defer os.RemoveAll(tempDir)

	c := &DockerCompose{}
	res := c.generateContainerVersion(tempDir, "Dockerfile")
	if res != "fcc849bd02e7f688f1704e82e1c3751a" {
		t.Fatalf("Expected 'fcc849bd02e7f688f1704e82e1c3751a', got '%s'", res)
	}
}
