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
	file, ok := files[firstMatch]
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

	lines, err := cmd.getColorizedLines(data, file)
	if err != nil {
		_, _ = fmt.Fprint(cmd.stderr, err.Error())
		cmd.exiter.Exit(1)
		return
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

func (cmd *Cmd) getColorizedLines(data []byte, file *covFile) ([]string, error) {
	lines := strings.Split(string(data), "\n")

	sort.Slice(file.Reports, func(i, j int) bool {
		if file.Reports[i].EndLine == file.Reports[j].EndLine {
			return file.Reports[i].StartColumn > file.Reports[j].StartColumn
		}
		return file.Reports[i].EndLine > file.Reports[j].EndLine
	})

	for _, report := range file.Reports {
		if report.Hits > 0 {
			continue
		}

		lineNum := report.EndLine - 1

		if len(lines) < lineNum+1 {
			return nil, errors.New("running inspect failed, please regenerate the coverage report again")
		}

		if len(lines[lineNum]) < report.EndColumn-1 {
			return nil, errors.New("running inspect failed, please regenerate the coverage report again")
		}

		lines[lineNum] = lines[lineNum][:report.EndColumn-1] + NoColor + lines[lineNum][report.EndColumn-1:]

		lineNum = report.StartLine - 1

		if len(lines) < lineNum+1 {
			return nil, errors.New("running inspect failed, please regenerate the coverage report again")
		}

		if len(lines[lineNum]) < report.StartColumn-1 {
			return nil, errors.New("running inspect failed, please regenerate the coverage report again")
		}

		lines[lineNum] = lines[lineNum][:report.StartColumn-1] + Red + lines[lineNum][report.StartColumn-1:]
	}

	return lines, nil
}
