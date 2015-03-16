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
				path.Join(builderLayers, myType), "/overlay", 100*time.Minute, "/overlay/sbin/provision_type",
				"/overlay")
			client.PushImage(l.BuilderLayerNameUnsafe(myType), false)
			fmt.Printf("\tdone %s\n ", l.BuilderLayerNameUnsafe(myType))
			wg.Done()
		}(appType)
	}
	wg.Wait()
}
