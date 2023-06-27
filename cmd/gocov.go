package cmd

import (
	"flag"
	"fmt"
	"io/fs"
	"os"
	"strings"

	"github.com/slavsan/gocov/internal"
)

const (
	depthFlagDesc    = "report on files and directories of certain depth"
	noColorFlagDesc  = "disable color output"
	withFullPathDesc = "include the full path column in the output"
)

func Exec() {
	var (
		err     error
		command internal.Command
		args    []string
		config  = &internal.Config{
			Color: true,
		}
		reportDepth  int
		noColor      bool
		withFullPath bool

		reportCmd = flag.NewFlagSet("report", flag.ExitOnError)
	)

	reportCmd.IntVar(&reportDepth, "depth", 0, depthFlagDesc)
	reportCmd.IntVar(&reportDepth, "d", 0, depthFlagDesc)
	reportCmd.BoolVar(&noColor, "no-color", false, noColorFlagDesc)
	reportCmd.BoolVar(&withFullPath, "with-full-path", false, noColorFlagDesc)

	reportCmd.Usage = func() {
		_, _ = fmt.Fprintf(
			os.Stdout, strings.Join([]string{
				`Usage of report:`,
				`  -d, --depth int`,
				`      %s`,
				`  --no-color`,
				`      %s`,
				`  --with-full-path`,
				`      %s`,
				``,
			}, "\n"),
			depthFlagDesc, noColorFlagDesc, withFullPathDesc,
		)
	}

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
		err = reportCmd.Parse(os.Args[2:])
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "failed to parse args: %s", err.Error())
			printUsage()
			os.Exit(1)
		}
		config.Depth = reportDepth
		config.Color = !noColor
		config.WithFullPath = withFullPath
		args = reportCmd.Args()
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
