package main

import (
	"fmt"
	"os"

	"github.com/mefellows/parity/command"
	_ "github.com/mefellows/parity/run"
	_ "github.com/mefellows/parity/sync"
	"github.com/mefellows/parity/version"
	"github.com/mitchellh/cli"
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
