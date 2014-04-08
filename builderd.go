package main

import (
	"atlantis/builder/api"
	"atlantis/builder/docker"
	"flag"
)

func main() {
	var layerPath = flag.String("layer-path", "/opt/atlantis/builder/layers", "path to overlay layers")
	var manifestDir = flag.String("manifest-dir", "/opt/atlantis/builder/manifests", "dir to store manifests")
	var registry = flag.String("registry", "localhost", "the registry to use")
	var port = flag.Int("port", 8080, "port to run on")
	flag.Parse()

	docker.LogOutput = true
	api.New(uint16(*port), *registry, *layerPath, *manifestDir).Run()
}
