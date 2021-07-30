package lib

import (
	"fmt"
	"scu/main/lib/cache"
	"sync"
)

type Runner struct {
	cwd     string
	project Project
	cache   cache.Cache
}

func NewRunner(cwd string) Runner {
	var cc = cache.NewCache(cwd)
	var proj = NewProject(cwd)
	return Runner{cwd: cwd, project: proj, cache: cc}
}

func (r Runner) GetCwd() string {
	return r.cwd
}

func (r Runner) Build() {
	fmt.Println("\nBuild:", r.GetCwd())
	var updated = r.project.Invalidate(make([]string, 0), r.cache)
	fmt.Println("Updated:", len(updated), "of", len(r.project.Workspaces))
	if len(updated) > 0 {
		fmt.Println(updated)
	}

	var wg sync.WaitGroup
	for ws_hash, ws_name := range updated {
		wg.Add(1)
		go func(ws_name string, ws_hash string) {
			fmt.Println("Compiling:", ws_name, ws_hash)
			var ws = r.project.Workspaces[ws_name]
			if r.cache.Has(ws_hash) {
				fmt.Println("Cache hit:", ws_name, ws_hash)
				r.cache.RestoreDir(ws_hash, ws.Path)
			} else {
				ws.Cache(&r.cache, ws_hash)
			}
			ws.CacheState(&r.cache, ws_hash)
			fmt.Println("Done:", ws_name)
			wg.Done()
		}(ws_name, ws_hash)
	}
	wg.Wait()
}
