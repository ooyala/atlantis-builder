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
		Repository:c.URL + "/" + repository,
		Registry:  c.URL,
	}

	err := c.client.PullImage(pullOpts)
	switch err {
	case docker.ErrNoSuchImage:
		return false
	case nil:
		return true
	}

	panic(err)
}

func (c *Client) PushImage(repository string, stream bool) {
	pushOpts := docker.PushImageOptions{
		Name: c.URL + "/" + repository,
		Registry: c.URL,
	}
	if stream {
		pushOpts.OutputStream = os.Stdout
	} else {
		pushOpts.OutputStream = ioutil.Discard
	}
	authConf := docker.AuthConfiguration{}

	if err := c.client.PushImage(pushOpts, authConf); err != nil {
		panic(err)
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
		c.client.KillContainer(container.ID)
		panic("run script timed out in 5 minutes")
	}

	// NOTE(jigish) Should we pass the bind mount and port configuration here during the build?
	opts := docker.CommitContainerOptions{Container: container.ID, Repository: c.URL + "/" + imageTo, Run: &docker.Config{}}
	c.client.CommitContainer(opts)
}
