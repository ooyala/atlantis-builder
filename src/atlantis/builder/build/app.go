package build

import (
	"atlantis/builder/docker"
	"atlantis/builder/git"
	"atlantis/builder/manifest"
	"atlantis/builder/template"
	"atlantis/builder/util"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"path"
	"path/filepath"
	"strings"
	"time"
)

// NOTE(manas) This programs panics in places you'd expect it to call log.Fatal(). The panic allows
// the deferred clean up functions in main() to execute before the program dies.

func copyApp(overlayDir, sourceDir string) string {
	appDir := path.Join(overlayDir, "/src")
	if err := os.MkdirAll(appDir, 0700); err != nil {
		panic(err)
	}

	walk := func(path string, info os.FileInfo, err error) error {
		// don't copy the git store
		if strings.Contains(path, "/.git") {
			return nil
		}

		target := strings.Replace(path, sourceDir, appDir, 1)
		if info.IsDir() {
			return os.MkdirAll(target, 0700)
		} else {
			src, err := os.Open(path)
			if err != nil {
				return err
			}
			defer src.Close()

			dst, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE, info.Mode())
			if err != nil {
				return err
			}
			defer dst.Close()

			if _, err := io.Copy(dst, src); err != nil {
				return err
			}
		}
		return nil
	}
	if err := filepath.Walk(sourceDir, walk); err != nil {
		panic(err)
	}

	return appDir
}

func writeConfigs(overlayDir string, manifest *manifest.Data) {
	for idx, cmd := range manifest.RunCommands {
		// create /etc/sv/app0
		relPath := fmt.Sprintf("/etc/sv/app%d", idx)
		absPath := path.Join(overlayDir, relPath)
		if err := os.MkdirAll(absPath, 0700); err != nil {
			panic(err)
		}

		// write /etc/sv/app0/run
		absPath = path.Join(absPath, "run")
		template.WriteRunitScript(absPath, cmd, idx)
	}

	// create /etc/rsyslog.d
	if err := os.MkdirAll(path.Join(overlayDir, "/etc/rsyslog.d"), 0700); err != nil {
		panic(err)
	}

	for idx, _ := range manifest.RunCommands {
		// write /etc/rsyslog.d/00.conf
		relPath := fmt.Sprintf("/etc/rsyslog.d/%02d.conf", idx)
		absPath := path.Join(overlayDir, relPath)
		template.WriteRsyslogConfig(absPath, idx)
	}

	// create /etc/atlantis/scripts
	if err := os.MkdirAll(path.Join(overlayDir, "/etc/atlantis/scripts"), 0700); err != nil {
		panic(err)
	}

	absPath := path.Join(overlayDir, "/etc/atlantis/scripts/setup")
	template.WriteSetupScript(absPath, manifest)
}

func writeInfo(overlayDir string, gitInfo git.Info) {
	infoDir := path.Join(overlayDir, "/etc/atlantis/info")
	if err := os.MkdirAll(infoDir, 0755); err != nil {
		panic(err)
	}

	data, err := json.MarshalIndent(gitInfo, "", "  ")
	if err != nil {
		panic(err)
	}

	if err := ioutil.WriteFile(path.Join(infoDir, "build.json"), data, 0644); err != nil {
		panic(err)
	}

	timestr := time.Now().UTC().Format(time.RFC822)
	if err := ioutil.WriteFile(path.Join(infoDir, "build_utc"), []byte(timestr), 0644); err != nil {
		panic(err)
	}
}

func runJavaPrebuild(appDir, javaType string) {
	var cmd *exec.Cmd

	switch javaType {
	case "scala":
		cmd = exec.Command("sbt", "assembly")
	case "maven":
		cmd = exec.Command("mvn", "build")
	}
	cmd.Dir = appDir
	util.EchoExec(cmd)

	walk := func(path string, info os.FileInfo, err error) error {
		if info.IsDir() || strings.HasSuffix(path, ".jar") {
			return nil
		} else {
			return os.RemoveAll(path)
		}
	}
	if err := filepath.Walk(path.Join(appDir, "target"), walk); err != nil {
		panic(err)
	}
}

func copyManifest(manifestDir, fname string) {
	// copy manifest
	copyFile, err := os.Create(path.Join(manifestDir, "manifest.toml"))
	if err != nil {
		panic(err)
	}
	defer copyFile.Close()

	manFile, err := os.Open(fname)
	if err != nil {
		panic(err)
	}
	defer manFile.Close()

	io.Copy(copyFile, manFile)
}

func App(client *docker.Client, buildURL, buildSha, relPath, manifestDir string, layers *Layers) {
	usr, err := user.Current()
	if err != nil {
		panic(err)
	}

	cloneDir, err := ioutil.TempDir(usr.HomeDir, path.Base(buildURL))
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(cloneDir)

	gitInfo := git.Checkout(buildURL, buildSha, cloneDir)

	sourceDir := path.Join(cloneDir, relPath)

	manifestFname := path.Join(sourceDir, "manifest.toml")
	if _, err := os.Stat(manifestFname); os.IsNotExist(err) {
		panic(err)
	}

	// read manifest
	manifest, err := manifest.ReadFile(manifestFname)
	if err != nil {
		panic(err)
	}
	copyManifest(manifestDir, manifestFname)

	builderLayer, err := layers.BuilderLayerName(manifest.AppType)
	if err != nil {
		panic(err)
	}

	overlayDir, err := ioutil.TempDir(usr.HomeDir, manifest.Name)
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(overlayDir)

	appDir := copyApp(overlayDir, sourceDir)

	appDockerName := fmt.Sprintf("apps/%s-%s", manifest.Name, gitInfo.Sha)

	if client.ImageExists(appDockerName) {
		if os.Getenv("REBUILD_IMAGE") == "" {
			fmt.Println("Image exists!")
			return
		}
	}

	writeInfo(overlayDir, gitInfo)
	writeConfigs(overlayDir, manifest)

	if manifest.AppType == "java1.7" {
		runJavaPrebuild(appDir, manifest.JavaType)
	}
	client.OverlayAndCommit(builderLayer, appDockerName, overlayDir, "/overlay", 5*time.Minute, "/etc/atlantis/scripts/build", "/overlay")
	client.PushImage(appDockerName, true)
}