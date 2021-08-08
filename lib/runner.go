package lib

import (
	"fmt"
	"math"
	"os"
	"os/exec"
	"runtime"
	"scu/main/lib/cache"
	"sync"
	"sync/atomic"
	"time"

	"github.com/briandowns/spinner"
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

func (r Runner) Build() {
	fmt.Println("\nBuild:", r.GetCwd())
	fmt.Println("\n===============")
	fmt.Println("")

	var updated = r.project.Invalidate(make([]string, 0), r.cache)

	fmt.Println("\nUpdated:", len(updated), "of", len(r.project.Workspaces))
	fmt.Println("")

	if len(updated) > 0 {
		fmt.Println("Creating build tasks")
		var tasks = r.create_tasks(&updated)
		fmt.Println("Building...")
		if len(tasks) > 0 {
			// spew.Dump(tasks)
			r.run_tasks(&tasks)
		}
	}

	if len(updated) > 0 {
		fmt.Println("\n\n===============")
		fmt.Println("")
	}
}

func (r Runner) CreateExec(dir string, name string, params []string) exec.Cmd {
	var cmd = exec.Command(name, params...)
	cmd.Dir = dir
	return *cmd
}

func (r Runner) create_tasks(workspaces *map[string]string) map[string]Task {
	var tasks = map[string]Task{}
	fmt.Println("Calculating affected packages...")
	var affected = r.project.GetAffected(workspaces)
	fmt.Println("Total affected ->", len(affected))

	fmt.Println("Creating tasks for affected packages...")
	for ws_name := range affected {
		var task = r.project.GetWs(ws_name).CreateBuildTask(&affected, workspaces)
		tasks[task.task_name] = task
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
