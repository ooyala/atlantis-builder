package build

import (
	"atlantis/builder/docker"
	"errors"
	"fmt"
	"path"
	"sync"
	"time"
)

type Layers struct {
	Version       string
	BaseLayer     string
	BuilderLayers []string
}

func (l *Layers) builderLayerName(appType string) string {
	return fmt.Sprintf("builder/%s-%s-%s", l.BaseLayer, appType, l.Version)
}

func (l *Layers) BuilderLayerName(appType string) (string, error) {
	for _, t := range l.BuilderLayers {
		if t == appType {
			return l.builderLayerName(appType), nil
		}
	}
	return "", errors.New("app type not supported!")
}

func (l *Layers) BaseLayerName() string {
	return fmt.Sprintf("base/%s-%s", l.BaseLayer, l.Version)
}

func Boot(client *docker.Client, overlayDir string, layers *Layers) {
	fmt.Println("Now building ...")
	var wg sync.WaitGroup

	builderLayers := path.Join(overlayDir, "builder")
	for _, appType := range layers.BuilderLayers {
		wg.Add(1)
		go func(myType string) {
			fmt.Printf("\tstart %s -> %s\n", layers.BaseLayerName(), layers.builderLayerName(myType))
			client.OverlayAndCommit(layers.BaseLayerName(), layers.builderLayerName(myType), path.Join(builderLayers, myType), "/overlay", 10*time.Minute, "/overlay/sbin/provision_type", "/overlay")
			client.PushImage(layers.builderLayerName(myType), false)
			fmt.Printf("\tdone %s\n ", layers.builderLayerName(myType))
			wg.Done()
		}(appType)
	}
	wg.Wait()
}
