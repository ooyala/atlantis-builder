package manifest

import (
	"strings"
	// vendored packages
	"github.com/BurntSushi/toml"
)

type Data struct {
	Name          string   `toml:"name"`
	Description   string   `toml:"description"`
	Internal      bool     `toml:"internal"`
	AppType       string   `toml:"app_type"`
	JavaType      string   `toml:"java_type"`
	RunCommands   []string `toml:"run_commands"`
	Dependencies  []string `toml:"dependencies"`
	SetupCommands []string `toml:"setup_commands"`
	CPUShares     uint     `toml:"cpu_shares"`
	MemoryLimit   uint     `toml:"memory_limit"`

	// FIXME(manas) Deprecated, TBD.
	RunCommand interface{} `toml:"run_command"`
}

func Read(fname string) Data {
	var manifest Data
	if _, err := toml.DecodeFile(fname, &manifest); err != nil {
		panic(err)
	}

	fixCompat(&manifest)
	return manifest
}

func fixCompat(manifest *Data) {
	app_type := strings.Split(manifest.AppType, "-")
	if app_type[0] == "java1.7" && len(app_type) > 1 {
		manifest.AppType = app_type[0]
		manifest.JavaType = app_type[1]
	}

	if manifest.RunCommands != nil && len(manifest.RunCommands) > 0 {
		return
	}

	switch runCommands := manifest.RunCommand.(type) {
	case string:
		manifest.RunCommands = []string{runCommands}
	case []interface{}:
		manifest.RunCommands = []string{}
		for _, runCmd := range runCommands {
			cmd, _ := runCmd.(string)
			manifest.RunCommands = append(manifest.RunCommands, cmd)
		}
	}
}
