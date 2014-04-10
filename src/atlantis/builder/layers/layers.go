package layers

import (
	"errors"
	"fmt"
	"io/ioutil"
	"path"
	"strings"
)

type Layers struct {
	Version       string
	BaseLayer     string
	BuilderLayers []string
}

func (l *Layers) BuilderLayerNameUnsafe(appType string) string {
	return fmt.Sprintf("builder/%s-%s-%s", l.BaseLayer, appType, l.Version)
}

func (l *Layers) BuilderLayerName(appType string) (string, error) {
	for _, t := range l.BuilderLayers {
		if t == appType {
			return l.BuilderLayerNameUnsafe(appType), nil
		}
	}
	return "", errors.New("app type not supported!")
}

func (l *Layers) BaseLayerName() string {
	return fmt.Sprintf("base/%s-%s", l.BaseLayer, l.Version)
}

func ReadLayerInfo(overlayDir string) *Layers {
	baseFile := path.Join(overlayDir, "basename.txt")
	baseName, err := ioutil.ReadFile(baseFile)
	if err != nil {
		panic(err)
	}

	versionFile := path.Join(overlayDir, "version.txt")
	versionNumber, err := ioutil.ReadFile(versionFile)
	if err != nil {
		panic(err)
	}

	layersDir := path.Join(overlayDir, "builder")
	dirs, err := ioutil.ReadDir(layersDir)
	if err != nil {
		panic(err)
	}

	layerNames := []string{}
	for _, dir := range dirs {
		layerNames = append(layerNames, dir.Name())
	}

	return &Layers{
		Version:       strings.TrimRight(string(versionNumber), "\n"),
		BaseLayer:     strings.TrimRight(string(baseName), "\n"),
		BuilderLayers: layerNames,
	}
}
