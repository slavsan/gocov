//nolint:funlen
package internal_test

import (
	"bytes"
	"io/fs"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/slavsan/gocov/internal"
)

func TestStdoutReport(t *testing.T) { //nolint:maintidx
	testCases := []struct {
		title            string
		fsys             fs.StatFS
		config           *internal.Config
		expectedStdout   string
		expectedStderr   string
		expectedExitCode int
		args             []string
	}{
		{
			title: "with example coverage.out file and stdout report and colors disabled",
			fsys: fstest.MapFS{
				"go.mod":       {Data: []byte(`module github.com/slavsan/gospec`)},
				"coverage.out": {Data: []byte(exampleCoverageOut)},
			},
			config: &internal.Config{
				Color: false,
			},
			expectedStdout: strings.Join([]string{
				`|------------------|---------|----------|------------|`,
				`| File             |   Stmts |  % Stmts | Progress   |`,
				`|------------------|---------|----------|------------|`,
				`| gospec           | 237/323 |   73.37% | ■■■■■■■    |`,
				`|   cmd            |    0/28 |    0.00% |            |`,
				`|     cover.go     |    0/28 |    0.00% |            |`,
				`|   expect.go      |   43/86 |   50.00% | ■■■■■      |`,
				`|   featurespec.go |  92/107 |   85.98% | ■■■■■■■■   |`,
				`|   gospec.go      | 102/102 |  100.00% | ■■■■■■■■■■ |`,
				`|------------------|---------|----------|------------|`,
				``,
			}, "\n"),
			expectedStderr:   "",
			expectedExitCode: 0,
		},
		{
			title: "with example coverage.out file and stdout report and colors enabled",
			fsys: fstest.MapFS{
				"go.mod":       {Data: []byte(`module github.com/slavsan/gospec`)},
				"coverage.out": {Data: []byte(exampleCoverageOut)},
			},
			config: &internal.Config{
				Color: true,
			},
			expectedStdout: strings.Join([]string{
				"|------------------|---------|----------|------------|",
				"| File             |   Stmts |  % Stmts | Progress   |",
				"|------------------|---------|----------|------------|",
				"|\033[0;33m gospec           \033[0m| \033[0;33m237/323\033[0m | \033[0;33m  73.37%\033[0m | \033[0;33m■■■■■■■   \033[0m |",
				"|\033[0;31m   cmd            \033[0m| \033[0;31m   0/28\033[0m | \033[0;31m   0.00%\033[0m | \033[0;31m          \033[0m |",
				"|\033[0;31m     cover.go     \033[0m| \033[0;31m   0/28\033[0m | \033[0;31m   0.00%\033[0m | \033[0;31m          \033[0m |",
				"|\033[0;33m   expect.go      \033[0m| \033[0;33m  43/86\033[0m | \033[0;33m  50.00%\033[0m | \033[0;33m■■■■■     \033[0m |",
				"|\033[0;32m   featurespec.go \033[0m| \033[0;32m 92/107\033[0m | \033[0;32m  85.98%\033[0m | \033[0;32m■■■■■■■■  \033[0m |",
				"|\033[0;32m   gospec.go      \033[0m| \033[0;32m102/102\033[0m | \033[0;32m 100.00%\033[0m | \033[0;32m■■■■■■■■■■\033[0m |",
				"|------------------|---------|----------|------------|",
				"",
			}, "\n"),
			expectedStderr:   "",
			expectedExitCode: 0,
		},
		{
			title: "with another example coverage.out file and stdout report",
			fsys: fstest.MapFS{
				"go.mod":       {Data: []byte(`module github.com/slavsan/gocov`)},
				"coverage.out": {Data: []byte(exampleCoverageOut2)},
			},
			config: &internal.Config{
				Color: false,
			},
			expectedStdout: strings.Join([]string{
				`|--------------|---------|----------|------------|`,
				`| File         |   Stmts |  % Stmts | Progress   |`,
				`|--------------|---------|----------|------------|`,
				`| gocov        | 133/142 |   93.66% | ■■■■■■■■■  |`,
				`|   cmd        |     0/3 |    0.00% |            |`,
				`|     gocov.go |     0/3 |    0.00% |            |`,
				`|   internal   | 133/138 |   96.38% | ■■■■■■■■■  |`,
				`|     gocov.go | 133/138 |   96.38% | ■■■■■■■■■  |`,
				`|   main.go    |     0/1 |    0.00% |            |`,
				`|--------------|---------|----------|------------|`,
				``,
			}, "\n"),
			expectedStderr:   "",
			expectedExitCode: 0,
		},
		{
			title: "with smaller example coverage.out file and stdout report",
			fsys: fstest.MapFS{
				"go.mod":       {Data: []byte(`module github.com/slavsan/gocov`)},
				"coverage.out": {Data: []byte(exampleCoverageOut3)},
			},
			config: &internal.Config{
				Color: false,
			},
			expectedStdout: strings.Join([]string{
				`|--------------|--------|----------|------------|`,
				`| File         |  Stmts |  % Stmts | Progress   |`,
				`|--------------|--------|----------|------------|`,
				`| gocov        |   4/15 |   26.67% | ■■         |`,
				`|   cmd        |   0/11 |    0.00% |            |`,
				`|     gocov.go |   0/11 |    0.00% |            |`,
				`|   internal   |    4/4 |  100.00% | ■■■■■■■■■■ |`,
				`|     gocov.go |    4/4 |  100.00% | ■■■■■■■■■■ |`,
				`|--------------|--------|----------|------------|`,
				``,
			}, "\n"),
			expectedStderr:   "",
			expectedExitCode: 0,
		},
		{
			title: "with .gocov file specifying one file to ignore",
			fsys: fstest.MapFS{
				"go.mod":       {Data: []byte(`module github.com/slavsan/gocov`)},
				"coverage.out": {Data: []byte(exampleCoverageOut2)},
				".gocov": {Data: []byte(strings.Join([]string{
					`{`,
					`	"ignore": [`,
					`		"gocov/main.go"`,
					`	]`,
					`}`,
				}, "\n"))},
			},
			config: &internal.Config{
				Color: false,
			},
			expectedStdout: strings.Join([]string{
				`|--------------|---------|----------|------------|`,
				`| File         |   Stmts |  % Stmts | Progress   |`,
				`|--------------|---------|----------|------------|`,
				`| gocov        | 133/141 |   94.33% | ■■■■■■■■■  |`,
				`|   cmd        |     0/3 |    0.00% |            |`,
				`|     gocov.go |     0/3 |    0.00% |            |`,
				`|   internal   | 133/138 |   96.38% | ■■■■■■■■■  |`,
				`|     gocov.go | 133/138 |   96.38% | ■■■■■■■■■  |`,
				`|--------------|---------|----------|------------|`,
				``,
			}, "\n"),
			expectedStderr:   "",
			expectedExitCode: 0,
		},
		{
			title: "with the depth flag provided",
			fsys: fstest.MapFS{
				"go.mod":       {Data: []byte(`module github.com/slavsan/gocov`)},
				"coverage.out": {Data: []byte(exampleCoverageOut2)},
			},
			config: &internal.Config{
				Color: false,
				Depth: 1,
			},
			expectedStdout: strings.Join([]string{
				`|--------------|---------|----------|------------|`,
				`| File         |   Stmts |  % Stmts | Progress   |`,
				`|--------------|---------|----------|------------|`,
				`| gocov        | 133/142 |   93.66% | ■■■■■■■■■  |`,
				`|   cmd        |     0/3 |    0.00% |            |`,
				`|   internal   | 133/138 |   96.38% | ■■■■■■■■■  |`,
				`|   main.go    |     0/1 |    0.00% |            |`,
				`|--------------|---------|----------|------------|`,
				``,
			}, "\n"),
			expectedStderr:   "",
			expectedExitCode: 0,
		},
		{
			title: "with invalid coverage.out file, invalid column value",
			fsys: fstest.MapFS{
				"go.mod": {Data: []byte(`module github.com/slavsan/gocov`)},
				"coverage.out": {Data: []byte(strings.Join([]string{
					`mode: atomic`,
					`github.com/slavsan/gocov/cmd/gocov.go:9.13,16.22 x 0`,
				}, "\n"))},
				".gocov": {Data: []byte(strings.Join([]string{
					`{`,
					`	"ignore": [`,
					`		"gocov/main.go"`,
					`	]`,
					`}`,
				}, "\n"))},
			},
			config: &internal.Config{
				Color: false,
			},
			expectedStdout:   "",
			expectedStderr:   "failed to parse coverage file on line 2",
			expectedExitCode: 1,
		},
		{
			title: "with invalid coverage.out file, invalid first line",
			fsys: fstest.MapFS{
				"go.mod": {Data: []byte(`module github.com/slavsan/gocov`)},
				"coverage.out": {Data: []byte(strings.Join([]string{
					`foo: atomic`,
					`github.com/slavsan/gocov/cmd/gocov.go:9.13,16.22 10 0`,
				}, "\n"))},
				".gocov": {Data: []byte(strings.Join([]string{
					`{`,
					`	"ignore": [`,
					`		"gocov/main.go"`,
					`	]`,
					`}`,
				}, "\n"))},
			},
			config: &internal.Config{
				Color: false,
			},
			expectedStdout:   "",
			expectedStderr:   "invalid coverage file",
			expectedExitCode: 1,
		},
		{
			title: "with invalid .gocov file",
			fsys: fstest.MapFS{
				"go.mod":       {Data: []byte(`module github.com/slavsan/gocov`)},
				"coverage.out": {Data: []byte(exampleCoverageOut)},
				".gocov": {Data: []byte(strings.Join([]string{
					`{`,
					`	"ignore": [`,
					`		"goc`,
					// }
				}, "\n"))},
			},
			config: &internal.Config{
				Color: false,
			},
			expectedStdout:   "",
			expectedStderr:   "failed to parse .gocov config file: unexpected end of JSON input",
			expectedExitCode: 1,
		},
		{
			title: "with full path column",
			fsys: fstest.MapFS{
				"go.mod":       {Data: []byte(`module github.com/slavsan/gospec`)},
				"coverage.out": {Data: []byte(exampleCoverageOut)},
			},
			config: &internal.Config{
				Color:        false,
				WithFullPath: true,
			},
			expectedStdout: strings.Join([]string{
				`|------------------|---------|----------|------------|-----------------------|`,
				`| File             |   Stmts |  % Stmts | Progress   | Full path             |`,
				`|------------------|---------|----------|------------|-----------------------|`,
				`| gospec           | 237/323 |   73.37% | ■■■■■■■    | gospec                |`,
				`|   cmd            |    0/28 |    0.00% |            | gospec/cmd            |`,
				`|     cover.go     |    0/28 |    0.00% |            | gospec/cmd/cover.go   |`,
				`|   expect.go      |   43/86 |   50.00% | ■■■■■      | gospec/expect.go      |`,
				`|   featurespec.go |  92/107 |   85.98% | ■■■■■■■■   | gospec/featurespec.go |`,
				`|   gospec.go      | 102/102 |  100.00% | ■■■■■■■■■■ | gospec/gospec.go      |`,
				`|------------------|---------|----------|------------|-----------------------|`,
				``,
			}, "\n"),
			expectedStderr:   "",
			expectedExitCode: 0,
		},
		{
			title: "with full path column and colors enabled",
			fsys: fstest.MapFS{
				"go.mod":       {Data: []byte(`module github.com/slavsan/gospec`)},
				"coverage.out": {Data: []byte(exampleCoverageOut)},
			},
			config: &internal.Config{
				Color:        true,
				WithFullPath: true,
			},
			expectedStdout: strings.Join([]string{
				"|------------------|---------|----------|------------|-----------------------|",
				"| File             |   Stmts |  % Stmts | Progress   | Full path             |",
				"|------------------|---------|----------|------------|-----------------------|",
				"|\033[0;33m gospec           \033[0m| \033[0;33m237/323\033[0m | \033[0;33m  73.37%\033[0m | \033[0;33m■■■■■■■   \033[0m | \033[0;33mgospec               \033[0m |",
				"|\033[0;31m   cmd            \033[0m| \033[0;31m   0/28\033[0m | \033[0;31m   0.00%\033[0m | \033[0;31m          \033[0m | \033[0;31mgospec/cmd           \033[0m |",
				"|\033[0;31m     cover.go     \033[0m| \033[0;31m   0/28\033[0m | \033[0;31m   0.00%\033[0m | \033[0;31m          \033[0m | \033[0;31mgospec/cmd/cover.go  \033[0m |",
				"|\033[0;33m   expect.go      \033[0m| \033[0;33m  43/86\033[0m | \033[0;33m  50.00%\033[0m | \033[0;33m■■■■■     \033[0m | \033[0;33mgospec/expect.go     \033[0m |",
				"|\033[0;32m   featurespec.go \033[0m| \033[0;32m 92/107\033[0m | \033[0;32m  85.98%\033[0m | \033[0;32m■■■■■■■■  \033[0m | \033[0;32mgospec/featurespec.go\033[0m |",
				"|\033[0;32m   gospec.go      \033[0m| \033[0;32m102/102\033[0m | \033[0;32m 100.00%\033[0m | \033[0;32m■■■■■■■■■■\033[0m | \033[0;32mgospec/gospec.go     \033[0m |",
				"|------------------|---------|----------|------------|-----------------------|",
				"",
			}, "\n"),
			expectedStderr:   "",
			expectedExitCode: 0,
		},
		{
			title: "with selected directory",
			fsys: fstest.MapFS{
				"go.mod":       {Data: []byte(`module github.com/slavsan/gocov`)},
				"coverage.out": {Data: []byte(exampleCoverageOut2)},
			},
			args: []string{"gocov/cmd"},
			config: &internal.Config{
				Color: false,
			},
			expectedStdout: strings.Join([]string{
				`|--------------|---------|----------|------------|`,
				`| File         |   Stmts |  % Stmts | Progress   |`,
				`|--------------|---------|----------|------------|`,
				`| gocov        | 133/142 |   93.66% | ■■■■■■■■■  |`,
				`|   cmd        |     0/3 |    0.00% |            |`,
				`|     gocov.go |     0/3 |    0.00% |            |`,
				`|--------------|---------|----------|------------|`,
				``,
			}, "\n"),
			expectedStderr:   "",
			expectedExitCode: 0,
		},
		{
			title: "with coverage report in set mode",
			fsys: fstest.MapFS{
				"go.mod":       {Data: []byte(`module github.com/slavsan/gocov`)},
				"coverage.out": {Data: []byte(exampleCoverageOut4)},
			},
			config: &internal.Config{
				Color: false,
			},
			expectedStdout: strings.Join([]string{
				`|--------------|--------|----------|------------|`,
				`| File         |  Stmts |  % Stmts | Progress   |`,
				`|--------------|--------|----------|------------|`,
				`| gocov        |   6/15 |   40.00% | ■■■■       |`,
				`|   cmd        |   2/11 |   18.18% | ■          |`,
				`|     gocov.go |   2/11 |   18.18% | ■          |`,
				`|   internal   |    4/4 |  100.00% | ■■■■■■■■■■ |`,
				`|     gocov.go |    4/4 |  100.00% | ■■■■■■■■■■ |`,
				`|--------------|--------|----------|------------|`,
				``,
			}, "\n"),
			expectedStderr:   "",
			expectedExitCode: 0,
		},
		{
			title: "with coverage report in set mode inverted",
			fsys: fstest.MapFS{
				"go.mod":       {Data: []byte(`module github.com/slavsan/gocov`)},
				"coverage.out": {Data: []byte(exampleCoverageOut5)},
			},
			config: &internal.Config{
				Color: false,
			},
			expectedStdout: strings.Join([]string{
				`|--------------|--------|----------|------------|`,
				`| File         |  Stmts |  % Stmts | Progress   |`,
				`|--------------|--------|----------|------------|`,
				`| gocov        |   6/15 |   40.00% | ■■■■       |`,
				`|   cmd        |   2/11 |   18.18% | ■          |`,
				`|     gocov.go |   2/11 |   18.18% | ■          |`,
				`|   internal   |    4/4 |  100.00% | ■■■■■■■■■■ |`,
				`|     gocov.go |    4/4 |  100.00% | ■■■■■■■■■■ |`,
				`|--------------|--------|----------|------------|`,
				``,
			}, "\n"),
			expectedStderr:   "",
			expectedExitCode: 0,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.title, func(t *testing.T) {
			var stdout bytes.Buffer
			var stderr bytes.Buffer
			exiter := &exiterMock{}
			internal.NewCommand(&stdout, &stderr, tc.fsys, tc.config, exiter, &fileWriterMock{}).Exec(internal.Report, tc.args)
			if tc.expectedStdout != stdout.String() {
				t.Errorf("table does not match\n\texpected:\n`%s`\n\tactual:\n`%s`\n", tc.expectedStdout, stdout.String())
			}
			if tc.expectedStderr != stderr.String() {
				t.Errorf("table does not match\n\texpected:\n`%s`\n\tactual:\n`%s`\n", tc.expectedStderr, stderr.String())
			}
			if tc.expectedExitCode != exiter.code {
				t.Errorf("exit code does not match\n\texpected:\n`%d`\n\tactual:\n`%d`\n", tc.expectedExitCode, exiter.code)
			}
		})
	}
}
