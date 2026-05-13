package main

import (
	"os"

	"github.com/gofurry/fiberx/v3/extra-light/cmd"
)

func main() {
	os.Exit(cmd.Execute(os.Args[1:]))
}
