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

func checkShaExists(sha string) bool {
	cmd := exec.Command("git", "rev-list", "--all")
	out := util.EchoExec(cmd)

	for _, s := range strings.Split(string(out), "\n") {
		if strings.Trim(s, "\n") == sha {
			return true
		}
	}
	return false
}

func fancyCheckout(url string, sha string) {
	cmd := exec.Command("git", "init")
	util.EchoExec(cmd)

	cmd = exec.Command("git", "remote", "add", "origin", url)
	util.EchoExec(cmd)

	cmd = exec.Command("git", "remote", "update")
	util.EchoExec(cmd)

	if !checkShaExists(sha) {
		panic("sha " + sha + " not found in repository!")
	}

	cmd = exec.Command("git", "fetch", "origin", sha)
	util.EchoExec(cmd)

	cmd = exec.Command("git", "reset", "--hard", sha)
	util.EchoExec(cmd)

	cmd = exec.Command("git", "submodule", "update", "--init")
	util.EchoExec(cmd)
}

func localCheckout(path string, sha string) {
	// Rsync with a trailing slash won't create a subdirectory
	cmd := exec.Command("rsync", "-a", path+"/", ".")
	util.EchoExec(cmd)

	if !checkShaExists(sha) {
		panic("sha " + sha + " not found in repository!")
	}

	cmd = exec.Command("git", "reset", "--hard", sha)
	util.EchoExec(cmd)
}

func Checkout(url, sha, path string) Info {
	if err := os.Chdir(path); err != nil {
		panic(err)
	}

	scheme := strings.SplitN(url, ":", 2)[0]

	if scheme == "file" {
		path := strings.TrimPrefix(url, "file://")
		localCheckout(path, sha)
	} else {
		fancyCheckout(url, sha)
	}

	cmd := exec.Command("git", "show-branch", "--list")
	out := util.EchoExec(cmd)
	commit := strings.Split(string(out), "\n")[0]

	cmd = exec.Command("git", "log", "--pretty=format:%H")
	out = util.EchoExec(cmd)
	revlist := strings.Split(string(out), "\n")

	return Info{
		Commit:  commit,
		Sha:     sha,
		RevList: revlist,
	}
}
