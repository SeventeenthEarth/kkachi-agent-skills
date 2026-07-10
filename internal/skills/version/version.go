package version

import (
	"fmt"
	"runtime/debug"
)

const (
	CommandName = "kkachi-agent-skills"
	CLIVersion  = "0.2.5"
)

type Info struct {
	Name          string `json:"name"`
	Version       string `json:"version"`
	ModulePath    string `json:"-"`
	ModuleVersion string `json:"-"`
	GitCommit     string `json:"-"`
	Dirty         bool   `json:"-"`
}

func Current() Info {
	info := Info{
		Name:          CommandName,
		Version:       CLIVersion,
		ModulePath:    "github.com/SeventeenthEarth/kkachi-agent-skills",
		ModuleVersion: "(devel)",
		GitCommit:     "",
		Dirty:         false,
	}
	buildInfo, ok := debug.ReadBuildInfo()
	if !ok {
		return info
	}
	if buildInfo.Main.Path != "" {
		info.ModulePath = buildInfo.Main.Path
	}
	if buildInfo.Main.Version != "" {
		info.ModuleVersion = buildInfo.Main.Version
	}
	for _, setting := range buildInfo.Settings {
		switch setting.Key {
		case "vcs.revision":
			info.GitCommit = setting.Value
		case "vcs.modified":
			info.Dirty = setting.Value == "true"
		}
	}
	return info
}

func Human() string {
	return fmt.Sprintf("%s v%s", CommandName, CLIVersion)
}
