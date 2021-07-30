package lib

import (
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
	w.CacheState(c, hash)
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
