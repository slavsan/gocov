package internal

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log"
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

type ProcessExiter struct {
}

func (p *ProcessExiter) Exit(code int) {
	os.Exit(code)
}

func loadConfig(fsys fs.FS) *GocovConfig {
	f, err := fsys.Open(".gocov")
	if err != nil {
		return nil
	}
	defer func() { _ = f.Close() }()

	b, err := io.ReadAll(f)
	check(err)

	var conf *GocovConfig
	err = json.Unmarshal(b, &conf)
	check(err)

	return conf
}

func Exec(command Command, args []string, stdout io.Writer, stderr io.Writer, fsys fs.FS, config *Config, exiter Exiter) {
	if command == Test {
		coverArgs := []string{"go", "test", "-coverprofile", "coverage.out", "-coverpkg", "./...", "./..."}
		fmt.Printf("executing: %s\n", strings.Join(coverArgs, " "))
		cmd := exec.Command(coverArgs[0], coverArgs[1:]...)
		cmd.Stdout = stdout
		cmd.Stderr = stderr
		err := cmd.Run()
		check(err)
		return
	}

	var (
		f           fs.File
		err         error
		colonIndex  int
		currentLine int
		all         int64
		covered     int64
		moduleDir   = filepath.Dir(getModule(fsys))
		files       = map[string]*covFile{}
		gocovConfig = loadConfig(fsys)
	)

	f, err = fsys.Open("coverage.out")
	check(err)
	defer func() { _ = f.Close() }()

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
			log.Fatalf("failed to parse coverage file on line %d\n", currentLine)
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
	fileMaxLen, stmtsMaxLen := tree.Accumulate()

	if command == Inspect {
		if len(args) < 1 {
			fmt.Fprintf(stdout, "no arguments provided to inspect command\n")
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
			panic("file not found")
		}
		f2, err := fsys.Open(targetFile)
		check(err)
		defer func() { _ = f2.Close() }()

		data, err := io.ReadAll(f2)
		check(err)

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
		return
	}

	if command == Report {
		printReport(tree, config, fileMaxLen, stmtsMaxLen)
		return
	}

	if command == Check {
		actualCoveragePercent := float64(tree.Root.covered) * 100 / float64(tree.Root.allStatements)
		if gocovConfig == nil {
			_, _ = fmt.Fprintf(stderr, "Coverage check failed: .gocov file with threshold needs to be set\n")
			exiter.Exit(1)
		}
		if actualCoveragePercent < gocovConfig.Threshold {
			_, _ = fmt.Fprintf(stderr, "Coverage check failed: expected to have %.2f coverage, but got %.2f\n", gocovConfig.Threshold, actualCoveragePercent)
			exiter.Exit(1)
		}
		return
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
	fmt.Fprintf(w, "|-%s-|-%s-|-%s-|\n", strings.Repeat("-", fileMaxLen), strings.Repeat("-", stmtsMaxLen+1), strings.Repeat("-", 8))
	fmt.Fprintf(w, "| %-*s | %*s | %*s |\n", fileMaxLen, "File", stmtsMaxLen+1, "Stmts", 8, "% Stmts")
	fmt.Fprintf(w, "|-%s-|-%s-|-%s-|\n", strings.Repeat("-", fileMaxLen), strings.Repeat("-", stmtsMaxLen+1), strings.Repeat("-", 8))

	sortOrder := make([]string, 0, len(t.Root.children))
	for k := range t.Root.children {
		sortOrder = append(sortOrder, k)
	}
	sort.Strings(sortOrder)

	for _, k := range sortOrder {
		c := t.Root.children[k]
		c.Render(w, config, 0, fileMaxLen, stmtsMaxLen)
	}
	fmt.Fprintf(w, "|-%s-|-%s-|-%s-|\n", strings.Repeat("-", fileMaxLen), strings.Repeat("-", stmtsMaxLen+1), strings.Repeat("-", 8))
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

func padPath(w io.Writer, maxFileLen int, path string, indent int) string {
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
	fmt.Fprintf(w,
		"|%s%s %s%s %s| %s%s%d/%d%s | %s%7.2f%%%s |\n",
		color, strings.Repeat("  ", indent), n.path, padPath(w, fileMaxLen, n.path, indent), noColorValue,
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

func getModule(fsys fs.FS) string {
	f, err := fsys.Open("go.mod")
	check(err)
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
			return nil, err
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

func check(err error) {
	if err != nil {
		panic(err)
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
