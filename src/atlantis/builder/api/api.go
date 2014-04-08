package api

import (
	"atlantis/builder/build"
	"atlantis/builder/docker"
	"atlantis/builder/layers"
	"atlantis/common"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"sync"
)

const (
	StatusInit  = "INIT"
	StatusDone  = "DONE"
	StatusError = "ERROR"
)

type Build struct {
	client      *docker.Client
	layerPath   string
	manifestDir string
	ID          string
	URL         string
	Sha         string
	RelPath     string
	Status      string
	Error       interface{}
}

func (b *Build) Run() {
	if err := os.MkdirAll(b.manifestDir, 0755); err != nil {
		b.Error = err
		b.Status = StatusError
		return
	}
	b.Status = "Building..."
	build.App(b.client, b.URL, b.Sha, b.RelPath, b.manifestDir, layers.ReadLayerInfo(b.layerPath))
	// catch panic
	defer func() {
		if err := recover(); err != nil {
			b.Error = err
			b.Status = StatusError
		} else {
			b.Status = StatusDone
		}
	}()
}

type Boot struct {
	client    *docker.Client
	layerPath string
	Status    string
	Error     interface{}
}

func (b *Boot) Run() {
	b.Status = "Booting..."

	fi, err := os.Stat(b.layerPath)
	if err == nil && fi.IsDir() {
		build.Boot(b.client, b.layerPath, layers.ReadLayerInfo(b.layerPath))
	} else {
		b.Error = errors.New(b.layerPath + " does not exist or not a directory")
		b.Status = StatusError
		return
	}

	// catch panic
	defer func() {
		if err := recover(); err != nil {
			b.Error = err
			b.Status = StatusError
		} else {
			b.Status = StatusDone
		}
	}()
}

type BuilderAPI struct {
	sync.RWMutex
	client          *docker.Client
	builds          map[string]*Build
	building        map[string]bool // "<url><sha><rel>" -> true
	booting         bool
	boot            *Boot
	Port            uint16
	LayerPath       string
	ManifestBaseDir string
}

func New(port uint16, registry, layerPath, manifestBaseDir string) *BuilderAPI {
	return &BuilderAPI{
		client:          docker.New(registry),
		Port:            port,
		LayerPath:       layerPath,
		ManifestBaseDir: manifestBaseDir,
	}
}

func (b *BuilderAPI) Run() {
	r := mux.NewRouter()
	r.HandleFunc("/boot", b.PostBootHandler).Methods("POST")
	r.HandleFunc("/boot", b.GetBootHandler).Methods("GET")
	r.HandleFunc("/build", b.PostBuildHandler).Methods("POST")
	r.HandleFunc("/build/{id}", b.GetBuildHandler).Methods("GET")
	r.HandleFunc("/build/{id}/manifest", b.GetManifestHandler).Methods("GET")
	s := &http.Server{
		Addr:    fmt.Sprintf(":%d", b.Port),
		Handler: r,
	}
	log.Fatal(s.ListenAndServe())
}

func (b *BuilderAPI) PostBootHandler(w http.ResponseWriter, r *http.Request) {
	b.Lock()
	if b.booting {
		http.Error(w, "Already Booting", http.StatusConflict)
		b.Unlock()
		return
	}
	b.booting = true
	b.boot = &Boot{
		client:    b.client,
		layerPath: b.LayerPath,
		Status:    StatusInit,
	}
	b.Unlock()

	go func() {
		b.boot.Run()
		b.Lock()
		b.booting = false
		b.Unlock()
	}()

	b.GetBootHandler(w, r)
}

func (b *BuilderAPI) GetBootHandler(w http.ResponseWriter, r *http.Request) {
	b.RLock()
	defer b.RUnlock()
	body, err := json.Marshal(b.boot)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(body)
}

func (b *BuilderAPI) PostBuildHandler(w http.ResponseWriter, r *http.Request) {
	b.RLock()
	if b.booting {
		http.Error(w, "Boot in progress", http.StatusConflict)
		b.RUnlock()
		return
	}
	b.RUnlock()

	var theBuild Build
	if err := json.NewDecoder(r.Body).Decode(&theBuild); err != nil {
		http.Error(w, "Error decoding theBuilduest: "+err.Error(), http.StatusBadRequest)
		return
	}
	if theBuild.URL == "" || theBuild.Sha == "" || theBuild.RelPath == "" {
		http.Error(w, "provide url, sha, and rel path!", http.StatusBadRequest)
		return
	}

	theBuild.Status = StatusInit
	if err := b.reserveBuild(&theBuild); err != nil {
		// can't create ID, must be a conflict
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}

	b.RLock()
	theBuild.client = b.client
	theBuild.layerPath = b.LayerPath
	theBuild.manifestDir = path.Join(b.ManifestBaseDir, theBuild.ID)
	b.RUnlock()

	go func() {
		theBuild.Run()
		b.releaseBuild(&theBuild)
	}()

	body, err := json.Marshal(&theBuild)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(body)
}

func (b *BuilderAPI) GetBuildHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	b.RLock()
	defer b.RUnlock()
	theBuild := b.builds[vars["ID"]]
	if theBuild == nil {
		http.Error(w, "No such build", http.StatusNotFound)
		return
	}
	body, err := json.Marshal(theBuild)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(body)
}

func (b *BuilderAPI) GetManifestHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	b.RLock()
	defer b.RUnlock()
	theBuild := b.builds[vars["ID"]]
	if theBuild == nil {
		http.Error(w, "No such build", http.StatusNotFound)
		return
	}
	if theBuild.Status == StatusError {
		http.Error(w, "Build Ended with Error", http.StatusBadRequest)
		return
	}
	if theBuild.Status != StatusDone {
		http.Error(w, "Build Not Finished", http.StatusBadRequest)
		return
	}
	manFile, err := os.Open(path.Join(theBuild.manifestDir, "manifest.toml"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer manFile.Close()
	io.Copy(w, manFile)
}

func (b *BuilderAPI) reserveBuild(r *Build) error {
	b.Lock()
	defer b.Unlock()
	concat := r.URL + r.Sha + r.RelPath
	// verify this url/sha/rel combo isn't currently being built
	if b.building[concat] {
		return errors.New("A build for this URL+Sha+RelPath in currently in progress")
	}
	b.building[concat] = true
	// reserve new ID
	for r.ID = common.CreateRandomID(20); b.builds[r.ID] != nil; r.ID = common.CreateRandomID(20) {
		// loop handles everything
	}
	b.builds[r.ID] = r
	return nil
}

func (b *BuilderAPI) releaseBuild(r *Build) {
	b.Lock()
	defer b.Unlock()
	concat := r.URL + r.Sha + r.RelPath
	delete(b.building, concat)
}
