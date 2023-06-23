package internal

import (
	"fmt"
	"io"
	"sort"
	"strings"
)

func (cmd *Cmd) Inspect(args []string, files map[string]*covFile, moduleDir string) {
	if len(args) < 1 {
		_, _ = fmt.Fprintf(cmd.stderr, "no arguments provided to inspect command\n")
		cmd.exiter.Exit(1)
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
		_, _ = fmt.Fprintf(cmd.stderr, "failed to open %s", moduleDir+"/"+relPath)
		cmd.exiter.Exit(1)
		return
	}
	f2, err := cmd.fsys.Open(targetFile)
	if err != nil {
		_, _ = fmt.Fprintf(cmd.stderr, "failed to open file to inspect: %s", err.Error())
		cmd.exiter.Exit(1)
		return
	}
	defer func() { _ = f2.Close() }()

	data, err := io.ReadAll(f2)
	if err != nil {
		_, _ = fmt.Fprintf(cmd.stderr, "failed to read target file to inspect: %s", err.Error())
		cmd.exiter.Exit(1)
		return
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
		_, _ = fmt.Fprintf(cmd.stdout, "%d| %s\n", num+1, line)
	}
}
