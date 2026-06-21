package main

import (
	"embed"
	"os"

	"github.com/SeventeenthEarth/kkachi-agent-skills/internal/skills/cli"
	"github.com/SeventeenthEarth/kkachi-agent-skills/internal/skills/discovery"
)

//go:embed skills skill-pack.yaml templates registries scripts
var embeddedSource embed.FS

func main() {
	discovery.ConfigureEmbeddedSource(embeddedSource, "embedded://github.com/SeventeenthEarth/kkachi-agent-skills")
	os.Exit(cli.Main(os.Args[1:], os.Stdout, os.Stderr, nil))
}
