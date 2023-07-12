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

const exampleCoverageOut6 = `mode: set
example/cmd/exec.go:5.13,7.2 1 0
example/main.go:5.13,7.2 1 0
example/internal/exec.go:8.42,11.13 2 1
example/internal/exec.go:15.2,15.13 1 0
example/internal/exec.go:19.2,19.43 1 0
example/internal/exec.go:11.13,13.3 1 1
example/internal/exec.go:15.13,17.3 1 0
example/internal/exec.go:22.24,24.2 1 1
example/internal/exec.go:26.29,28.2 1 0
`

const expectedHTMLOutput = `<!doctype html>
<head>
<title>My coverage</title>
<style>
    div { border: 1px solid transparent; }
    .table { border: 1px solid #aaa; border-collapse: collapse; }
    .source { display: none; }
    .visible { display: block; }
    td { border: 1px solid #aaa; padding: 5px; }
    a { text-decoration: none; }
    .ok { background: rgb(233,245,212); }
    .ok2 { background: rgb(94,145,53); }
    .warn { background: rgb(253,245,200); }
    .warn2 { background: rgb(242,202,83); }
    .error { background: pink; }
    .progress { width: 100px; height: 20px; }
    .progress > div { height: calc(20px - 1px); }
    .ok .progress { border: 1px solid rgb(94,145,53); }
    .ok .progress > div { background: rgb(94,145,53); }
    .warn .progress { border: 1px solid rgb(242,202,83); }
    .warn .progress > div { background: rgb(242,202,83); }
    .error .progress { border: 1px solid darkred; }
    .error .progress > div { background: darkred; }
    .indicator { height: 10px; margin: 10px 0; }
    .indicator.ok { background: rgb(94,145,53) }
    .indicator.warn { background: rgb(242,202,83) }
    .indicator.error { background: darkred }
</style>
</head>
<body>
<div class="breadcrumbs"></div>
<div class="stats"></div>
<div class="indicator"></div>
<table class="table"></table>
<script class="tree-data" type="application/json">{"name":"example","all":10,"covered":4,"percent":40.00,"path":"example","level":0,"type":"directory","children":[{"name":"cmd","all":1,"covered":0,"percent":0.00,"path":"example/cmd","level":1,"type":"directory","children":[{"name":"exec.go","all":1,"covered":0,"percent":0.00,"path":"example/cmd/exec.go","level":2,"type":"file"}]},{"name":"internal","all":8,"covered":4,"percent":50.00,"path":"example/internal","level":1,"type":"directory","children":[{"name":"exec.go","all":8,"covered":4,"percent":50.00,"path":"example/internal/exec.go","level":2,"type":"file"}]},{"name":"main.go","all":1,"covered":0,"percent":0.00,"path":"example/main.go","level":1,"type":"file"}]}</script>
<div class="source" id="example/cmd/exec.go"><pre>1| package cmd
2| 
3| import "example/internal"
4| 
5| func Exec() <span style="background: pink">{
6| 	internal.Exec(1, 2, 3)
7| }</span>
8| 
</pre></div>
<div class="source" id="example/internal/exec.go"><pre> 1| package internal
 2| 
 3| import (
 4| 	"errors"
 5| 	"fmt"
 6| )
 7| 
 8| func Exec(op int, a, b int) (int, error) {
 9| 	fmt.Printf("here...\n")
10| 
11| 	if op == 1 {
12| 		return sum(a, b), nil
13| 	}
14| 
15| 	<span style="background: pink">if op == 2 </span><span style="background: pink">{
16| 		return subtract(a, b), nil
17| 	}</span>
18| 
19| 	<span style="background: pink">return 0, errors.New("unknown operation")</span>
20| }
21| 
22| func sum(a, b int) int {
23| 	return a + b
24| }
25| 
26| func subtract(a, b int) int <span style="background: pink">{
27| 	return a - b
28| }</span>
29| 
</pre></div>
<div class="source" id="example/main.go"><pre>1| package main
2| 
3| import "example/cmd"
4| 
5| func main() <span style="background: pink">{
6| 	cmd.Exec()
7| }</span>
8| 
</pre></div>

<script>
<!-- SCRIPT -->
</script>
</body>
`

func TestStdoutReport2(t *testing.T) {
	testCases := []struct {
		title                    string
		fsys                     fs.StatFS
		config                   *internal.Config
		expectedStdout           string
		expectedStderr           string
		expectedFileWriterOutput string
		expectedExitCode         int
		args                     []string
	}{
		{
			title: "with example coverage.out file and stdout report and colors disabled",
			fsys: fstest.MapFS{
				"go.mod":       {Data: []byte(`module github.com/slavsan/gospec`)},
				"coverage.out": {Data: []byte(exampleCoverageOut6)},
				"cmd/exec.go": {Data: []byte(strings.Join([]string{
					`package cmd`,
					``,
					`import "example/internal"`,
					``,
					`func Exec() {`,
					`	internal.Exec(1, 2, 3)`,
					`}`,
					``,
				}, "\n"))},
				"internal/exec.go": {Data: []byte(strings.Join([]string{
					`package internal`,
					``,
					`import (`,
					`	"errors"`,
					`	"fmt"`,
					`)`,
					``,
					`func Exec(op int, a, b int) (int, error) {`,
					`	fmt.Printf("here...\n")`,
					``,
					`	if op == 1 {`,
					`		return sum(a, b), nil`,
					`	}`,
					``,
					`	if op == 2 {`,
					`		return subtract(a, b), nil`,
					`	}`,
					``,
					`	return 0, errors.New("unknown operation")`,
					`}`,
					``,
					`func sum(a, b int) int {`,
					`	return a + b`,
					`}`,
					``,
					`func subtract(a, b int) int {`,
					`	return a - b`,
					`}`,
					``,
				}, "\n"))},
				"main.go": {Data: []byte(strings.Join([]string{
					`package main`,
					``,
					`import "example/cmd"`,
					``,
					`func main() {`,
					`	cmd.Exec()`,
					`}`,
					``,
				}, "\n"))},
			},
			config: &internal.Config{
				Color:      false,
				HTMLOutput: true,
			},
			expectedStdout: strings.Join([]string{
				``,
			}, "\n"),
			expectedStderr:           "",
			expectedExitCode:         0,
			expectedFileWriterOutput: strings.ReplaceAll(expectedHTMLOutput, "<!-- SCRIPT -->", internal.Script),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.title, func(t *testing.T) {
			var (
				stdout     bytes.Buffer
				stderr     bytes.Buffer
				fileWriter = &fileWriterMock{f: &bytes.Buffer{}}
			)
			exiter := &exiterMock{}
			internal.NewCommand(&stdout, &stderr, tc.fsys, tc.config, exiter, fileWriter).Exec(internal.Report, tc.args)
			if tc.expectedStdout != stdout.String() {
				t.Errorf("table does not match\n\texpected:\n`%s`\n\tactual:\n`%s`\n", tc.expectedStdout, stdout.String())
			}
			if tc.expectedStderr != stderr.String() {
				t.Errorf("table does not match\n\texpected:\n`%s`\n\tactual:\n`%s`\n", tc.expectedStderr, stderr.String())
			}
			if tc.expectedExitCode != exiter.code {
				t.Errorf("exit code does not match\n\texpected:\n`%d`\n\tactual:\n`%d`\n", tc.expectedExitCode, exiter.code)
			}
			if tc.expectedFileWriterOutput != fileWriter.f.(*bytes.Buffer).String() {
				t.Errorf("table does not match\n\texpected:\n`%s`\n\tactual:\n`%s`\n", tc.expectedFileWriterOutput, fileWriter.f.(*bytes.Buffer).String())
				linesLeft := strings.Split(tc.expectedFileWriterOutput, "\n")
				linesRight := strings.Split(fileWriter.f.(*bytes.Buffer).String(), "\n")
				for i := range linesLeft {
					if linesLeft[i] != linesRight[i] {
						t.Errorf("line %d does not match\n\t left: '%s'\n\tright: '%s'\n",
							i,
							strings.ReplaceAll(linesLeft[i], "\t", "__TAB__"),
							strings.ReplaceAll(linesRight[i], "\t", "__TAB__"),
						)
						break
					}
				}
			}
		})
	}
}
