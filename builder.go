package main

import (
	"build"
	"docker"
	"flag"
	"os"
)

var layers = &build.Layers{
	Version:       "0.2.0",
	BaseLayer:     "precise64",
	BuilderLayers: []string{"go1.1.2", "go1.2", "java1.7", "python2.7.3", "ruby1.9.3"},
}

func main() {
	var boot = flag.Bool("boot", false, "bootstrap")

	var url = flag.String("url", "", "url of git repo")
	var sha = flag.String("sha", "", "git sha to build")
	var rel = flag.String("rel", "", "relative path in repository to build")
	flag.Parse()

	registry := os.Getenv("REGISTRY")
	if registry == "" {
		panic("REGISTRY is not in the environment!")
	}
	client := docker.New(registry)

	if *boot {
		build.Boot(client, layers)
	} else {
		if *url == "" || *sha == "" || *rel == "" {
			panic("provide url, sha and rel path!")
		}
		build.App(client, *url, *sha, *rel, layers)
	}

}
