package internal

import (
	"fmt"
	"io"
	"sort"
	"strings"
)

const (
	percentFillSymbol  = "\u25A0"
	percentEmptySymbol = " "
)

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

func (t *Tree) Render(config *Config, stats Stats, args []string) {
	w := t.writer
	_, _ = fmt.Fprintf(w, "|-%s-|-%s-|-%s-|-%s|", strings.Repeat("-", stats.FileMaxLen), strings.Repeat("-", stats.StmtsMaxLen+1), strings.Repeat("-", 8), strings.Repeat("-", 11))
	if config.WithFullPath {
		_, _ = fmt.Fprintf(w, "-%s-|", strings.Repeat("-", stats.FullPathMaxLen))
	}
	_, _ = fmt.Fprintf(w, "\n")
	_, _ = fmt.Fprintf(w, "| %-*s | %*s | %*s | %-*s |", stats.FileMaxLen, "File", stats.StmtsMaxLen+1, "Stmts", 8, "% Stmts", 10, "Progress")
	if config.WithFullPath {
		_, _ = fmt.Fprintf(w, " %-*s |", stats.FullPathMaxLen, "Full path")
	}
	_, _ = fmt.Fprintf(w, "\n")
	_, _ = fmt.Fprintf(w, "|-%s-|-%s-|-%s-|-%s-|", strings.Repeat("-", stats.FileMaxLen), strings.Repeat("-", stats.StmtsMaxLen+1), strings.Repeat("-", 8), strings.Repeat("-", 10))
	if config.WithFullPath {
		_, _ = fmt.Fprintf(w, "-%s-|", strings.Repeat("-", stats.FullPathMaxLen))
	}
	_, _ = fmt.Fprintf(w, "\n")

	sortOrder := make([]string, 0, len(t.Root.children))
	for k := range t.Root.children {
		sortOrder = append(sortOrder, k)
	}
	sort.Strings(sortOrder)

	for _, k := range sortOrder {
		c := t.Root.children[k]
		c.Render(w, config, stats, args)
	}
	_, _ = fmt.Fprintf(w, "|-%s-|-%s-|-%s-|-%s-|", strings.Repeat("-", stats.FileMaxLen), strings.Repeat("-", stats.StmtsMaxLen+1), strings.Repeat("-", 8), strings.Repeat("-", 10))
	if config.WithFullPath {
		_, _ = fmt.Fprintf(w, "-%s-|", strings.Repeat("-", stats.FullPathMaxLen))
	}
	_, _ = fmt.Fprintf(w, "\n")

	// fmt.Printf("/full/path/to/gocov/go.mod:1\n")
	// fmt.Printf("gocov/go.mod:1\n")
}

type Stats struct {
	All            int
	Covered        int
	FileMaxLen     int
	StmtsMaxLen    int
	FullPathMaxLen int
}

func (t *Tree) Accumulate() Stats {
	return t.Root.Accumulate()
}

func (n *Node) Accumulate() Stats {
	var all, covered, maxPathLength, maxStmtsLength, fullPathMaxLen int
	if n.value != nil {
		all = n.value.AllStatements
		covered = n.value.Covered
	}
	for _, cn := range n.children {
		stats := cn.Accumulate()
		all, covered = all+stats.All, covered+stats.Covered
		if stats.FileMaxLen > maxPathLength {
			maxPathLength = stats.FileMaxLen
		}
		if stats.StmtsMaxLen > maxStmtsLength {
			maxStmtsLength = stats.StmtsMaxLen
		}
		if stats.FullPathMaxLen > fullPathMaxLen {
			fullPathMaxLen = stats.FullPathMaxLen
		}
	}
	n.allStatements = all
	n.covered = covered
	pathLength := (n.level * 2) + len(n.path)
	if pathLength > maxPathLength {
		maxPathLength = pathLength
	}
	nodeStmtsLength := digitsCount(all) + digitsCount(covered)
	if nodeStmtsLength > maxStmtsLength {
		maxStmtsLength = nodeStmtsLength
	}
	fullPathLen := len(n.fullPath)
	if fullPathLen > fullPathMaxLen {
		fullPathMaxLen = fullPathLen
	}
	return Stats{
		All:            all,
		Covered:        covered,
		FileMaxLen:     maxPathLength,
		StmtsMaxLen:    maxStmtsLength,
		FullPathMaxLen: fullPathMaxLen,
	}
}

func getColor(percent float64) (string, string) {
	color := Red
	noColorValue := NoColor
	if percent >= 80 {
		color = Green
	} else if percent >= 50 {
		color = Yellow
	}
	return color, noColorValue
}

func (n *Node) Render(w io.Writer, config *Config, stats Stats, args []string) {
	if config.Depth != 0 && n.level > config.Depth {
		return
	}
	var filterBySelectedPath bool
	var found bool
	if len(args) > 0 {
		filterBySelectedPath = true
		for _, search := range args {
			if strings.HasPrefix(n.fullPath, search) {
				found = true
			}
		}
	}

	percent := getPercent(n)
	color, noColorValue := getColor(percent)
	if !config.Color {
		color = ""
		noColorValue = ""
	}
	stmtsPadding := stats.StmtsMaxLen - digitsCount(n.allStatements) - digitsCount(n.covered)
	if !filterBySelectedPath || found || n.level == 0 {
		_, _ = fmt.Fprintf(w,
			"|%s%s %s%s %s| %s%s%d/%d%s | %s%7.2f%%%s | %s%s%s |",
			color, strings.Repeat("  ", n.level), n.path, padPath(stats.FileMaxLen, n.path, n.level), noColorValue,
			color, strings.Repeat(" ", stmtsPadding), n.covered, n.allStatements, noColorValue,
			color, percent, noColorValue,
			color, strings.Repeat(percentFillSymbol, progressbar(percent))+strings.Repeat(percentEmptySymbol, 10-progressbar(percent)), noColorValue,
		)
		if config.WithFullPath {
			_, _ = fmt.Fprintf(w,
				" %s%-*s%s |",
				color, stats.FullPathMaxLen, n.fullPath, noColorValue,
			)
		}
		_, _ = fmt.Fprintf(w, "\n")
	}
	sortOrder := make([]string, 0, len(n.children))
	for k := range n.children {
		sortOrder = append(sortOrder, k)
	}
	sort.Strings(sortOrder)
	for _, k := range sortOrder {
		c := n.children[k]
		c.Render(w, config, stats, args)
	}
}

type Node struct {
	path          string
	fullPath      string
	value         *covFile
	allStatements int
	covered       int
	children      map[string]*Node
	level         int
}

func (n *Node) Add(path string, fullPath string, value *covFile, level int) {
	index := strings.IndexByte(path, '/')

	if index < 0 {
		if _, ok := n.children[path]; !ok {
			n.children[path] = &Node{level: level, path: path, fullPath: getFullPath(fullPath, level), value: value, children: map[string]*Node{}}
		}
		return
	}

	if _, ok := n.children[path[:index]]; !ok {
		n.children[path[:index]] = &Node{level: level, path: path[:index], fullPath: getFullPath(fullPath, level), children: map[string]*Node{}}
	}
	n.children[path[:index]].Add(path[index+1:], fullPath, value, level+1)
}

func progressbar(percent float64) int {
	return int(percent / 10)
}

func getFullPath(path string, level int) string {
	parts := strings.Split(path, "/")
	return strings.Join(parts[:level+1], "/")
}
