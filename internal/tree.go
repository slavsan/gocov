package internal

import (
	"fmt"
	"io"
	"sort"
	"strings"
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
