package lib

import (
	"crypto/sha1"
	"encoding/hex"
	"io"
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
	Rules    map[string]Rule
	cache    *cache.Cache
}

type WorkspacesMap = map[string]Workspace

func NewWorkspace(project_path string, ws_path string, includes []string, excludes []string, cc *cache.Cache, rules map[string]Rule) Workspace {
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
		Rules:    rules,
		cache:    cc,
	}
}

func (w Workspace) GetRule(name string) Rule {
	return w.Rules[name]
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

func (w Workspace) get_deps_hash(workspaces *WorkspacesMap) string {
	var h = sha1.New()
	var deps_list = []string{}

	for dep_name, dep_version := range w.Deps {
		if dep_ws, ok := (*workspaces)[dep_name]; ok {
			deps_list = append(deps_list, dep_name+":"+dep_ws.GetCacheState())
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
	var rules_list = []string{}

	for rule_name, _ := range w.Rules {
		rules_list = append(rules_list, rule_name)
	}

	sort.Strings(rules_list)

	for _, rule_name := range rules_list {
		io.WriteString(h, w.Rules[rule_name].String())
	}

	return hex.EncodeToString(h.Sum(nil))
}

func GetWorkspaces(cwd string, conf *Config, cc *cache.Cache) WorkspacesMap {
	var workspaces = make(map[string]Workspace)
	var wg sync.WaitGroup
	var queue = make(chan Workspace)

	for _, wc := range conf.Workspaces {
		var ws_glob = path.Join(cwd, wc, "package.json")
		var matches, _ = filepath.Glob(ws_glob)
		for _, ws_path := range matches {
			wg.Add(1)
			go func(ws_path string) {
				var includes, excludes = conf.GetInputs(ws_path)
				var rules = conf.GetAllRulesForWS(ws_path)
				queue <- NewWorkspace(cwd, ws_path, includes, excludes, cc, rules)
			}(path.Dir(ws_path))
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
			var ws_hash = ws.Hash(workspaces)
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
