package workspace

import (
	"encoding/json"
	"evo/internal/cache"
	"evo/internal/fsutils"
	"evo/internal/hash_utils"
	"evo/internal/target"
	"sort"
	"strings"
	"sync"
)

type WorkspaceDependency struct {
	Name     string
	Version  string
	Provider string
	Type     string
}

type WorkspaceDependencyMap = map[string]WorkspaceDependency

type Workspace struct {
	Name          string
	Path          string
	Deps          WorkspaceDependencyMap
	Excludes      []string
	Outputs       []string
	Targets       target.TargetsMap
	Files         []string
	FilesHash     string
	ExtDepsHash   string
	LocalDepsHash string
	Hash          string
}

func New(name string, wsAbsPath string, targets target.TargetsMap, excludes []string) *Workspace {
	var targetsOutputs = target.GetAllTargetsOutputs(&targets)

	return &Workspace{
		Name:     name,
		Path:     wsAbsPath,
		Deps:     WorkspaceDependencyMap{},
		Targets:  targets,
		Outputs:  targetsOutputs,
		Excludes: append(excludes, targetsOutputs...),
	}
}

func (ws *Workspace) GetFilesList() []string {
	var files []string = fsutils.GlobFiles(ws.Path, &[]string{}, &ws.Excludes)
	sort.Strings(files)
	return files
}

func (ws *Workspace) Rehash(wm *sync.Map) {
	ws.Files = ws.GetFilesList()
	ws.FilesHash = fsutils.GetFileListHash(ws.Files)
	ws.LocalDepsHash = ws.getLocalDepsHash(wm)
	ws.ExtDepsHash = ws.getExternalDepsHash()

	ws.Hash = hash_utils.HashStringList([]string{
		ws.FilesHash,
		ws.ExtDepsHash,
	})
}

func (ws *Workspace) RetriveStateFromCache(c *cache.Cache) (Workspace, error) {
	var key = ws.CleanName() + "__state"
	if !c.Has(key) {
		return *ws, nil
	}
	var data = c.ReadData(key)
	var cachedWs Workspace
	var err = json.Unmarshal([]byte(data), &cachedWs)
	return cachedWs, err
}

func (ws *Workspace) CacheState(c *cache.Cache) {
	var key = ws.CleanName() + "__state"
	var data, err = json.Marshal(ws)
	if err == nil {
		c.CacheData(key, string(data))
	} else {
		println(err.Error())
	}
}

func (ws *Workspace) CleanName() string {
	return strings.Replace(ws.Name, "/", "__", -1)
}

func (ws *Workspace) getLocalDepsHash(wm *sync.Map) string {
	var depsList = []string{}

	for _, dep := range ws.Deps {
		if dep.Type == "local" {
			depsList = append(depsList, dep.Name+":"+dep.Provider)
		}
	}

	sort.Strings(depsList)
	return hash_utils.HashStringList(depsList)
}

func (ws *Workspace) getExternalDepsHash() string {
	var depsList = []string{}

	for _, dep := range ws.Deps {
		if dep.Type == "external" {
			depsList = append(depsList, dep.Name+":"+dep.Version+":"+dep.Provider)
		}
	}

	sort.Strings(depsList)
	return hash_utils.HashStringList(depsList)
}
