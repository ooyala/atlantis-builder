package api

import (
	"atlantis/builder/api/types"
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
	"runtime"
	"sync"
)

// this will ensure only one build happens at a time
var buildLock = sync.Mutex{}

type Build struct {
	types.Build
	client      *docker.Client
	layerPath   string
	manifestDir string
}

func (b *Build) Run() {
	buildLock.Lock()
	if err := os.MkdirAll(b.manifestDir, 0755); err != nil {
		b.Error = err
		b.Status = types.StatusError
		buildLock.Unlock()
		return
	}
	// catch panic
	defer func() {
		if err := recover(); err != nil {
			log.Printf("Error building "+b.URL+"/"+b.RelPath+"@"+b.Sha+": %v", err)
			// print stack so we can trace the error when it happens
			buf := []byte{}
			runtime.Stack(buf, false)
			fmt.Println(string(buf))
			// return an error to the client
			b.Error = err
			b.Status = types.StatusError
		} else {
			b.Status = types.StatusDone
		}
	}()
	defer buildLock.Unlock()
	b.Status = types.StatusBuilding
	build.App(b.client, b.URL, b.Sha, b.RelPath, b.manifestDir, layers.ReadLayerInfo(b.layerPath))
}

type Boot struct {
	types.Boot
	client    *docker.Client
	layerPath string
}

func (b *Boot) Run() {
	b.Status = types.StatusBooting

	fi, err := os.Stat(b.layerPath)
	if err == nil && fi.IsDir() {
		build.Boot(b.client, b.layerPath, layers.ReadLayerInfo(b.layerPath))
	} else {
		b.Error = errors.New(b.layerPath + " does not exist or not a directory")
		b.Status = types.StatusError
		return
	}

	// catch panic
	defer func() {
		if err := recover(); err != nil {
			b.Error = err
			b.Status = types.StatusError
		} else {
			b.Status = types.StatusDone
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
		Boot: types.Boot{
			Status: types.StatusInit,
		},
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
	body, err := json.Marshal(b.boot.Boot)
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

	var tbuild types.Build
	if err := json.NewDecoder(r.Body).Decode(&tbuild); err != nil {
		http.Error(w, "Error decoding theBuilduest: "+err.Error(), http.StatusBadRequest)
		return
	}
	if tbuild.URL == "" || tbuild.Sha == "" || tbuild.RelPath == "" {
		http.Error(w, "provide url, sha, and rel path!", http.StatusBadRequest)
		return
	}

	tbuild.Status = types.StatusInit
	theBuild := Build{
		Build: tbuild,
	}
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

	body, err := json.Marshal(&theBuild.Build)
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
	body, err := json.Marshal(theBuild.Build)
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
	if theBuild.Status == types.StatusError {
		http.Error(w, "Build Ended with Error", http.StatusBadRequest)
		return
	}
	if theBuild.Status != types.StatusDone {
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
