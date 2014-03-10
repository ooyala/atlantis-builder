package manifest

import (
	"os"
	"path"
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
	CpuShares     uint     `toml:"cpu_shares"`
	MemoryLimit   uint     `toml:"memory_limit"`
}

func New(sourceDir string) Data {
	fname := path.Join(sourceDir, "manifest.toml")
	if _, err := os.Stat(fname); os.IsNotExist(err) {
		panic(err)
	}

	var manifest Data
	if _, err := toml.DecodeFile(fname, &manifest); err != nil {
		panic(err)
	}

	return manifest
}
