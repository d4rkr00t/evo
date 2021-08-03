package lib

import (
	"fmt"
	"path"
	"path/filepath"
	"scu/main/lib/cache"
	"scu/main/lib/fileutils"
	"scu/main/lib/typescript"
	"sort"
	"strings"
)

type Workspace struct {
	Name    string
	Path    string
	RelPath string
	Deps    map[string]string
	Meta    WorkspaceMeta
}

type WorkspaceMeta struct {
	IsTS bool
}

func NewWorkspace(project_path string, ws_path string) Workspace {
	var package_json_path = path.Join(ws_path, "package.json")
	var package_json = NewPackageJson(package_json_path)
	var rel_path, _ = filepath.Rel(project_path, ws_path)

	return Workspace{
		Name:    package_json.Name,
		Path:    ws_path,
		RelPath: rel_path,
		Deps:    package_json.Dependencies,
		Meta: WorkspaceMeta{
			IsTS: typescript.IsTypeScriptWS(ws_path),
		},
	}
}

func (w Workspace) Hash() string {
	var files = w.get_files()
	return fileutils.GetFileListHash(files)
}

func (w Workspace) GetStateKey() string {
	return strings.Replace(w.RelPath+"__"+w.Name, "/", "__", -1)
}

func (w Workspace) Cache(c *cache.Cache, hash string) {
	var ignores = cache.CacheDirIgnores{
		"node_modules": true,
	}
	c.CacheDir(hash, w.Path, ignores)
}

func (w Workspace) CacheState(c *cache.Cache, hash string) {
	c.CacheData(w.GetStateKey(), hash)
}

func (w Workspace) get_files() []string {
	var files []string = []string{path.Join(w.Path, "package.json")}

	if w.Meta.IsTS {
		files = append(files, typescript.GetTSConfigPath(w.Path))
		files = append(files, typescript.GetFilesFromTSConfig(w.Path)...)
	}

	sort.Strings(files)

	return files
}

func (w Workspace) CreateBuildTask(affected *map[string]string, updated *map[string]string) Task {
	var task_name = w.Name + ":build"
	var deps = []string{}

	for dep := range w.Deps {
		if _, ok := (*affected)[dep]; ok {
			deps = append(deps, dep+":build")
		}
	}

	return NewTask(
		w.Name,
		task_name,
		deps,
		func(c *cache.Cache) {
			var ws_hash = (*affected)[w.Name]
			fmt.Println("Compiling:", w.Name, "->", task_name, "->", ws_hash)
			var _, was_updated = (*updated)[w.Name]

			if was_updated {
				if c.Has(ws_hash) {
					fmt.Println("Cache hit:", w.Name, ws_hash)
					c.RestoreDir(ws_hash, w.Path)
				} else {
					w.Cache(c, ws_hash)
				}
			} else {
				fmt.Println("Force compiling updated deps:", w.Name, ws_hash)
				w.Cache(c, ws_hash)
			}

			w.CacheState(c, ws_hash)
		},
	)
}
