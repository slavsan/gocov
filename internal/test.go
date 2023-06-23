package internal

import (
	"fmt"
	"os/exec"
	"strings"
)

func (cmd *Cmd) Test() {
	coverArgs := []string{"test", "-coverprofile", "coverage.out", "-coverpkg", "./...", "./..."}
	_, _ = fmt.Fprintf(cmd.stdout, "executing: go %s\n", strings.Join(coverArgs, " "))
	execCmd := exec.Command("go", coverArgs...)
	execCmd.Stdout = cmd.stdout
	execCmd.Stderr = cmd.stderr
	err := execCmd.Run()
	if err != nil {
		_, _ = fmt.Fprintf(cmd.stderr, "failed to run `go test` command: %s", err.Error())
		cmd.exiter.Exit(1)
	}
}
