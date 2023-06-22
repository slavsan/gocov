package internal

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Command int

const (
	Report Command = iota
	Check
	Inspect
	Test
	ConfigFile
)

const (
	NoColor = "\033[0m"
	Red     = "\033[0;31m"
	Green   = "\033[0;32m"
	Yellow  = "\033[0;33m"
)

type Config struct {
	Color bool
}

type GocovConfig struct {
	Ignore    []string `json:"ignore"`
	Threshold float64  `json:"threshold"`
	Contents  []byte
}

type covReportLine struct {
	StartLine       int
	StartColumn     int
	EndLine         int
	EndColumn       int
	StatementsCount int
	Hits            int
}

type covFile struct {
	Name          string
	Path          string
	AllStatements int
	Percent       float64
	Covered       int
	Lines         []*covReportLine
}

type Exiter interface {
	Exit(code int)
}

type ProcessExiter struct{}

func (p *ProcessExiter) Exit(code int) {
	os.Exit(code)
}

func (cmd *Cmd) loadConfig() *GocovConfig {
	f, err := cmd.fsys.Open(".gocov")
	if err != nil {
		return nil
	}
	defer func() { _ = f.Close() }()

	var buf bytes.Buffer
	tee := io.TeeReader(f, &buf)

	b, err := io.ReadAll(tee)
	if err != nil {
		_, _ = fmt.Fprintf(cmd.stderr, "internal error: failed to load .gocov config in memory: %s", err.Error())
		cmd.exiter.Exit(1)
	}

	var conf *GocovConfig
	err = json.Unmarshal(b, &conf)
	if err != nil {
		_, _ = fmt.Fprintf(cmd.stderr, "failed to parse .gocov config file: %s\n", err)
		cmd.exiter.Exit(1)
	}

	if conf != nil {
		conf.Contents = buf.Bytes()
	}

	return conf
}

type Cmd struct {
	stdout io.Writer
	stderr io.Writer
	fsys   fs.FS
	config *Config
	exiter Exiter
}

func NewCommand(stdout io.Writer, stderr io.Writer, fsys fs.FS, config *Config, exiter Exiter) *Cmd {
	return &Cmd{
		stdout: stdout,
		stderr: stderr,
		fsys:   fsys,
		config: config,
		exiter: exiter,
	}
}

func parseCoverageFile(f io.Reader, stdout, stderr io.Writer, exiter Exiter, gocovConfig *GocovConfig, moduleDir string) (*Tree, map[string]*covFile) {
	var (
		colonIndex  int
		currentLine int
		all         int64
		covered     int64
		files       = map[string]*covFile{}
	)
	scanner := bufio.NewScanner(f)
	scanner.Scan() // skip the `mode` line
	currentLine++
	if line := scanner.Text(); !strings.HasPrefix(line, "mode: ") {
		panic("invalid coverage file")
	}
	for scanner.Scan() {
		currentLine++
		line := scanner.Text()
		colonIndex = strings.Index(line, ":")
		name := line[:colonIndex]

		if _, ok := files[name]; !ok {
			files[name] = &covFile{Name: name}
		}

		covLine, err := parseLine(line[colonIndex+1:])
		if err != nil {
			_, _ = fmt.Fprintf(stderr, "failed to parse coverage file on line %d\n", currentLine)
			exiter.Exit(1)
		}
		files[name].Lines = append(files[name].Lines, covLine)

		files[name].AllStatements += covLine.StatementsCount
		if covLine.Hits > 0 {
			files[name].Covered += covLine.StatementsCount
		}
		files[name].Percent = float64(files[name].Covered) * 100 / float64(files[name].AllStatements)
	}

	for _, f := range files {
		all += int64(f.AllStatements)
		covered += int64(f.Covered)
		f.Path = strings.TrimPrefix(f.Name, moduleDir+"/")
	}

	tree := NewTree(stdout)
	for _, v := range files {
		if isIgnored(v, gocovConfig) {
			continue
		}
		tree.Add(v.Path, v)
	}

	return tree, files
}

func (cmd *Cmd) Exec(command Command, args []string) {
	if command == Test {
		cmd.Test()
		return
	}

	var (
		f           fs.File
		err         error
		moduleDir   = filepath.Dir(getModule(cmd.fsys, cmd.stderr, cmd.exiter))
		gocovConfig = cmd.loadConfig()
	)

	if command == ConfigFile {
		cmd.Config(gocovConfig)
		return
	}

	f, err = cmd.fsys.Open("coverage.out")
	if err != nil {
		_, _ = fmt.Fprintf(cmd.stderr, "failed to open coverage.out: %s", err.Error())
		cmd.exiter.Exit(1)
	}
	defer func() { _ = f.Close() }()

	tree, files := parseCoverageFile(f, cmd.stdout, cmd.stderr, cmd.exiter, gocovConfig, moduleDir)
	fileMaxLen, stmtsMaxLen := tree.Accumulate()

	if command == Inspect {
		cmd.Inspect(args, files, moduleDir)
		return
	}

	if command == Report {
		cmd.Report(tree, cmd.config, fileMaxLen, stmtsMaxLen)
		return
	}

	if command == Check {
		cmd.Check(tree, gocovConfig)
		return
	}
}

func padPath(maxFileLen int, path string, indent int) string {
	return strings.Repeat(" ", maxFileLen-len(path)-(indent*2))
}

func getPercent(n *Node) float64 {
	return float64(n.covered) * 100 / float64(n.allStatements)
}

func (t *Tree) Add(path string, value *covFile) {
	t.Root.Add(path, value)
}

func isIgnored(f *covFile, config *GocovConfig) bool {
	if config == nil {
		return false
	}
	for _, ignore := range config.Ignore {
		if strings.HasPrefix(f.Path, ignore) {
			return true
		}
	}
	return false
}

func getModule(fsys fs.FS, stderr io.Writer, exiter Exiter) string {
	f, err := fsys.Open("go.mod")
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "failed to open go.mod file: %s", err.Error())
		exiter.Exit(1)
	}
	defer func() { _ = f.Close() }()
	scanner := bufio.NewScanner(f)
	scanner.Scan()
	line := scanner.Text()
	if !strings.HasPrefix(line, "module ") {
		panic("invalid go.mod file")
	}
	return strings.TrimPrefix(line, "module ")
}

func parseLine(line string) (*covReportLine, error) {
	var (
		covLine   = &covReportLine{}
		prevIndex = -1
		column    int
	)
	for index, c := range append([]byte(line), '\n') {
		if c != ' ' && c != '.' && c != ',' && c != '\n' {
			continue
		}
		value, err := strconv.Atoi(line[prevIndex+1 : index])
		if err != nil {
			return nil, fmt.Errorf("failed to parse line (%d) in coverage.out file: %w", index+1, err)
		}
		prevIndex = index

		switch column {
		case 0:
			covLine.StartLine = value
		case 1:
			covLine.StartColumn = value
		case 2:
			covLine.EndLine = value
		case 3:
			covLine.EndColumn = value
		case 4:
			covLine.StatementsCount = value
		case 5:
			covLine.Hits = value
		}

		column++
	}
	return covLine, nil
}

func digitsCount(num int) int {
	if num == 0 {
		return 1
	}
	var digits int
	for num != 0 {
		num /= 10
		digits++
	}
	return digits
}
