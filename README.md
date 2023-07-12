# gocov

A coverage reporting tool for the Go programming language.

## Install

### Build the binary with
```
go install github.com/slavsan/gocov@latest
```

### Include in your project's vendor directory

If you don't want to install the binary, you could easily add `gocov` to your project by just creating a main file like the following:
```
package main

import "github.com/slavsan/gocov/cmd"

func main() {
	cmd.Exec()
}
```
This way you can vendor it easily and use it during your CI workflow.

## Usage

The `gocov` tool supports several commands:
* `report` - output a pretty table
* `check` - check against the target threshold
* `inspect` - show covered vs uncovered lines in stdout
* `test` - generate a coverage profile with the go test command
* `config` - output current config or the default one

You can provide a `.gocov` file to your project looking like this:
```
{
    "threshold": 80,
    "ignore": [
        "gocov/main.go"
    ]
}
```
so that you could just run
```
gocov test && gocov report && gocov check
```
in order to
* test and generate a coverage profile
* output the coverage report in a table
* make sure the coverage percentage hasn't dropped below the defined threshold

### report

The `report` command outputs a pretty table showing the coverage percentage for each file and directory.

```
$ gocov report
|----------------|---------|----------|------------|
| File           |   Stmts |  % Stmts | Progress   |
|----------------|---------|----------|------------|
| gocov          | 312/403 |   77.42% | ■■■■■■■    |
|   cmd          |    0/64 |    0.00% |            |
|     gocov.go   |    0/64 |    0.00% |            |
|   internal     | 312/339 |   92.04% | ■■■■■■■■■  |
|     check.go   |     8/8 |  100.00% | ■■■■■■■■■■ |
|     config.go  |     4/4 |  100.00% | ■■■■■■■■■■ |
|     gocov.go   | 128/134 |   95.52% | ■■■■■■■■■  |
|     inspect.go |   72/84 |   85.71% | ■■■■■■■■   |
|     report.go  |     1/1 |  100.00% | ■■■■■■■■■■ |
|     test.go    |     0/9 |    0.00% |            |
|     tree.go    |   99/99 |  100.00% | ■■■■■■■■■■ |
|----------------|---------|----------|------------|
```

The command supports several options
```
$ gocov report --help
Usage of report:
  -f, --file string
      coverage profile file (default is coverage.out)
  -d, --depth int
      report on files and directories of certain depth
  --html
      output the coverage in html format
  --no-color
      disable color output
  --with-full-path
      include the full path column in the output
```

### check

The `check` command make sure you haven't dropped below the desired coverage percentage.

The threshold can be set in the `.gocov` config file.

```
$ gocov check
Coverage check failed: expected to have 80.00 coverage, but got 77.42
```

### inspect

The `inspect` command outputs a target file by colouring the non-covered statements.

It's a quick way to show which lines have been covered or not.

### test

The `test` command is just a utility function which runs the `go test` command with the appropriate flags.

Namely:
```
$ gocov test
executing: go test -coverprofile coverage.out -coverpkg ./... ./...
?       github.com/slavsan/gocov        [no test files]
?       github.com/slavsan/gocov/cmd    [no test files]
ok      github.com/slavsan/gocov/internal       0.349s  coverage: 77.2% of statements in ./...
```

You might want to run `go test` with a different set of flags. You can then still use `gocov report` or `gocov check`

### config

The `config` command outputs the current `.gocov` file's contents.

If there is no `.gocov` file created, you can generate a new one with the same command

```
gocov config > .gocov
```

because when the `.gocov` file is missing, `gocov config` outputs the default configuration template.

### .gocov file

The `.gocov` config file is meant to be included in your version control.