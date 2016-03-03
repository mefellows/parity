# Parity

Docker on OSX and Windows, as it was meant to be.

Docker is awesome, but it suffers from a few annoying drawbacks:

* Users on Mac OSX or Windows need to go by an intermediary - a virtual machine. This means code synchronisation is a pain, and constant rebuilds are necessary.
* Only certain directories are automatically shared into the vm (e.g. `/Users` on OSX)
* Docker Machine requires you to manage multiple machines, but generally you only ever deal with one of them.
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


## Development
go-bindata  --pkg install --o install/assets.go templates/
