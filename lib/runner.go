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
	fmt.Println("\n===============\n")

	var updated = r.project.Invalidate(make([]string, 0), r.cache)
	var affected = r.project.GetAffected(&updated)

	fmt.Println("\nUpdated:", len(updated), "of", len(r.project.Workspaces), updated)

	if len(affected) > 0 {
		fmt.Println("\nAffected:", affected)
	}

	fmt.Println("\n===============\n")

	for ws_name, ws_hash := range affected {
		updated[ws_name] = ws_hash
	}

	var wg sync.WaitGroup
	for ws_name, ws_hash := range updated {
		wg.Add(1)
		go func(ws_name string, ws_hash string) {
			fmt.Println("Compiling:", ws_name, ws_hash)

			var ws = r.project.GetWs(ws_name)
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

	if len(updated) > 0 {
		fmt.Println("\n===============\n")
	}
}
