package main

import (
	"atlantis/builder/build"
	"atlantis/builder/docker"
	"atlantis/builder/layers"
	"flag"
	"fmt"
	"os"
)

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
			build.Boot(client, *path, layers.ReadLayerInfo(*path))
		} else {
			fmt.Fprintf(os.Stderr, "%s does not exist or not a directory", *path)
		}
	} else {
		docker.LogOutput = true
		if *url == "" || *sha == "" || *rel == "" || *manifestDir == "" {
			panic("provide url, sha, rel path, and manifest dir!")
		}
		build.App(client, *url, *sha, *rel, *manifestDir, layers.ReadLayerInfo(*path))
	}
}
