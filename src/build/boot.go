package build

import (
	"docker"
	"errors"
	"fmt"
	"path"
	"sync"
	"time"
)

const overlayDir = "/opt/atlantis/builder/layers/builder"

type Layers struct {
	Version       string
	BaseLayer     string
	BuilderLayers []string
}

func (l *Layers) builderLayerName(appType string) string {
	return fmt.Sprintf("builder/%s/%s-%s", l.BaseLayer, appType, l.Version)
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

func Boot(client *docker.Client, layers *Layers) {
	fmt.Println("Now building ...")
	var wg sync.WaitGroup

	for _, appType := range layers.BuilderLayers {
		wg.Add(1)
		go func(myType string) {
			fmt.Println("\tstart " + myType)
			client.OverlayAndCommit(layers.BaseLayerName(), layers.builderLayerName(myType), path.Join(overlayDir, myType), "/overlay", 10*time.Minute, "/overlay/sbin/provision_type", "/overlay")
			fmt.Println("\tdone " + myType)
			wg.Done()
		}(appType)
	}
	wg.Wait()
}
