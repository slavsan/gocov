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
	reportFileFlagDesc = "coverage profile file (default is coverage.out)"
	depthFlagDesc      = "report on files and directories of certain depth"
	noColorFlagDesc    = "disable color output"
	withFullPathDesc   = "include the full path column in the output"
	htmlOutputFlagDesc = "output the coverage in html format"
	// inspect flags.
	exactFlagDesc = "specify exact path to file"
	// check flags.
	thresholdFlagDesc = "specify the desired coverage threshold"
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
		reportFile   string
		reportDepth  int
		noColor      bool
		withFullPath bool
		exactPath    bool
		threshold    float64
		htmlOutput   bool

		reportCmd  = flag.NewFlagSet("report", flag.ExitOnError)
		checkCmd   = flag.NewFlagSet("check", flag.ExitOnError)
		inspectCmd = flag.NewFlagSet("inspect", flag.ExitOnError)
		configCmd  = flag.NewFlagSet("config", flag.ExitOnError)
	)

	reportCmd.StringVar(&reportFile, "file", "coverage.out", reportFileFlagDesc)
	reportCmd.StringVar(&reportFile, "f", "coverage.out", reportFileFlagDesc)
	reportCmd.IntVar(&reportDepth, "depth", 0, depthFlagDesc)
	reportCmd.IntVar(&reportDepth, "d", 0, depthFlagDesc)
	reportCmd.BoolVar(&noColor, "no-color", false, noColorFlagDesc)
	reportCmd.BoolVar(&withFullPath, "with-full-path", false, noColorFlagDesc)
	reportCmd.BoolVar(&htmlOutput, "html", false, htmlOutputFlagDesc)

	checkCmd.Float64Var(&threshold, "threshold", 0, thresholdFlagDesc)

	inspectCmd.BoolVar(&exactPath, "exact", false, noColorFlagDesc)

	reportCmd.Usage = func() {
		_, _ = fmt.Fprintf(
			os.Stdout, strings.Join([]string{
				`Usage of report:`,
				`  -f, --file string`,
				`      %s`,
				`  -d, --depth int`,
				`      %s`,
				`  --html`,
				`      %s`,
				`  --no-color`,
				`      %s`,
				`  --with-full-path`,
				`      %s`,
				``,
			}, "\n"),
			reportFileFlagDesc, depthFlagDesc, htmlOutputFlagDesc,
			noColorFlagDesc, withFullPathDesc,
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

	checkCmd.Usage = func() {
		_, _ = fmt.Fprintf(
			os.Stdout, strings.Join([]string{
				`Usage of check:`,
				`  --threshold int`,
				`      %s`,
				``,
			}, "\n"),
			thresholdFlagDesc,
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
		config.ReportFile = reportFile
		config.HTMLOutput = htmlOutput
		args = reportCmd.Args()
	case "test":
		command = internal.Test
	case "config":
		command = internal.ConfigFile
		err = configCmd.Parse(os.Args[2:])
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "failed to parse args: %s", err.Error())
			printUsage()
			os.Exit(1)
		}
		args = configCmd.Args()
	case "check":
		command = internal.Check
		err = checkCmd.Parse(os.Args[2:])
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "failed to parse args: %s", err.Error())
			printUsage()
			os.Exit(1)
		}
		config.Threshold = threshold
		args = checkCmd.Args()
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
		NewCommand(os.Stdout, os.Stderr, os.DirFS(".").(fs.StatFS), config, &internal.ProcessExiter{}, &internal.FileWriter{}). //nolint:forcetypeassert
		Exec(command, args)
}

const usage = `gocov - Go coverage reporting tool

  test     - run tests with coverage
  report   - print out a coverage report to stdout
  check    - check whether the defined coverage requirements are met
  inspect  - show the covered vs not covered statements in a file
  config   - print a default config or the current config if one is defined
  help     - show this help message
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
