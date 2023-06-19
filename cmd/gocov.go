package cmd

import (
	"os"

	"github.com/slavsan/gocov/internal"
)

func Exec() {
	config := &internal.Config{}
	config.Color = true

	command := internal.Report

	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "check":
			command = internal.Check
		}
	}

	internal.Exec(
		command,
		os.Stdout,
		os.Stderr,
		os.DirFS("."),
		config,
		&internal.ProcessExiter{},
	)
}
