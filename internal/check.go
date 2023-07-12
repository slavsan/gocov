package internal

import (
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
)

func (cmd *Cmd) Check(tree *Tree) {
	err := cmd.check(tree)
	if err != nil {
		_, _ = fmt.Fprint(cmd.stderr, err.Error(), "\n")
		cmd.exiter.Exit(1)
	}
}

func (cmd *Cmd) check(tree *Tree) error {
	actualCoveragePercent := float64(tree.Root.covered) * 100 / float64(tree.Root.allStatements)
	if cmd.config.File == nil {
		return fmt.Errorf("Coverage check failed: missing .gocov file with defined threshold")
	}
	if actualCoveragePercent < cmd.config.Threshold {
		return fmt.Errorf("Coverage check failed: expected to have %.2f coverage, but got %.2f", cmd.config.Threshold, actualCoveragePercent)
	}

	if cmd.config.File.ReadmeThresholdRegex == "" {
		return nil
	}

	if _, err := cmd.fsys.Stat("README.md"); errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("README.md not found")
	}

	f, err := cmd.fsys.Open("README.md")
	if err != nil {
		return fmt.Errorf("failed to open README.md")
	}
	defer func() { _ = f.Close() }()

	b, err := io.ReadAll(f)
	if err != nil {
		return fmt.Errorf("failed to read README.md")
	}

	r, err := regexp.Compile(cmd.config.File.ReadmeThresholdRegex)
	if err != nil {
		return fmt.Errorf("failed to parse README.md regex")
	}

	var (
		matches         []string
		readmeThreshold float64
	)
	for _, line := range strings.Split(string(b), "\n") {
		matches = r.FindStringSubmatch(line)
		if len(matches) == 2 {
			readmeThreshold, err = strconv.ParseFloat(matches[1], 64)
			if err != nil {
				return fmt.Errorf("failed to parse threshold in readme, threshold is not a valid float: %w", err)
			}

			if actualCoveragePercent < readmeThreshold {
				return fmt.Errorf("Coverage check failed in README.md: expected to have %.2f coverage, but got %.2f", readmeThreshold, actualCoveragePercent)
			}
		}
	}

	return nil
}
