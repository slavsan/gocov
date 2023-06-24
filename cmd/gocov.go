package cmd

import (
	"fmt"
	"io/fs"
	"os"

	"github.com/slavsan/gocov/internal"
)

func Exec() {
	var (
		command internal.Command
		args    []string
		config  = &internal.Config{
			Color: true,
		}
	)

	if len(os.Args) == 1 {
		printUsage()
		return
	}

	switch os.Args[1] {
	case "help":
		printUsage()
		return
	case "report":
		command = internal.Report
	case "test":
		command = internal.Test
	case "config":
		command = internal.ConfigFile
	case "check":
		command = internal.Check
	case "inspect":
		command = internal.Inspect
		if len(os.Args) > 2 {
			args = append(args, os.Args[2])
		}
	default:
		printUsage()
		return
	}

	internal.
		NewCommand(os.Stdout, os.Stderr, os.DirFS(".").(fs.StatFS), config, &internal.ProcessExiter{}). //nolint:forcetypeassert
		Exec(command, args)
}

const usage = `gocov

test		- run tests with coverage
report		- print out a coverage report to stdout
check		- check whether the defined coverage requirements are met
inspect		- show the covered vs not covered statements in a file
config		- print a default config or the current config if one is defined
help		- show this help message
`

func printUsage() {
	_, _ = fmt.Fprint(os.Stdout, usage)
}
