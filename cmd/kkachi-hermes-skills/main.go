package main

import (
	"os"

	"github.com/SeventeenthEarth/kkachi-agent-skills/internal/skills/cli"
)

func main() {
	os.Exit(cli.Main(os.Args[1:], os.Stdout, os.Stderr, nil))
}
