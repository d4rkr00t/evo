package lib

import (
	"fmt"
	"scu/main/lib/cache"
	"sync"

	"github.com/davecgh/go-spew/spew"
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
	fmt.Println("\n===============")
	fmt.Println("")

	var updated = r.project.Invalidate(make([]string, 0), r.cache)
	// var affected = r.project.GetAffected(&updated)

	fmt.Println("\nUpdated:", len(updated), "of", len(r.project.Workspaces), updated)

	// if len(affected) > 0 {
	// 	fmt.Println("\nAffected:", affected)
	// }

	var tasks = r.create_tasks(&updated)
	fmt.Println("Build tasks:")
	spew.Dump(tasks)
	r.run_tasks(&tasks)

	// for ws_name, ws_hash := range affected {
	// 	updated[ws_name] = ws_hash
	// }

	// var wg sync.WaitGroup
	// for ws_name, ws_hash := range updated {
	// 	wg.Add(1)
	// 	go func(ws_name string, ws_hash string) {
	// 		fmt.Println("Compiling:", ws_name, ws_hash)

	// 		var ws = r.project.GetWs(ws_name)
	// 		if r.cache.Has(ws_hash) {
	// 			fmt.Println("Cache hit:", ws_name, ws_hash)
	// 			r.cache.RestoreDir(ws_hash, ws.Path)
	// 		} else {
	// 			ws.Cache(&r.cache, ws_hash)
	// 		}

	// 		ws.CacheState(&r.cache, ws_hash)
	// 		fmt.Println("Done:", ws_name)
	// 		wg.Done()
	// 	}(ws_name, ws_hash)
	// }

	// wg.Wait()

	if len(updated) > 0 {
		fmt.Println("\n===============")
		fmt.Println("")
	}
}

func (r Runner) create_tasks(workspaces *map[string]string) map[string]Task {
	var tasks = map[string]Task{}
	var affected = r.project.GetAffected(workspaces)

	for ws_name := range affected {
		var task_name = ws_name + ":build"
		var deps = []string{}

		for dep := range r.project.GetWs(ws_name).Deps {
			if _, ok := affected[dep]; ok {
				deps = append(deps, dep)
			}
		}

		tasks[task_name] = NewTask(ws_name, task_name, append([]string{}, deps...))
		tasks[ws_name] = NewTask(ws_name, ws_name, append(append([]string{}, deps...), task_name))
	}

	return tasks
}

func (r Runner) run_tasks(tasks *map[string]Task) {
	var in_progress_queue = make(chan string)
	var wg sync.WaitGroup
	var mu sync.Mutex

	for task_id, task := range *tasks {
		if len(task.Deps) == 0 {
			fmt.Println(task_id, task.Deps)
			wg.Add(1)
			task.status = TASK_STATUS_RUNNING
			(*tasks)[task_id] = task
			go func(id string) {
				in_progress_queue <- id
			}(task_id)
		}
	}

	fmt.Println(wg)

	go func() {
		for task_id := range in_progress_queue {
			var task = (*tasks)[task_id]
			fmt.Println("Running task:", task_id, "for", task.ws_name)
			mu.Lock()

			task.status = TASK_STATUS_SUCCESS
			(*tasks)[task_id] = task

			var next_tasks = find_unblocked_tasks(tasks)
			fmt.Println("Unblocked tasks:", next_tasks)
			wg.Add(len(next_tasks))

			for _, ntask_id := range next_tasks {
				var ntask = (*tasks)[ntask_id]
				ntask.status = TASK_STATUS_RUNNING
				(*tasks)[ntask_id] = ntask

				go func(id string) {
					in_progress_queue <- id
				}(ntask_id)
			}

			wg.Done()
			mu.Unlock()
		}
	}()

	wg.Wait()
}

func find_unblocked_tasks(tasks *map[string]Task) []string {
	var result = []string{}

	for task_id, task := range *tasks {
		if task.status != TASK_STATUS_PENDING {
			continue
		}

		var all_deps_finished = true
		for _, dep_id := range task.Deps {
			var dep = (*tasks)[dep_id]
			if dep.status != TASK_STATUS_SUCCESS && dep.status != TASK_STATUS_FAILURE {
				all_deps_finished = false
				break
			}
		}

		if all_deps_finished {
			result = append(result, task_id)
		}
	}

	return result
}
