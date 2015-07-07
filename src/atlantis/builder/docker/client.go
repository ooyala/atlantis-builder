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

package docker

import (
	"fmt"
	"github.com/fsouza/go-dockerclient"
	"io/ioutil"
	"os"
	"path"
	"time"
)

var LogOutput bool

type Client struct {
	URL    string
	client *docker.Client
}

func New(url string) *Client {
	dockerClient, err := docker.NewClient("unix:///var/run/docker.sock")
	if err != nil {
		panic(err)
	}
	return &Client{URL: url, client: dockerClient}
}

func (c *Client) PullImage(repository string) bool {
	pullOpts := docker.PullImageOptions{
		Repository:   c.URL + "/" + repository,
		Registry:     c.URL,
		OutputStream: os.Stdout,
	}

	// If PullImage succeeds, image exists and
	// we return true.
	return c.client.PullImage(pullOpts, docker.AuthConfiguration{}) == nil
}

func (c *Client) PushImage(repository string, stream bool) {
	pushOpts := docker.PushImageOptions{
		Name:         c.URL + "/" + repository,
		Registry:     c.URL,
		OutputStream: ioutil.Discard,
	}
	if stream {
		pushOpts.OutputStream = os.Stdout
	}

	authConf := docker.AuthConfiguration{}

	if err := c.client.PushImage(pushOpts, authConf); err != nil {
		fmt.Fprintf(os.Stderr, "PushImage error: %s\n", err.Error())
		time.Sleep(time.Duration(30) * time.Second)
		if err = c.client.PushImage(pushOpts, authConf); err != nil {
			defer c.client.RemoveImage(repository)
			panic(err)
		}
	}
}

func (c *Client) ImageExists(repository string) bool {
	imageName := c.URL + "/" + repository

	_, err := c.client.InspectImage(imageName)
	switch err {
	case docker.ErrNoSuchImage:
		return c.PullImage(repository)
	case nil:
		return true
	}

	panic(err)
}

func (c *Client) OverlayAndCommit(imageFrom, imageTo, bindFrom, bindTo string, tout time.Duration, runScript ...string) {
	containerConfig := &docker.Config{
		Cmd:   runScript,
		Image: c.URL + "/" + imageFrom,
		Volumes: map[string]struct{}{
			bindTo: struct{}{},
		},
	}
	hostConfig := &docker.HostConfig{
		Privileged: true,
		Binds: []string{
			fmt.Sprintf("%s:%s", bindFrom, bindTo),
		},
	}

	uniqName := fmt.Sprintf("%s-%d", path.Base(imageTo), time.Now().Unix())
	container, err := c.client.CreateContainer(docker.CreateContainerOptions{Name: uniqName, Config: containerConfig})
	if err != nil {
		panic(err)
	}

	// Don't clean the container if we paniced, exporting it is useful to get to provisioning logs.
	defer func() {
		if r := recover(); r == nil {
			c.client.RemoveContainer(docker.RemoveContainerOptions{ID: container.ID})
		} else {
			panic(r)
		}
	}()

	if err = c.client.StartContainer(container.ID, hostConfig); err != nil {
		panic(err)
	}

	if LogOutput {
		attachOptions := docker.AttachToContainerOptions{
			Container:    container.ID,
			OutputStream: os.Stdout,
			Stdout:       true,
			Stream:       true,
		}

		if err = c.client.AttachToContainer(attachOptions); err != nil {
			panic(err)
		}
	}

	result := make(chan int)
	go func() {
		for {
			inspect, err := c.client.InspectContainer(container.ID)
			if err != nil {
				panic(err)
			}

			if !inspect.State.Running {
				result <- inspect.State.ExitCode
				return
			}
			time.Sleep(time.Second)
		}
	}()

	select {
	case ec := <-result:
		if ec != 0 {
			panic(fmt.Sprintf("run script failed: %d", ec))
		}
	case <-time.After(tout):
		c.client.KillContainer(docker.KillContainerOptions{ID: container.ID})
		panic(fmt.Sprintf("run script timed out in %s", tout))
	}

	// NOTE(jigish) Should we pass the bind mount and port configuration here during the build?
	opts := docker.CommitContainerOptions{Container: container.ID, Repository: c.URL + "/" + imageTo, Run: &docker.Config{}}
	c.client.CommitContainer(opts)
}
