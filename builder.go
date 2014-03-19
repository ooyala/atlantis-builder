package main

import (
	"build"
	"docker"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

// FIXME(manas) Putting this in util creates cyclic imports.
func readLayerInfo(overlayDir string) *build.Layers {
	baseFile := path.Join(overlayDir, "basename.txt")
	baseName, err := ioutil.ReadFile(baseFile)
	if err != nil {
		panic(err)
	}

	versionFile := path.Join(overlayDir, "version.txt")
	versionNumber, err := ioutil.ReadFile(versionFile)
	if err != nil {
		panic(err)
	}

	layersDir := path.Join(overlayDir, "builder")
	dirs, err := ioutil.ReadDir(layersDir)
	if err != nil {
		panic(err)
	}

	layerNames := []string{}
	for _, dir := range dirs {
		layerNames = append(layerNames, dir.Name())
	}

	return &build.Layers{
		Version:       strings.TrimRight(string(versionNumber), "\n"),
		BaseLayer:     strings.TrimRight(string(baseName), "\n"),
		BuilderLayers: layerNames,
	}
}

func main() {
	// Builder bootstrap.
	var boot = flag.Bool("boot", false, "bootstrap builder layers")
	var path = flag.String("path", "/opt/atlantis/builder/layers", "path to overlay layers")

	// App container builds.
	var url = flag.String("url", "", "url of git repo")
	var sha = flag.String("sha", "", "git sha to build")
	var rel = flag.String("rel", "", "relative path in repository to build")
	var manifestDir = flag.String("manifest-dir", "", "the directory to copy the manifest to")
	flag.Parse()

	registry := os.Getenv("REGISTRY")
	if registry == "" {
		panic("REGISTRY is not in the environment!")
	}
	client := docker.New(registry)

	if *boot {
		fi, err := os.Stat(*path)
		if err == nil && fi.IsDir() {
			build.Boot(client, *path, readLayerInfo(*path))
		} else {
			fmt.Fprintf(os.Stderr, "%s does not exist or not a directory", *path)
		}
	} else {
		docker.LogOutput = true
		if *url == "" || *sha == "" || *rel == "" || *manifestDir == "" {
			panic("provide url, sha, rel path, and manifest dir!")
		}
		build.App(client, *url, *sha, *rel, *manifestDir, readLayerInfo(*path))
	}
}
