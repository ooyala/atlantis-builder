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
