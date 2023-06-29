package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"strings"

	"github.com/slavsan/gocov/internal"
)

const (
	// report flags.
	depthFlagDesc    = "report on files and directories of certain depth"
	noColorFlagDesc  = "disable color output"
	withFullPathDesc = "include the full path column in the output"
	// inspect flags.
	exactFlagDesc = "specify exact path to file"
)

func Exec() { //nolint:funlen
	var (
		err     error
		command internal.Command
		args    []string
		config  = &internal.Config{
			Color:  true,
			Global: loadGlobalConf(),
		}
		reportDepth  int
		noColor      bool
		withFullPath bool
		exactPath    bool

		reportCmd  = flag.NewFlagSet("report", flag.ExitOnError)
		inspectCmd = flag.NewFlagSet("inspect", flag.ExitOnError)
	)

	reportCmd.IntVar(&reportDepth, "depth", 0, depthFlagDesc)
	reportCmd.IntVar(&reportDepth, "d", 0, depthFlagDesc)
	reportCmd.BoolVar(&noColor, "no-color", false, noColorFlagDesc)
	reportCmd.BoolVar(&withFullPath, "with-full-path", false, noColorFlagDesc)

	inspectCmd.BoolVar(&exactPath, "exact", false, noColorFlagDesc)

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

	inspectCmd.Usage = func() {
		_, _ = fmt.Fprintf(
			os.Stdout, strings.Join([]string{
				`Usage of inspect:`,
				`  --exact`,
				`      %s`,
				``,
			}, "\n"),
			exactFlagDesc,
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
		err = inspectCmd.Parse(os.Args[2:])
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "failed to parse args: %s", err.Error())
			printUsage()
			os.Exit(1)
		}
		config.ExactPath = exactPath
		args = inspectCmd.Args()
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

func loadGlobalConf() *internal.GocovConfig {
	homedir, err := os.UserHomeDir()
	if err != nil {
		// TODO: in debug mode, show the error
		return nil
	}

	pathToFile := path.Join(homedir, ".gocov")
	if _, err := os.Stat(pathToFile); errors.Is(err, os.ErrNotExist) {
		// TODO: in debug mode, show the error
		return nil
	}

	f, err := os.Open(pathToFile)
	if err != nil {
		// TODO: in debug mode, show the error
		return nil
	}
	defer func() { _ = f.Close() }()

	var buf bytes.Buffer
	tee := io.TeeReader(f, &buf)

	b, err := io.ReadAll(tee)
	if err != nil {
		// TODO: in debug mode, show the error
		return nil
	}

	var conf *internal.GocovConfig

	err = json.Unmarshal(b, &conf)
	if err != nil {
		// TODO: in debug mode, show the error
		return nil
	}

	if conf != nil {
		conf.Contents = buf.Bytes()
	}

	return conf
}
