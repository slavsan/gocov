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

func TestConfigFile(t *testing.T) {
	testCases := []struct {
		title            string
		fsys             fs.StatFS
		config           *internal.Config
		expectedStdout   string
		expectedStderr   string
		expectedExitCode int
	}{
		{
			title: "with missing config file should return a default config",
			fsys: fstest.MapFS{
				"go.mod":       {Data: []byte(`module github.com/slavsan/gospec`)},
				"coverage.out": {Data: []byte(exampleCoverageOut)},
			},
			config: &internal.Config{
				Color: false,
			},
			expectedStdout: strings.Join([]string{
				`{`,
				`  "threshold": 50,`,
				`  "ignore": [`,
				`  ]`,
				`}`,
				``,
			}, "\n"),
			expectedStderr:   "",
			expectedExitCode: 0,
		},
		{
			title: "with defined config file should just return it",
			fsys: fstest.MapFS{
				"go.mod":       {Data: []byte(`module github.com/slavsan/gospec`)},
				"coverage.out": {Data: []byte(exampleCoverageOut)},
				".gocov": {Data: []byte(strings.Join([]string{
					`{`,
					`  "threshold": 75.52`,
					`}`,
				}, "\n"))},
			},
			config: &internal.Config{
				Color: false,
			},
			expectedStdout: strings.Join([]string{
				`{`,
				`  "threshold": 75.52`,
				`}`,
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
			internal.NewCommand(&stdout, &stderr, tc.fsys, tc.config, exiter, &fileWriterMock{}).Exec(internal.ConfigFile, []string{})
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
