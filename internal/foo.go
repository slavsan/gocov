package internal

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	noColor = "\033[0m"
	red     = "\033[0;31m"
	green   = "\033[0;32m"
	yellow  = "\033[0;33m"
)

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

var files = map[string]*covFile{}

func Exec() {
	var (
		f           *os.File
		err         error
		colonIndex  int
		currentLine int
		all         int64
		covered     int64
		moduleDir   = filepath.Dir(getModule())
	)

	f, err = os.Open("./coverage.out")
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

	report(files)
}

type Tree struct {
	Root *Node
}

func NewTree() *Tree {
	return &Tree{
		Root: &Node{path: "root", children: map[string]*Node{}},
	}
}

func (t *Tree) Render() {
	for _, c := range t.Root.children {
		c.Render(0)
	}
}

func (t *Tree) Accumulate() {
	_, _ = t.Root.Accumulate()
}

func (n *Node) Accumulate() (int, int) {
	var all, covered int
	if n.value != nil {
		all = n.value.AllStatements
		covered = n.value.Covered
	}
	for _, cn := range n.children {
		a, c := cn.Accumulate()
		all, covered = all+a, covered+c
	}
	n.allStatements = all
	n.covered = covered
	return all, covered
}

func (n *Node) Render(indent int) {
	percent := getPercent(n)
	color := red
	if percent >= 80 {
		color = green
	} else if percent >= 50 {
		color = yellow
	}
	fmt.Printf("|%s%s %-30s | %10d/%10d | %8.2f%%%s |\n", color, strings.Repeat("  ", indent), n.path, n.covered, n.allStatements, percent, noColor)
	for _, c := range n.children {
		c.Render(indent + 1)
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

func report(f map[string]*covFile) {
	tree := NewTree()
	for _, v := range f {
		tree.Add(v.Path, v)
	}
	tree.Accumulate()
	tree.Render()
}

func getModule() string {
	f, err := os.Open("go.mod")
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
