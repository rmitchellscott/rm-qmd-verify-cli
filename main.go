package main

import (
	"github.com/rmitchellscott/rm-qmd-verify-cli/internal/commands"
)

var (
	version = "dev"
)

func main() {
	commands.Version = version

	commands.Execute()
}
