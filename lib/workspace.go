package lib

import (
	"path"
	"path/filepath"
	"scu/main/lib/cache"
	"scu/main/lib/fileutils"
	"sort"
	"sync"
)

type Workspace struct {
	Name     string
	Path     string
	RelPath  string
	Deps     map[string]string
	Includes []string
	Excludes []string
}

type WorkspacesMap = map[string]Workspace

func NewWorkspace(project_path string, ws_path string, includes []string, excludes []string) Workspace {
	var package_json_path = path.Join(ws_path, "package.json")
	var package_json = NewPackageJson(package_json_path)
	var rel_path, _ = filepath.Rel(project_path, ws_path)

	return Workspace{
		Name:     package_json.Name,
		Path:     ws_path,
		RelPath:  rel_path,
		Deps:     package_json.Dependencies,
		Includes: includes,
		Excludes: excludes,
	}
}

func (w Workspace) Hash() string {
	var files = w.get_files()
	return fileutils.GetFileListHash(files)
}

func (w Workspace) Cache(c *cache.Cache, cache_key string) {
	var ignores = cache.CacheDirIgnores{
		"node_modules": true,
	}
	c.CacheDir(cache_key, w.Path, ignores)
}

func (w Workspace) get_files() []string {
	var files []string = fileutils.GlobFiles(w.Path, &w.Includes, &w.Excludes)
	sort.Strings(files)
	return files
}

func GetWorkspaces(cwd string, conf *Config) WorkspacesMap {
	var workspaces = make(map[string]Workspace)
	var wg sync.WaitGroup
	var queue = make(chan Workspace)

	for _, wc := range conf.Workspaces {
		var ws_glob = path.Join(cwd, wc, "package.json")
		var matches, _ = filepath.Glob(ws_glob)
		for _, ws_path := range matches {
			wg.Add(1)
			go func(ws_path string) {
				var includes, excludes = conf.GetInputs(path.Dir(ws_path))
				queue <- NewWorkspace(cwd, path.Dir(ws_path), includes, excludes)
			}(ws_path)
		}
	}

	go func() {
		for ws := range queue {
			workspaces[ws.Name] = ws
			wg.Done()
		}
	}()

	wg.Wait()

	return workspaces
}

func InvalidateWorkspaces(workspaces *WorkspacesMap, target string, cc *cache.Cache) map[string]string {
	var updated = map[string]string{}
	var wg sync.WaitGroup
	var queue = make(chan []string)

	for name, ws := range *workspaces {
		wg.Add(1)
		go func(name string, ws Workspace) {
			var ws_hash = ws.Hash()
			var state_key = ClearTaskName(GetTaskName(target, ws.Name))
			if cc.ReadData(state_key) == ws_hash {
				queue <- []string{}
			} else {
				queue <- []string{name, ws_hash}
			}
		}(name, ws)
	}

	go func() {
		for dat := range queue {
			if len(dat) > 0 {
				updated[dat[0]] = dat[1]
			}
			wg.Done()
		}
	}()

	wg.Wait()

	return updated
}
