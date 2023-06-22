package internal

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
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

func loadConfig(fsys fs.FS, stderr io.Writer, exiter Exiter) *GocovConfig {
	f, err := fsys.Open(".gocov")
	if err != nil {
		return nil
	}
	defer func() { _ = f.Close() }()

	var buf bytes.Buffer
	tee := io.TeeReader(f, &buf)

	b, err := io.ReadAll(tee)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "internal error: failed to load .gocov config in memory: %s", err.Error())
		exiter.Exit(1)
	}

	var conf *GocovConfig
	err = json.Unmarshal(b, &conf)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "failed to parse .gocov config file: %s\n", err)
		exiter.Exit(1)
	}

	if conf != nil {
		conf.Contents = buf.Bytes()
	}

	return conf
}

func testCommand(stdout, stderr io.Writer, exiter Exiter) {
	coverArgs := []string{"test", "-coverprofile", "coverage.out", "-coverpkg", "./...", "./..."}
	_, _ = fmt.Fprintf(stdout, "executing: go %s\n", strings.Join(coverArgs, " "))
	cmd := exec.Command("go", coverArgs...)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	err := cmd.Run()
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "failed to run `go test` command: %s", err.Error())
		exiter.Exit(1)
	}
}

func config2(gocovConfig *GocovConfig, stdout io.Writer) {
	if gocovConfig != nil {
		_, _ = fmt.Fprintf(stdout, "%s\n", gocovConfig.Contents)
		return
	}
	_, _ = fmt.Fprintf(stdout, strings.Join([]string{ //nolint:staticcheck // SA1006
		`{`,
		`  "threshold": 50,`,
		`  "ignore": [`,
		`  ]`,
		`}`,
		``,
	}, "\n"))
}

type Cmd struct {
	// ..
}

func NewCommand() *Cmd {
	return &Cmd{
		// ..
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

func (cmd *Cmd) Exec(command Command, args []string, stdout io.Writer, stderr io.Writer, fsys fs.FS, config *Config, exiter Exiter) {
	if command == Test {
		testCommand(stdout, stderr, exiter)
		return
	}

	var (
		f           fs.File
		err         error
		moduleDir   = filepath.Dir(getModule(fsys, stderr, exiter))
		gocovConfig = loadConfig(fsys, stderr, exiter)
	)

	if command == ConfigFile {
		config2(gocovConfig, stdout)
		return
	}

	f, err = fsys.Open("coverage.out")
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "failed to open coverage.out: %s", err.Error())
		exiter.Exit(1)
	}
	defer func() { _ = f.Close() }()

	tree, files := parseCoverageFile(f, stdout, stderr, exiter, gocovConfig, moduleDir)
	fileMaxLen, stmtsMaxLen := tree.Accumulate()

	if command == Inspect {
		inspect(args, stdout, stderr, fsys, files, moduleDir, exiter)
		return
	}

	if command == Report {
		printReport(tree, config, fileMaxLen, stmtsMaxLen)
		return
	}

	if command == Check {
		check(stderr, tree, gocovConfig, exiter)
		return
	}
}

func check(stderr io.Writer, tree *Tree, gocovConfig *GocovConfig, exiter Exiter) {
	actualCoveragePercent := float64(tree.Root.covered) * 100 / float64(tree.Root.allStatements)
	if gocovConfig == nil {
		_, _ = fmt.Fprintf(stderr, "Coverage check failed: .gocov file with threshold needs to be set\n")
		exiter.Exit(1)
		return
	}
	if actualCoveragePercent < gocovConfig.Threshold {
		_, _ = fmt.Fprintf(stderr, "Coverage check failed: expected to have %.2f coverage, but got %.2f\n", gocovConfig.Threshold, actualCoveragePercent)
		exiter.Exit(1)
	}
}

type Tree struct {
	Root   *Node
	writer io.Writer
}

func NewTree(w io.Writer) *Tree {
	return &Tree{
		writer: w,
		Root:   &Node{path: "root", children: map[string]*Node{}},
	}
}

func (t *Tree) Render(config *Config, fileMaxLen, stmtsMaxLen int) {
	w := t.writer
	_, _ = fmt.Fprintf(w, "|-%s-|-%s-|-%s-|\n", strings.Repeat("-", fileMaxLen), strings.Repeat("-", stmtsMaxLen+1), strings.Repeat("-", 8))
	_, _ = fmt.Fprintf(w, "| %-*s | %*s | %*s |\n", fileMaxLen, "File", stmtsMaxLen+1, "Stmts", 8, "% Stmts")
	_, _ = fmt.Fprintf(w, "|-%s-|-%s-|-%s-|\n", strings.Repeat("-", fileMaxLen), strings.Repeat("-", stmtsMaxLen+1), strings.Repeat("-", 8))

	sortOrder := make([]string, 0, len(t.Root.children))
	for k := range t.Root.children {
		sortOrder = append(sortOrder, k)
	}
	sort.Strings(sortOrder)

	for _, k := range sortOrder {
		c := t.Root.children[k]
		c.Render(w, config, 0, fileMaxLen, stmtsMaxLen)
	}
	_, _ = fmt.Fprintf(w, "|-%s-|-%s-|-%s-|\n", strings.Repeat("-", fileMaxLen), strings.Repeat("-", stmtsMaxLen+1), strings.Repeat("-", 8))
}

func (t *Tree) Accumulate() (int, int) {
	var fileMaxLen int
	var stmtsMaxLen int
	_, _, fileMaxLen, stmtsMaxLen = t.Root.Accumulate(0)
	return fileMaxLen, stmtsMaxLen
}

func (n *Node) Accumulate(indent int) (int, int, int, int) {
	var all, covered, maxPathLength, maxStmtsLength int
	if n.value != nil {
		all = n.value.AllStatements
		covered = n.value.Covered
	}
	for _, cn := range n.children {
		a, c, fileMaxLen, stmtsMaxLen := cn.Accumulate(indent + 1)
		all, covered = all+a, covered+c
		if fileMaxLen > maxPathLength {
			maxPathLength = fileMaxLen
		}
		if stmtsMaxLen > maxStmtsLength {
			maxStmtsLength = stmtsMaxLen
		}
	}
	n.allStatements = all
	n.covered = covered
	pathLength := (indent * 2) + len(n.path)
	if pathLength > maxPathLength {
		maxPathLength = pathLength
	}
	nodeStmtsLength := digitsCount(all) + digitsCount(covered)
	if nodeStmtsLength > maxStmtsLength {
		maxStmtsLength = nodeStmtsLength
	}
	return all, covered, maxPathLength, maxStmtsLength
}

func padPath(maxFileLen int, path string, indent int) string {
	return strings.Repeat(" ", maxFileLen-len(path)-(indent*2))
}

func (n *Node) Render(w io.Writer, config *Config, indent int, fileMaxLen int, stmtsMaxLen int) {
	percent := getPercent(n)
	color := Red
	noColorValue := NoColor
	if percent >= 80 {
		color = Green
	} else if percent >= 50 {
		color = Yellow
	}
	if !config.Color {
		color = ""
		noColorValue = ""
	}
	stmtsPadding := stmtsMaxLen - digitsCount(n.allStatements) - digitsCount(n.covered)
	_, _ = fmt.Fprintf(w,
		"|%s%s %s%s %s| %s%s%d/%d%s | %s%7.2f%%%s |\n",
		color, strings.Repeat("  ", indent), n.path, padPath(fileMaxLen, n.path, indent), noColorValue,
		color, strings.Repeat(" ", stmtsPadding), n.covered, n.allStatements, noColorValue,
		color, percent, noColorValue,
	)
	sortOrder := make([]string, 0, len(n.children))
	for k := range n.children {
		sortOrder = append(sortOrder, k)
	}
	sort.Strings(sortOrder)
	for _, k := range sortOrder {
		c := n.children[k]
		c.Render(w, config, indent+1, fileMaxLen, stmtsMaxLen)
	}
}

func getPercent(n *Node) float64 {
	return float64(n.covered) * 100 / float64(n.allStatements)
}

func (t *Tree) Add(path string, value *covFile) {
	t.Root.Add(path, value)
}

type Node struct {
	path          string
	value         *covFile
	allStatements int
	covered       int
	children      map[string]*Node
}

func (n *Node) Add(path string, value *covFile) {
	index := strings.IndexByte(path, '/')

	if index < 0 {
		if _, ok := n.children[path]; !ok {
			n.children[path] = &Node{path: path, value: value, children: map[string]*Node{}}
		}
		return
	}

	if _, ok := n.children[path[:index]]; !ok {
		n.children[path[:index]] = &Node{path: path[:index], children: map[string]*Node{}}
	}
	n.children[path[:index]].Add(path[index+1:], value)
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

func printReport(tree *Tree, config *Config, fileMaxLen, stmtsMaxLen int) {
	tree.Render(config, fileMaxLen, stmtsMaxLen)
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

func inspect(args []string, stdout, stderr io.Writer, fsys fs.FS, files map[string]*covFile, moduleDir string, exiter Exiter) {
	if len(args) < 1 {
		_, _ = fmt.Fprintf(stderr, "no arguments provided to inspect command\n")
		return
	}
	relPath := args[0]
	var targetFile string
	index := strings.IndexByte(relPath, '/')
	if index == -1 {
		targetFile = relPath
	} else {
		targetFile = relPath[index+1:]
	}
	file, ok := files[moduleDir+"/"+relPath]
	if !ok {
		_, _ = fmt.Fprintf(stderr, "failed to open %s", moduleDir+"/"+relPath)
		exiter.Exit(1)
		return
	}
	f2, err := fsys.Open(targetFile)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "failed to open file to inspect: %s", err.Error())
		exiter.Exit(1)
	}
	defer func() { _ = f2.Close() }()

	data, err := io.ReadAll(f2)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "failed to read target file to inspect: %s", err.Error())
		exiter.Exit(1)
	}

	lines := strings.Split(string(data), "\n")

	sort.Slice(file.Lines, func(i, j int) bool {
		if file.Lines[i].EndLine == file.Lines[j].EndLine {
			return file.Lines[i].StartColumn > file.Lines[j].StartColumn
		}
		return file.Lines[i].EndLine > file.Lines[j].EndLine
	})

	for _, x := range file.Lines {
		if x.Hits > 0 {
			continue
		}

		lineNum := x.EndLine - 1
		lines[lineNum] = lines[lineNum][:x.EndColumn-1] + NoColor + lines[lineNum][x.EndColumn-1:]

		lineNum = x.StartLine - 1
		lines[lineNum] = lines[lineNum][:x.StartColumn-1] + Red + lines[lineNum][x.StartColumn-1:]
	}

	for num, line := range lines {
		_, _ = fmt.Fprintf(stdout, "%d| %s\n", num+1, line)
	}
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
