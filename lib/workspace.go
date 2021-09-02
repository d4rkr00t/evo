package lib

import (
	"crypto/sha1"
	"encoding/hex"
	"evo/main/lib/cache"
	"evo/main/lib/fileutils"
	"io"
	"path"
	"path/filepath"
	"sort"
)

type Workspace struct {
	Name     string
	Path     string
	RelPath  string
	Deps     map[string]string
	Includes []string
	Excludes []string
	Rules    map[string]Rule
	cache    *cache.Cache
}

func NewWorkspace(root_path string, ws_path string, includes []string, excludes []string, cc *cache.Cache, rules map[string]Rule) Workspace {
	var package_json_path = path.Join(ws_path, "package.json")
	var package_json = NewPackageJson(package_json_path)
	var rel_path, _ = filepath.Rel(root_path, ws_path)

	var Deps = package_json.Dependencies

	for dep, ver := range package_json.DevDependencies {
		Deps[dep] = ver
	}

	return Workspace{
		Name:     package_json.Name,
		Path:     ws_path,
		RelPath:  rel_path,
		Deps:     package_json.Dependencies,
		Includes: includes,
		Excludes: excludes,
		Rules:    rules,
		cache:    cc,
	}
}

func (w Workspace) GetRule(name string) (Rule, bool) {
	var rule, ok = w.Rules[name]
	return rule, ok
}

func (w Workspace) Hash(workspaces *WorkspacesMap) string {
	var files = w.get_files()
	var fileshash = fileutils.GetFileListHash(files)
	var depshash = w.get_deps_hash(workspaces)
	var ruleshash = w.get_rules_hash()
	var h = sha1.New()
	io.WriteString(h, depshash+":"+fileshash+":"+ruleshash)
	return hex.EncodeToString(h.Sum(nil))
}

func (w Workspace) GetStateKey() string {
	return ClearTaskName(w.Name)
}

func (w Workspace) CacheState(c *cache.Cache, ws_hash string) {
	c.CacheData(w.GetStateKey(), ws_hash)
}

func (w Workspace) GetCacheState() string {
	return w.cache.ReadData(w.GetStateKey())
}

func (w Workspace) get_files() []string {
	var files []string = fileutils.GlobFiles(w.Path, &w.Includes, &w.Excludes)
	sort.Strings(files)
	return files
}

func (w Workspace) get_deps_hash(wm *WorkspacesMap) string {
	var h = sha1.New()
	var deps_list = []string{}

	for dep_name, dep_version := range w.Deps {
		if ws, ok := wm.workspaces[dep_name]; ok {
			var hash, ok = wm.hashes[dep_name]
			if !ok {
				hash = ws.GetCacheState()
			}
			deps_list = append(deps_list, dep_name+":"+hash)
		} else {
			deps_list = append(deps_list, dep_name+":"+dep_version)
		}
	}

	sort.Strings(deps_list)

	for _, dep := range deps_list {
		io.WriteString(h, dep)
	}

	return hex.EncodeToString(h.Sum(nil))
}

func (w Workspace) get_rules_hash() string {
	var h = sha1.New()
	var rules_list = w.get_rules_names()

	for _, rule := range rules_list {
		io.WriteString(h, rule)
	}

	return hex.EncodeToString(h.Sum(nil))
}

func (w Workspace) get_rules_names() []string {
	var rules_list = []string{}
	var rules = []string{}

	for rule_name := range w.Rules {
		rules_list = append(rules_list, rule_name)
	}

	sort.Strings(rules_list)

	for _, rule_name := range rules_list {
		rules = append(rules, w.Rules[rule_name].String())
	}

	return rules
}
