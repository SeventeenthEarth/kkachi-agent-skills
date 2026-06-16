package version

import (
	"fmt"
	"runtime/debug"
)

const (
	CommandName = "kkachi-agent-skills"
	CLIVersion  = "0.1.4"
)

type Info struct {
	OK            bool   `json:"ok"`
	Command       string `json:"command"`
	CLIVersion    string `json:"cli_version"`
	ModulePath    string `json:"module_path"`
	ModuleVersion string `json:"module_version"`
	GitCommit     string `json:"git_commit"`
	Dirty         bool   `json:"dirty"`
}

func Current() Info {
	info := Info{
		OK:            true,
		Command:       "version",
		CLIVersion:    CLIVersion,
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
	return fmt.Sprintf("%s %s", CommandName, CLIVersion)
}
