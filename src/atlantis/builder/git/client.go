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

package git

import (
	"atlantis/builder/util"
	"os"
	"os/exec"
	"strings"
)

type Info struct {
	Commit  string   `json:"commit"`
	Sha     string   `json:"sha"`
	RevList []string `json:"rev_list"`
}

func Checkout(url, sha, path string) Info {
	if err := os.Chdir(path); err != nil {
		panic(err)
	}

	cmd := exec.Command("git", "init")
	util.EchoExec(cmd)

	cmd = exec.Command("git", "remote", "add", "origin", url)
	util.EchoExec(cmd)

	cmd = exec.Command("git", "remote", "update")
	util.EchoExec(cmd)

	cmd = exec.Command("git", "rev-list", "--all")
	out := util.EchoExec(cmd)

	var found bool
	for _, s := range strings.Split(string(out), "\n") {
		if strings.Trim(s, "\n") == sha {
			found = true
			break
		}
	}
	if !found {
		panic("sha " + sha + " not found in repository!")
	}

	cmd = exec.Command("git", "fetch", "origin", sha)
	util.EchoExec(cmd)

	cmd = exec.Command("git", "reset", "--hard", sha)
	util.EchoExec(cmd)

	cmd = exec.Command("git", "show-branch", "--list")
	out = util.EchoExec(cmd)
	commit := strings.Split(string(out), "\n")[0]

	cmd = exec.Command("git", "log", "--pretty=format:%H")
	out = util.EchoExec(cmd)
	revlist := strings.Split(string(out), "\n")

	cmd = exec.Command("git", "submodule", "update", "--init")
	util.EchoExec(cmd)

	return Info{
		Commit:  commit,
		Sha:     sha,
		RevList: revlist,
	}
}
