package main

import (
	"atlantis/builder/api"
	"atlantis/builder/docker"
	"flag"
	"github.com/BurntSushi/toml"
	"github.com/jigish/go-flags"
	"log"
)

type ServerOpts struct {
	ConfigFile string `long:"config-file" default:"/etc/atlantis/builder/server.toml" description:"the config file to use"`
	Registry   string `long:"registry" default:"localhost" description:"the registry to use")`
}

type BuilderConfig struct {
	Registry string `toml:"registry_host"`
}

func main() {
	var layerPath = flag.String("layer-path", "/opt/atlantis/builder/layers", "path to overlay layers")
	var manifestDir = flag.String("manifest-dir", "/opt/atlantis/builder/manifests", "dir to store manifests")
	var port = flag.Int("port", 8080, "port to run on")
	flag.Parse()

	builderdOpts := &ServerOpts{}
	config := &BuilderConfig{}

	parser := flags.NewParser(builderdOpts, flags.Default)
	parser.Parse()

	_, err := toml.DecodeFile(builderdOpts.ConfigFile, config)
	if err != nil {
		log.Fatalln(err)
	}
	docker.LogOutput = true
	api.New(uint16(*port), config.Registry, *layerPath, *manifestDir).Run()
}
