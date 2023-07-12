package internal

import (
	"fmt"
	"strings"
)

func (cmd *Cmd) Config() {
	if cmd.config.File != nil {
		_, _ = fmt.Fprintf(cmd.stdout, "%s\n", cmd.config.File.Contents)
		return
	}
	_, _ = fmt.Fprint(cmd.stdout, strings.Join([]string{
		`{`,
		`  "threshold": 50,`,
		`  "ignore": [`,
		`  ]`,
		`}`,
		``,
	}, "\n"))
}
