# Parity

Docker on OSX and Windows, as it was meant to be.

Docker is awesome, but it suffers from a few annoying drawbacks:

1. Users on Mac OSX or Windows need to go by an intermediary - a virtual machine. This means code synchronisation is a pain, and constant rebuilds are necessary.
1. Only certain directories are automatically shared into the vm (e.g. `/Users` on OSX)
1. Docker Machine requires you to manage multiple machines, but generally you only ever deal with one of them. But we still have to configure with env vars or an annoying `eval`.
1. It is too flexible, resulting in many different ways to achieve 'normal' things. Most people just want to build their app without having to worry about orchestrating containers.

Parity addresses this shortcomings, simplifying Docker for local development.

*NOTE*: This project, in particular the file syncing problem, will only be partially superceded by [Docker](https://blog.docker.com/2016/03/docker-for-mac-windows-beta/) when it itself is out of beta. It will also replace items 1-3, leaving Parity to deal with the opinionated setup problem and improving the workflow for development. This is great news for everyone! As Parity is completely plugin-based, you can simple omit plugins (e.g. the `sync` plugin) that are superceded by Docker, with no change to your workflow.

[![wercker status](https://app.wercker.com/status/be9372da6e34efdf671fb7ebdea591ec/s "wercker status")](https://app.wercker.com/project/bykey/be9372da6e34efdf671fb7ebdea591ec)
[![Coverage Status](https://coveralls.io/repos/github/mefellows/parity/badge.svg?branch=master)](https://coveralls.io/github/mefellows/parity?branch=master)

#### Watch Parity in Action:
[![asciicast](https://asciinema.org/a/1ewj8cep4kcxrwj61vh5xzwm7.png)](https://asciinema.org/a/1ewj8cep4kcxrwj61vh5xzwm7)

## Goals

Simplify Docker for non LXC-native environments (Windows, Mac OSX) by:

* Automatically synchronising code into running containers - no rebuilds!
* Providing a simplified, automatic development and CI workflow for common application types/scenarios
* Automatically configuring Docker (including Docker Machine) for local development
* Ensuring environment [parity](http://12factor.net/dev-prod-parity) with CI and Production - all environments use Docker containers
* Allowing customisation to the workflow, or extension of the platform via a plugin interface

Additionally, Parity will not require any other external dependencies, except the Docker ecosystem.

## Project Status

*Beta*: The first 2 of 4 items are working, with the following features in a beta status:

* Simple _installation_ (`parity install`)
* _Code synchronisation_ into the Docker VM, from any directory, including file pattern exclusions (`parity run`)
* Automatically _run_ a docker compose file via Parity (`parity run`)
* Automatically _shell_ into a running service to look around (`parity interactive`)
* Automatically _attach_ into a running service to look around (`parity attach`)
* Windows support - see the [Windows node example](examples/node-windows). Currently, you need to provide full context paths. We plan on submitting patches to [libcompose](https://github.com/docker/libcompose) to move this into upstream and make it simpler.

## Getting Started

[Download](releases) a Parity release and put it somewhere on your `PATH`.

### On Mac OSX using Homebrew

```bash
brew install https://raw.githubusercontent.com/mefellows/parity/spike/scripts/parity.rb
```

### Download a release

Grab the latest [release](/mefellows/parity/releases) and put it somewhere on your `PATH`.

### Using Go Get

```bash
go get github.com/mefellows/parity
```

## Installation

Ensure the usual Docker [environment variables](https://docs.docker.com/machine/get-started/#create-a-machine) are exported, then simply run `parity install` and follow the prompts.

### Creating default host entry

To create a default host entry for http://parity.local (e.g. `/etc/hosts`) you can run the Parity installer with the `--dns` flag enabled:

On MacOSX:
```
sudo -E parity install --dns true
```

On Windows, run from an elevated PowerShell prompt.

Note: You will need elevated privileges to perform this function.

## Running

A typical invocation would look something like this:

```
parity init # Will create a default, sane parity.yml file
parity run"
```

Parity will then start up your Docker Services (in `./docker-compose.yml`) and synchronise files automatically into the Docker Virtual Machine.

By default, Parity will exclude any files containing `git`, `tmp` or ending with `.log`.

* `--config` - Path to the configuration file. Defaults to `./parity.yml`.
* `--verbose` - Enable verbose logging.

See `parity run --help` for more detail.


### Enabling GUI

*NOTE*: _This is a MacOSX only feature._

You will need to install XQuartz (`brew install Caskroom/cask/xquartz` or see https://xquartz.macosforge.org/trac for details).

First, ensure your X Server is running:

```
open -a XQuartz
```

From within a Parity project, you can simply run `parity run` and an proxy between Docker and your X Server will automatically be setup for you,
including the `$DISPLAY` environment variables in your Docker containers.

If you just want to setup a proxy for another non-Parity managed project, you can run `parity x`. This will setup create the Proxy as per above, but
you'll need to manually setup the `$DISPLAY variable`. Parity will log to console the environment variable setup. It will look something like `export DISPLAY=192.168.99.1:0`.

## Scaffolding projects

If you are starting a brand new project, you might like to opt for Parity's opinionated workflow, which enforces Docker and continuous delivery best practices.

Parity current has templates for Rails and Node, and you can get started with the `setup` command:

```
parity setup --template rails --base my-project2
```

## Configuration File format

```yaml
## The Project Name
name: My Awesome Project

## Log Level (0 = Trace, 1 = Debug, 2 = Info, 3 = Warn, 4 = Error, 5 = Fatal)
loglevel: 2

## Plugin configuration.
##
## Parity is essentially a wrapper for Plugins. You can use as much or as little
## as you need. e.g. If you don't need to sync files, simple remove the 'sync' plugin.

## Runtime plugin configuration
##
## Configures the Docker Compose runner
run:
  - name: compose
    config:
      composefile: .parity/docker-compose.yml.dev

## File synchronisation plugin configuration.
##
## Configures the file synchronisation Plugin, using Mirror (https://github.com/mefellows/mirror) by default
## This will eventually be superceded by Docker, when native virtualisation comes to OSX + Windows
sync:
  - name: mirror
    config:
      verbose: false
      exclude:
        - tmp
        - \.log$
        - \.git

## Shell plugin: Enables shelling into an Interactive Docker terminal.
##
## This Plugin allows us to shell into an Interactive terminal via Docker Compose
shell:
  - name: compose

## Docker Registry plugin configuration.
##
## Configures the location images are retrieved from/pushed to.
registry:
  - name: default
    config:
      host: parity.local:5000

## Docker Image Builder plugin configuration.  
##
## Configures how images are built and pushed to a Registry.
builder:
  - name: compose
    config:
      - image_name: parity-test

```

## Parity Templates

Templates exist for the following language/frameworks:

* [Rails](https://github.com/mefellows/parity-rails)
* [Node](https://github.com/mefellows/parity-node)

If you create your own (see below), you can have Parity scaffold your project as follows:

```
parity setup --templateSourceUrl=https://raw.githubusercontent.com/mefellows/parity-my-awesome-lang/master --base my-project2
```

### Creating your own Templates

Parity Templates must adhere to a specific pattern and must be internet accessible. The easiest way to go is creating a public GitHub repository, with the following layout:

```
├── .parity                  Contains template configuration files (e.g. DB init scripts etc.)
│   ├── TEMPLATE SPECIFIC CONFIGURATION FILES
├── Dockerfile               The Base Docker Image.
├── Dockerfile.ci            The CI/Build Docker Image. Inherits from Base.
├── Dockerfile.dist          The Production Docker Image. Inherits from Base.
├── docker-compose.yml       Production Docker Compose setup.
├── docker-compose.yml.dev   Local development Docker Compose setup.
├── parity.yml               Pre-configured parity.yml file for the Template.
└── index.txt                A file containing a manifest of all files required in the template.
```

Additional files may be included, provided that are noted in the Parity Template manifest `/index.txt`. By convention these files should lie in the `./.parity` folder.

*Template expansion variables*

Within the Template files, you can use the following variables using the usual golang
[text/template expansion](https://golang.org/pkg/text/template/) rules e.g. `{{.Base}}:{{.Version}}`:

|  Variable  |       Description       |  Default         |    Example        |  Required  |
|------------|:------------------------|:-----------------|:------------------|:-----------|
| Base       | The Base docker image   | n/a              | awesome-proj      | yes        |
| Ci         | CI container image      | `{{.Base}}-ci`   | awesome-proj-ci   | no         |
| Dist       | Prod container image    | `{{.Base}}-dist` | awesome-proj-dist | no         |
| Version    | Application version     | latest           | 1.0.0             | yes        |


## Similar Projects

All of the below attempt to address the first 2 or 3 goals of this project, but not the fourth (opinionated build and local dev process). Additionally, they all bring in some other dependency and none of them work on Windows.

* https://github.com/brikis98/docker-osx-dev - This is my favourite and much of this project owes to the works of it. It doesn't work on Windows, however, and has a few small pre-requisites. If you're afraid of Parity, consider this!
* https://allysonjulian.com/setting-up-docker-with-xhyve/ and https://github.com/zchee/docker-machine-driver-xhyve - This is even more seamless, using a native containerisation system for MacOSX called [xhyve](https://github.com/mist64/xhyve) and an experimental Docker driver. Again, this doesn't work on Windows but is very good. One downside is that it needs to allocate the entire drive space, unlike typical vdockerirtualisation applications that will have a dynamically resizable volume.
* https://github.com/nlf/dlite - This aims to solve the top 3 goals of the project. I was never quite able to get this working, but it is quite an interesting project.
* See https://github.com/brikis98/docker-osx-dev#alternatives for others.
