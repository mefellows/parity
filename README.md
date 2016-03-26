# Parity

Docker on OSX and Windows, as it was meant to be.

Docker is awesome, but it suffers from a few annoying drawbacks:

1. Users on Mac OSX or Windows need to go by an intermediary - a virtual machine. This means code synchronisation is a pain, and constant rebuilds are necessary.
1. Only certain directories are automatically shared into the vm (e.g. `/Users` on OSX)
1. Docker Machine requires you to manage multiple machines, but generally you only ever deal with one of them. But we still have to configure with env vars or an annoying `eval`.
1. It is too flexible, resulting in many different ways to achieve 'normal' things. Most people just want to build their app without having to worry about orchestrating containers.

Parity addresses this shortcomings, simplifying Docker for local development.

*NOTE*: This project, in particular the file syncing problem, will only be partially superceded by [Docker](https://blog.docker.com/2016/03/docker-for-mac-windows-beta/) when it itself is out of beta. It will also replace items 1-3, leaving Parity to deal with the opinionated setup problem and improving the workflow for development. This is great news for everyone!

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

### Using Go Get

```bash
go get github.com/mefellows/parity
```

## Installation

Simply run `parity install` and follow the prompts.

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

## Configuration File format

TODO

## Similar Projects

All of the below attempt to address the first 2 or 3 goals of this project, but not the fourth (opinionated build and local dev process). Additionally, they all bring in some other dependency and none of them work on Windows.

* https://github.com/brikis98/docker-osx-dev - This is my favourite and much of this project owes to the works of it. It doesn't work on Windows, however, and has a few small pre-requisites. If you're afraid of Parity, consider this!
* https://allysonjulian.com/setting-up-docker-with-xhyve/ and https://github.com/zchee/docker-machine-driver-xhyve - This is even more seamless, using a native containerisation system for MacOSX called [xhyve](https://github.com/mist64/xhyve) and an experimental Docker driver. Again, this doesn't work on Windows but is very good. One downside is that it needs to allocate the entire drive space, unlike typical vdockerirtualisation applications that will have a dynamically resizable volume.
* https://github.com/nlf/dlite - This aims to solve the top 3 goals of the project. I was never quite able to get this working, but it is quite an interesting project.
* See https://github.com/brikis98/docker-osx-dev#alternatives for others.
