package cmd

import (
	"os"

	"github.com/slavsan/gocov/internal"
)

func Exec() {
	var args []string
	config := &internal.Config{}
	config.Color = true

	command := internal.Report

	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "check":
			command = internal.Check
		case "inspect":
			command = internal.Inspect
			if len(os.Args) > 2 {
				args = append(args, os.Args[2])
			}
			//os.Args[1]
		}
	}

	internal.Exec(
		command,
		args,
		os.Stdout,
		os.Stderr,
		os.DirFS("."),
		config,
		&internal.ProcessExiter{},
	)
}