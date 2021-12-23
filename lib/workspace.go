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
	Name        string
	Path        string
	RelPath     string
	Deps        map[string]string
	Excludes    []string
	Outputs     []string
	Rules       map[string]Rule
	FilesHash   string
	RulesHash   string
	ExtDepsHash string
	PkgJson     PackageJson
	cache       *cache.Cache
	hash        string
}

func NewWorkspace(root_path string, ws_path string, excludes []string, cc *cache.Cache, rules map[string]Rule) Workspace {
	var package_json_path = path.Join(ws_path, "package.json")
	var package_json, _ = NewPackageJson(package_json_path)
	var rel_path, _ = filepath.Rel(root_path, ws_path)

	var Deps = package_json.Dependencies

	if Deps == nil {
		Deps = map[string]string{}
	}

	for dep, ver := range package_json.DevDependencies {
		Deps[dep] = ver
	}

	var outputs = []string{}

	for _, rule := range rules {
		outputs = append(outputs, rule.Outputs...)
	}

	excludes = append(excludes, outputs...)

	var files []string = fileutils.GlobFiles(ws_path, &[]string{}, &excludes)
	sort.Strings(files)
	var fileshash = fileutils.GetFileListHash(files)

	return Workspace{
		Name:      package_json.Name,
		Path:      ws_path,
		RelPath:   rel_path,
		Deps:      Deps,
		Excludes:  excludes,
		Rules:     rules,
		Outputs:   outputs,
		FilesHash: fileshash,
		RulesHash: get_rules_hash(rules),
		PkgJson:   package_json,
		cache:     cc,
	}
}

func (w Workspace) GetRule(name string) (Rule, bool) {
	var rule, ok = w.Rules[name]
	return rule, ok
}

func (w *Workspace) GetHashForTask() string {
	return HashStringList([]string{
		w.ExtDepsHash,
		w.FilesHash,
	})
}

func (w *Workspace) Rehash(workspaces *WorkspacesMap) string {
	var depshash = w.get_deps_hash(workspaces)
	w.ExtDepsHash = w.get_ext_deps_hash(workspaces)
	w.hash = HashStringList([]string{
		depshash,
		w.ExtDepsHash,
		w.FilesHash,
		w.RulesHash,
	})
	return w.hash
}

func (w *Workspace) GetStateKey() string {
	return ClearTaskName(w.Name)
}

func (w Workspace) CacheState(c *cache.Cache, ws_hash string) {
	c.CacheData(w.GetStateKey(), ws_hash)
}

func (w Workspace) GetCacheState() string {
	return w.cache.ReadData(w.GetStateKey())
}

func (w Workspace) get_files() []string {
	var files []string = fileutils.GlobFiles(w.Path, &[]string{}, &w.Excludes)
	sort.Strings(files)
	return files
}

func (w Workspace) get_deps_hash(wm *WorkspacesMap) string {
	var deps_list = []string{}

	for dep_name := range w.Deps {
		if ws, ok := wm.Load(dep_name); ok {
			deps_list = append(deps_list, dep_name+":"+ws.hash)
		}
	}

	sort.Strings(deps_list)
	return HashStringList(deps_list)
}

func (w Workspace) get_ext_deps_hash(wm *WorkspacesMap) string {
	var deps_list = []string{}

	for dep_name, dep_version := range w.Deps {
		if _, ok := wm.Load(dep_name); ok {
			deps_list = append(deps_list, dep_name+":"+dep_version)
		}
	}

	sort.Strings(deps_list)
	return HashStringList(deps_list)
}

func get_rules_hash(rules map[string]Rule) string {
	var h = sha1.New()
	var rules_list = get_rules_names(&rules)

	for _, rule := range rules_list {
		io.WriteString(h, rule)
	}

	return hex.EncodeToString(h.Sum(nil))
}

func get_rules_names(Rules *map[string]Rule) []string {
	var rules_list = []string{}
	var rules = []string{}

	for rule_name := range *Rules {
		rules_list = append(rules_list, rule_name)
	}

	sort.Strings(rules_list)

	for _, rule_name := range rules_list {
		rules = append(rules, (*Rules)[rule_name].String())
	}

	return rules
}
