package build

import (
	"atlantis/builder/docker"
	"atlantis/builder/layers"
	"fmt"
	"path"
	"sync"
	"time"
)

func Boot(client *docker.Client, overlayDir string, l *layers.Layers) {
	fmt.Println("Now building ...")
	var wg sync.WaitGroup

	builderLayers := path.Join(overlayDir, "builder")
	for _, appType := range l.BuilderLayers {
		wg.Add(1)
		go func(myType string) {
			fmt.Printf("\tstart %s -> %s\n", l.BaseLayerName(), l.BuilderLayerNameUnsafe(myType))
			client.OverlayAndCommit(l.BaseLayerName(), l.BuilderLayerNameUnsafe(myType),
				path.Join(builderLayers, myType), "/overlay", 10*time.Minute, "/overlay/sbin/provision_type",
				"/overlay")
			client.PushImage(l.BuilderLayerNameUnsafe(myType), false)
			fmt.Printf("\tdone %s\n ", l.BuilderLayerNameUnsafe(myType))
			wg.Done()
		}(appType)
	}
	wg.Wait()
}
