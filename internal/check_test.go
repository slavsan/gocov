package internal_test

import (
	"bytes"
	"io/fs"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/slavsan/gocov/internal"
)

func TestCheckCoverage(t *testing.T) {
	testCases := []struct {
		title            string
		fsys             fs.StatFS
		config           *internal.Config
		expectedStdout   string
		expectedStderr   string
		expectedExitCode int
	}{
		{
			title: "with coverage below threshold",
			fsys: fstest.MapFS{
				"go.mod":       {Data: []byte(`module github.com/slavsan/gospec`)},
				"coverage.out": {Data: []byte(exampleCoverageOut)},
				".gocov": {Data: []byte(strings.Join([]string{
					`{`,
					`	"threshold": 75.52`,
					`}`,
				}, "\n"))},
			},
			config: &internal.Config{
				Color: false,
			},
			expectedStdout:   "",
			expectedStderr:   "Coverage check failed: expected to have 75.52 coverage, but got 73.37\n",
			expectedExitCode: 1,
		},
		{
			title: "with coverage above threshold",
			fsys: fstest.MapFS{
				"go.mod":       {Data: []byte(`module github.com/slavsan/gospec`)},
				"coverage.out": {Data: []byte(exampleCoverageOut)},
				".gocov": {Data: []byte(strings.Join([]string{
					`{`,
					`	"threshold": 23.88`,
					`}`,
				}, "\n"))},
			},
			config: &internal.Config{
				Color: false,
			},
			expectedStdout:   "",
			expectedStderr:   "",
			expectedExitCode: 0,
		},
		{
			title: "with missing .gocov file",
			fsys: fstest.MapFS{
				"go.mod":       {Data: []byte(`module github.com/slavsan/gospec`)},
				"coverage.out": {Data: []byte(exampleCoverageOut)},
			},
			config: &internal.Config{
				Color: false,
			},
			expectedStdout:   "",
			expectedStderr:   "Coverage check failed: missing .gocov file with defined threshold\n",
			expectedExitCode: 1,
		},
		{
			title: "with missing go.mod file",
			fsys: fstest.MapFS{
				"coverage.out": {Data: []byte(exampleCoverageOut)},
			},
			config: &internal.Config{
				Color: false,
			},
			expectedStdout:   "",
			expectedStderr:   "failed to open go.mod file: open go.mod: file does not exist",
			expectedExitCode: 1,
		},
		{
			title: "with invalid go.mod file",
			fsys: fstest.MapFS{
				"go.mod":       {Data: []byte(`foo github.com/slavsan/gospec`)},
				"coverage.out": {Data: []byte(exampleCoverageOut)},
			},
			config: &internal.Config{
				Color: false,
			},
			expectedStdout:   "",
			expectedStderr:   "invalid go.mod file",
			expectedExitCode: 1,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.title, func(t *testing.T) {
			var stdout bytes.Buffer
			var stderr bytes.Buffer
			exiter := &exiterMock{}
			internal.NewCommand(&stdout, &stderr, tc.fsys, tc.config, exiter, &fileWriterMock{}).Exec(internal.Check, []string{})
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
