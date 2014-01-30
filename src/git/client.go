package git

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"strings"
)

func echoExec(cmd *exec.Cmd) []byte {
	// make streaming copies of stdout
	var buf bytes.Buffer
	outWriter := io.MultiWriter(&buf, os.Stdout)

	cmd.Stderr = os.Stderr
	cmd.Stdout = outWriter

	if err := cmd.Start(); err != nil {
		panic(err)
	}
	cmd.Wait()

	return buf.Bytes()
}

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
	echoExec(cmd)

	cmd = exec.Command("git", "remote", "add", "origin", url)
	echoExec(cmd)

	cmd = exec.Command("git", "remote", "update")
	echoExec(cmd)

	cmd = exec.Command("git", "rev-list", "--all")
	out := echoExec(cmd)

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
	echoExec(cmd)

	cmd = exec.Command("git", "reset", "--hard", sha)
	echoExec(cmd)

	cmd = exec.Command("git", "show-branch", "--list")
	out = echoExec(cmd)
	commit := strings.Split(string(out), "\n")[0]

	cmd = exec.Command("git", "log", "--pretty=format:'%H'")
	out = echoExec(cmd)
	revlist := strings.Split(string(out), "\n")

	cmd = exec.Command("git", "submodule", "update", "--init")
	echoExec(cmd)

	return Info{
		Commit:  commit,
		Sha:     sha,
		RevList: revlist,
	}
}
