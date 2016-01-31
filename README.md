# Parity

Docker on OSX and Windows, as it was meant to be.

Docker is awesome, but it suffers from a few annoying drawbacks:

* Users on Mac OSX or Windows need to go by an intermediary - a virtual machine. This means code synchronisation is a pain, and constant rebuilds are necessary.
* Docker Machine requires you to manage multiple machines, but generally you only ever deal with one of them.
* It is too flexible, resulting in many different ways to achieve 'normal' things. Most people just want to build their app without having to worry about orchestrating containers.

## Goals

Simplify Docker for non LXC-native environments (Windows, Mac OSX) by:

* Automatically synchronising code into running containers - no rebuilds!
* Providing a simplified, automatic development and CI workflow for common application types/scenarios
* Automatically configuring Docker (including Docker Machine) for local development
* Ensuring environment [parity](http://12factor.net/dev-prod-parity) with CI and Production - all environments use Docker containers
* Allowing customisation to the workflow, or extension of the platform via a plugin interface

Additionally, Parity will not require any other external dependencies, except the Docker ecosystem.
