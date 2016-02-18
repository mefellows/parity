package main

import (
	"fmt"
	"github.com/mefellows/parity/command"
	"github.com/mefellows/parity/version"
	"github.com/mitchellh/cli"
	"os"
)

func main() {
	cli := cli.NewCLI("parity", version.Version)
	cli.Args = os.Args[1:]
	cli.Commands = command.Commands

	exitStatus, err := cli.Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
	}
	os.Exit(exitStatus)
}
