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
