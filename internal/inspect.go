package internal

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"sort"
	"strings"
)

func (cmd *Cmd) findExactFile(args []string, files map[string]*covFile, moduleDir string) (*covFile, string, error) {
	var (
		targetFile string
		relPath    = args[0]
		index      = strings.IndexByte(relPath, '/')
	)
	if index == -1 {
		targetFile = relPath
	} else {
		targetFile = relPath[index+1:]
	}
	file, ok := files[moduleDir+"/"+relPath]
	if !ok {
		return nil, "", fmt.Errorf("failed to open %s", moduleDir+"/"+relPath) //nolint:goerr113
	}
	return file, targetFile, nil
}

func (cmd *Cmd) findPartialMatchFile(args []string, files map[string]*covFile, moduleDir string) (*covFile, string, int, error) {
	var (
		targetFile string
		skipped    int
	)
	sorted := make([]string, 0, len(files))
	for k := range files {
		if strings.Contains(k, args[0]) {
			sorted = append(sorted, k)
		}
	}
	if len(sorted) == 0 {
		return nil, "", 0, errors.New("no file found for the given search")
	}
	sort.Strings(sorted)
	firstMatch := sorted[0]
	skipped = len(sorted) - 1

	relPath := strings.TrimPrefix(firstMatch, moduleDir+"/")
	index := strings.IndexByte(relPath, '/')
	if index == -1 {
		targetFile = relPath
	} else {
		targetFile = relPath[index+1:]
	}
	file, ok := files[sorted[0]]
	if !ok {
		return nil, "", 0, fmt.Errorf("failed to open %s", firstMatch)
	}
	return file, targetFile, skipped, nil
}

func (cmd *Cmd) Inspect(args []string, files map[string]*covFile, moduleDir string) {
	if len(args) < 1 {
		_, _ = fmt.Fprintf(cmd.stderr, "no arguments provided to inspect command\n")
		cmd.exiter.Exit(1)
		return
	}
	var (
		targetFile string
		err        error
		f          fs.File
		file       *covFile
		data       []byte
		skipped    int
	)
	if cmd.config.ExactPath {
		file, targetFile, err = cmd.findExactFile(args, files, moduleDir)
	} else {
		file, targetFile, skipped, err = cmd.findPartialMatchFile(args, files, moduleDir)
	}
	if err != nil {
		_, _ = fmt.Fprint(cmd.stderr, err.Error())
		cmd.exiter.Exit(1)
		return
	}
	f, err = cmd.fsys.Open(targetFile)
	if err != nil {
		_, _ = fmt.Fprintf(cmd.stderr, "failed to open file to inspect: %s", err.Error())
		cmd.exiter.Exit(1)
		return
	}
	defer func() { _ = f.Close() }()

	data, err = io.ReadAll(f)
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

	width := digitsCount(len(lines) - 1)
	if !cmd.config.ExactPath {
		_, _ = fmt.Fprintf(cmd.stdout, "inspect for file: %s\n", targetFile)
	}
	for num, line := range lines {
		_, _ = fmt.Fprintf(cmd.stdout, "%*d| %s\n", width, num+1, line)
	}
	if !cmd.config.ExactPath && skipped > 0 {
		_, _ = fmt.Fprintf(cmd.stdout, "skipped %d other files which matched\n", skipped)
	}
}
