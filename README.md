# Parity

Docker on OSX and Windows, as it was meant to be.

Docker is awesome, but it suffers from a few annoying drawbacks:

* Users on Mac OSX or Windows need to go by an intermediary - a virtual machine. This means code synchronisation is a pain, and constant rebuilds are necessary.
* Only certain directories are automatically shared into the vm (e.g. `/Users` on OSX)
* Docker Machine requires you to manage multiple machines, but generally you only ever deal with one of them. But I still have to configure with env vars or an annoying `eval`.
* It is too flexible, resulting in many different ways to achieve 'normal' things. Most people just want to build their app without having to worry about orchestrating containers.

[![wercker status](https://app.wercker.com/status/be9372da6e34efdf671fb7ebdea591ec/s "wercker status")](https://app.wercker.com/project/bykey/be9372da6e34efdf671fb7ebdea591ec)
[![Coverage Status](https://coveralls.io/repos/github/mefellows/parity/badge.svg?branch=master)](https://coveralls.io/github/mefellows/parity?branch=master)

## Goals

Simplify Docker for non LXC-native environments (Windows, Mac OSX) by:

* Automatically synchronising code into running containers - no rebuilds!
* Providing a simplified, automatic development and CI workflow for common application types/scenarios
* Automatically configuring Docker (including Docker Machine) for local development
* Ensuring environment [parity](http://12factor.net/dev-prod-parity) with CI and Production - all environments use Docker containers
* Allowing customisation to the workflow, or extension of the platform via a plugin interface

Additionally, Parity will not require any other external dependencies, except the Docker ecosystem.

## Getting Started

[Download](releases) a Parity release and put it somewhere on your `PATH`.

#### On Mac OSX using Homebrew

```bash
brew install https://raw.githubusercontent.com/mefellows/parity/master/scripts/parity.rb
```

#### Using Go Get

```bash
go get github.com/mefellows/parity
```

### Installation

Simple run `parity install` and follow the prompts.

### Running

A typical invocation would look something like this:

```
parity run --exclude "git" --exclude "tmp" --exclude "\.log$"
```

This will run the Parity file synchroniser tool, excluding any files containing `git`, `tmp` or ending with `.log`.

* `--src` - The source folder to sync from. Defaults to any volumes specified in a local `docker-compose.yml` or the `$PWD` if you're not using Compose.
* `--dest` - The destination folder to sync to. Defaults to any volumes specified in a local `docker-compose.yml` or the `$PWD` if you're not using Compose.
* `--exclude` - a POSIX regular expression to exclude when synchronising files. Can be specified multiple times.

See `parity run --help` for more detail.

### Configuring Parity with `.parityrc` files

If you are on a Mac and have installed the bash shims, whenever you enter a folder containing a `.parityrc` file, Parity will automatically run for you based on that configuration and begin syncing files.
Logs will be redirected to `/tmp/parity-<project>.log` should you wish to see what's happening.

#### File format

TODO

## Similar Projects

All of the below attempt to address the first 2 or 3 goals of this project, but not the fourth (opinionated build and local dev process) and they all bring in some other dependency.

* https://github.com/brikis98/docker-osx-dev - This is my favourite and much of this project owes to the works of it. It doesn't work on Windows, however, and has a few small pre-requisites. If you're afraid of Parity, consider this!
* https://allysonjulian.com/setting-up-docker-with-xhyve/ and https://github.com/zchee/docker-machine-driver-xhyve - This is even more seamless, using a native containerisation system for MacOSX called [xhyve](https://github.com/mist64/xhyve) and an experimental Docker driver. Again, this doesn't work on Windows but is very good.
* https://github.com/nlf/dlite - This aims to solve the top 3 goals of the project. I was never quite able to get this working, but it is quite an interesting project.
* See https://github.com/brikis98/docker-osx-dev#alternatives for others.
