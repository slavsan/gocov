package internal

import (
	"fmt"
)

func (cmd *Cmd) Check(tree *Tree, gocovConfig *GocovConfig) {
	actualCoveragePercent := float64(tree.Root.covered) * 100 / float64(tree.Root.allStatements)
	if gocovConfig == nil {
		_, _ = fmt.Fprintf(cmd.stderr, "Coverage check failed: missing .gocov file with defined threshold\n")
		cmd.exiter.Exit(1)
		return
	}
	if actualCoveragePercent < gocovConfig.Threshold {
		_, _ = fmt.Fprintf(cmd.stderr, "Coverage check failed: expected to have %.2f coverage, but got %.2f\n", gocovConfig.Threshold, actualCoveragePercent)
		cmd.exiter.Exit(1)
	}
}
