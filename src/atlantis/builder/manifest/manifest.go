/* Copyright 2014 Ooyala, Inc. All rights reserved.
 *
 * This file is licensed under the Apache License, Version 2.0 (the "License"); you may not use this file
 * except in compliance with the License. You may obtain a copy of the License at
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License is
 * distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and limitations under the License.
 */

package manifest

import (
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"
	// vendored packages
	"github.com/BurntSushi/toml"
)

type Data struct {
	Name          string                       `toml:"name"`
	Description   string                       `toml:"description"`
	Internal      bool                         `toml:"internal"`
	AppType       string                       `toml:"app_type"`
	JavaType      string                       `toml:"java_type"`
	RunCommands   []string                     `toml:"run_commands"`
	Dependencies  []string                     `toml:"dependencies"`
	SetupCommands []string                     `toml:"setup_commands"`
	CPUShares     uint                         `toml:"cpu_shares"`
	MemoryLimit   uint                         `toml:"memory_limit"`
	Logging       map[string]map[string]string `toml:"logging"`

	// FIXME(manas) Deprecated, TBD.
	RunCommand interface{} `toml:"run_command"`
}

func Read(r io.Reader) (*Data, error) {
	var manifest Data
	if _, err := toml.DecodeReader(r, &manifest); err != nil {
		return nil, err
	}

	fixCompat(&manifest)
	return &manifest, nil
}

func ReadFile(fname string) (*Data, error) {
	var manifest Data
	if _, err := toml.DecodeFile(fname, &manifest); err != nil {
		return nil, err
	}

	fixCompat(&manifest)
	return &manifest, nil
}

var LoggingKeys = map[string]bool{"name": true, "panic": true, "alert": true, "crit": true, "error": true, "warn": true, "notice": true, "info": true, "debug": true}

func (man *Data) ValidateFacility(fac string) error {
	facProps := man.Logging[fac]
	dirRegex := regexp.MustCompile("^[\\w\\-]+$")
	fileRegex := regexp.MustCompile("^[\\w\\-.]+$")
	for key, val := range facProps {
		lkey := strings.ToLower(key)
		if lkey == "name" {
			if !dirRegex.MatchString(val) {
				return errors.New(fmt.Sprintf("Invalid directory name %s provided for %s!", val, fac))
			}
			if key != lkey {
				facProps["name"] = val
				delete(facProps, key)
			}
		} else if LoggingKeys[lkey] {
			if !fileRegex.MatchString(val) {
				return errors.New(fmt.Sprintf("Invalid file name %s provided for %s.%s!", val, fac, key))
			}
		} else {
			return errors.New(fmt.Sprintf("Invalid key %s provided for facility %s! Please only provide name and syslog priorities as keys!", key, fac))
		}
	}
	if facProps["name"] == "" {
		facProps["name"] = fac
	}
	return nil
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
			cmd := runCmd.(string)
			manifest.RunCommands = append(manifest.RunCommands, cmd)
		}
	}
}
