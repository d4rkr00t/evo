package lib

import (
	"path"
	"path/filepath"
	"scu/main/lib/cache"
	"scu/main/lib/fileutils"
	"sort"
	"strings"
)

type Workspace struct {
	Name     string
	Path     string
	RelPath  string
	Deps     map[string]string
	Includes []string
	Excludes []string
}

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

func (w Workspace) Invalidate(cmd string, cc *cache.Cache) (bool, string) {
	var ws_hash = w.Hash()
	if cc.ReadData(w.GetStateKey(cmd)) == ws_hash {
		return false, ws_hash
	}
	return true, ws_hash
}

func (w Workspace) Hash() string {
	var files = w.get_files()
	return fileutils.GetFileListHash(files)
}

func (w Workspace) GetStateKey(cmd string) string {
	return cmd + ":" + strings.Replace(w.RelPath+"__"+w.Name, "/", "__", -1)
}

func (w Workspace) Cache(c *cache.Cache, hash string) {
	var ignores = cache.CacheDirIgnores{
		"node_modules": true,
	}
	c.CacheDir(hash, w.Path, ignores)
}

func (w Workspace) CacheState(c *cache.Cache, cmd string, hash string) {
	c.CacheData(w.GetStateKey(cmd), hash)
}

func (w Workspace) get_files() []string {
	var files []string = fileutils.GlobFiles(w.Path, &w.Includes, &w.Excludes)
	sort.Strings(files)
	return files
}
