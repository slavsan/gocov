package cmd

import (
	"os"

	"github.com/slavsan/gocov/internal"
)

func Exec() {
	config := &internal.Config{}
	config.Color = true
	internal.Exec(os.Stdout, os.DirFS("."), config)
}
