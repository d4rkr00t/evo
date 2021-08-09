package lib

import (
	"fmt"
	"math"
	"os"
	"runtime"
	"scu/main/lib/cache"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/briandowns/spinner"
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
	os.Setenv("PATH", proj.GetNodeModulesBinPath()+":"+os.ExpandEnv("$PATH"))
	return Runner{cwd: cwd, project: proj, cache: cc}
}

func (r Runner) GetCwd() string {
	return r.cwd
}

func (r Runner) Run(cmd string) {
	fmt.Println("\n"+cmd+" ->", r.GetCwd())
	fmt.Println("\n===============")
	fmt.Println("")

	if r.project.Invalidate(&r.cache) {
		fmt.Println("INSTALL")
		r.project.InstallDeps(&r)
	}

	var updated = r.project.InvalidateWorkspaces(make([]string, 0), cmd, &r.cache)

	fmt.Println("\nUpdated:", len(updated), "of", len(r.project.Workspaces))
	fmt.Println("")

	if len(updated) > 0 {
		fmt.Println("Creating build tasks")
		var tasks = r.create_tasks(cmd, &updated)
		fmt.Println("Building...")
		if len(tasks) > 0 {
			spew.Dump(tasks)
			r.run_tasks(&tasks)
		}
	}

	r.project.CacheState(&r.cache)

	if len(updated) > 0 {
		fmt.Println("\n\n===============")
		fmt.Println("")
	}
}

func (r Runner) create_tasks(cmd string, workspaces *map[string]string) map[string]Task {
	var tasks = map[string]Task{}
	fmt.Println("Calculating affected packages...")
	var affected = r.project.GetAffected(workspaces)
	fmt.Println("Total affected ->", len(affected))
	fmt.Println("Creating tasks for affected packages...")

	var __create_tasks func(cmd string, ws_name string)
	__create_tasks = func(cmd string, ws_name string) {
		var task_name = ws_name + ":" + cmd

		if _, ok := tasks[task_name]; ok {
			return
		}

		var ws = r.project.GetWs(ws_name)
		var rule = r.project.GetRule(cmd, ws.Path)
		var deps = []string{}

		spew.Dump(ws_name, rule)

		for _, dep := range rule.Deps {
			if dep[0] == '@' {
				dep = dep[1:]
				for dep_name := range ws.Deps {
					if _, ok := affected[dep_name]; ok {
						deps = append(deps, dep_name+":"+dep)
						__create_tasks(dep, dep_name)
					}
				}
			} else {
				deps = append(deps, ws_name+":"+dep)
				__create_tasks(dep, ws_name)
			}
		}

		tasks[task_name] = NewTask(ws_name, task_name, deps, func(r *Runner) {
			var ws_hash = affected[ws.Name]
			// fmt.Println(task_name, "-> compiling")

			var run = func() {
				var args = strings.Split(rule.Cmd, " ")
				var cmd_name = args[0]
				var cmd_args = args[1:]

				var cmd = NewCmd(task_name, ws.Path, cmd_name, cmd_args)
				cmd.Run()
			}

			var cache_key = cmd + ":" + ws_hash

			if r.cache.Has(cache_key) {
				// fmt.Println(task_name, "-> cache hit:", w.Name, ws_hash)
				if rule.CacheOutput {
					r.cache.RestoreDir(cache_key, ws.Path)
				}
			} else {
				run()
				if rule.CacheOutput {
					ws.Cache(&r.cache, cache_key)
				}
			}

			ws.CacheState(&r.cache, cmd, ws_hash)
		}, false)

		// spew.Dump(ws_name, rule, deps)
	}

	for ws := range affected {
		__create_tasks(cmd, ws)
	}

	return tasks
}

func (r Runner) run_tasks(tasks *map[string]Task) {
	var wg sync.WaitGroup
	var mu sync.RWMutex
	var num_goroutines = int(math.Min(float64(runtime.NumCPU())*0.8, float64(len(*tasks))))
	var queue_size = num_goroutines * 2
	var pqueue = make(chan string, queue_size)
	var dqueue = make(chan string)
	var s = spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	var count_done = 0
	var in_progress int64
	var total = len(*tasks)

	fmt.Println("Num goroutins ->", num_goroutines, "| queue size ->", queue_size, "| num cpus ->", runtime.NumCPU())

	wg.Add(num_goroutines)

	fmt.Println("Creating go routines...")

	for i := 0; i < num_goroutines; i++ {
		go func() {
			defer wg.Done()
			for task_id := range pqueue {
				mu.RLock()
				var task = (*tasks)[task_id]
				mu.RUnlock()

				atomic.AddInt64(&in_progress, 1)
				s.Suffix = build_spinner_text(total, count_done, len(pqueue), int(in_progress))

				// var start = time.Now()
				task.Run(&r)

				atomic.AddInt64(&in_progress, -1)
				// var duration = time.Since(start)
				// fmt.Println(task_id, "-> duration:", duration)

				dqueue <- task_id
			}
		}()
	}

	fmt.Println("Starting done routine...")
	go func() {
		for task_id := range dqueue {
			count_done += 1
			fmt.Println("Finished task ->", task_id)

			mu.Lock()
			var task = (*tasks)[task_id]
			task.status = TASK_STATUS_SUCCESS
			(*tasks)[task_id] = task

			var next_tasks = find_unblocked_tasks(tasks)

			for _, ntask_id := range next_tasks {
				var ntask = (*tasks)[ntask_id]
				ntask.status = TASK_STATUS_RUNNING
				(*tasks)[ntask_id] = ntask
				go func(tid string) {
					pqueue <- tid
					fmt.Println("Adding task ->", tid)
				}(ntask_id)
			}
			mu.Unlock()

			if count_done == len(*tasks) {
				close(pqueue)
			}
		}
	}()

	fmt.Println("Adding initial tasks...")
	for task_id, task := range *tasks {
		if len(task.Deps) == 0 {
			task.status = TASK_STATUS_RUNNING
			(*tasks)[task_id] = task
			pqueue <- task_id
		}
	}

	s.Suffix = build_spinner_text(total, count_done, len(pqueue), int(in_progress))

	s.Start()
	wg.Wait()
	s.Stop()
}

func build_spinner_text(total int, done int, queued int, in_progress int) string {
	return " [" + "total:" + fmt.Sprint(total) + " | waiting:" + fmt.Sprint(total-done-in_progress) + " | done:" + fmt.Sprint(done) + " | queued:" + fmt.Sprint(queued) + " | running:" + fmt.Sprint(in_progress) + "] "
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
