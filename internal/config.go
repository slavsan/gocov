package internal

import (
	"fmt"
	"strings"
)

func (cmd *Cmd) Config(gocovConfig *GocovConfig) {
	if gocovConfig != nil {
		_, _ = fmt.Fprintf(cmd.stdout, "%s\n", gocovConfig.Contents)
		return
	}
	_, _ = fmt.Fprintf(cmd.stdout, strings.Join([]string{ //nolint:staticcheck // SA1006
		`{`,
		`  "threshold": 50,`,
		`  "ignore": [`,
		`  ]`,
		`}`,
		``,
	}, "\n"))
}
