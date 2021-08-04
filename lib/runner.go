package lib

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"scu/main/lib/cache"
	"strings"
	"sync"
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

	fmt.Println("Creating build tasks")
	var tasks = r.create_tasks(&updated)
	fmt.Println("Building...")
	if len(tasks) > 0 {
		// spew.Dump(tasks)
		r.run_tasks(&tasks)
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

	fmt.Println("Creating tasks for affected packages...")
	for ws_name := range affected {
		var task = r.project.GetWs(ws_name).CreateBuildTask(&affected, workspaces)
		tasks[task.task_name] = task
	}

	return tasks
}

func (r Runner) run_tasks(tasks *map[string]Task) {
	var in_progress_queue = make(chan string)
	var wg sync.WaitGroup
	var mu sync.Mutex
	var guard = make(chan struct{}, runtime.NumCPU())

	var s = spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Start()

	for task_id, task := range *tasks {
		if len(task.Deps) == 0 {
			wg.Add(1)
			guard <- struct{}{}
			task.status = TASK_STATUS_RUNNING
			(*tasks)[task_id] = task
			go func(id string) {
				in_progress_queue <- id
			}(task_id)
		}
	}

	s.Suffix = build_spinner_text(tasks)

	go func() {
		for task_id := range in_progress_queue {
			go func(task_id string) {
				var task = (*tasks)[task_id]
				// fmt.Println("\n"+task_id, "-> running")
				// var start = time.Now()
				task.Run(&r)
				// var duration = time.Since(start)
				// fmt.Println(task_id, "-> duration:", duration)

				mu.Lock()

				task.status = TASK_STATUS_SUCCESS
				(*tasks)[task_id] = task

				var next_tasks = find_unblocked_tasks(tasks)
				// fmt.Println(task_id, "-> unblocked tasks:", next_tasks)
				wg.Add(len(next_tasks))
				<-guard

				for _, ntask_id := range next_tasks {
					var ntask = (*tasks)[ntask_id]
					guard <- struct{}{}
					if ntask.status != TASK_STATUS_PENDING {
						<-guard
						continue
					}
					ntask.status = TASK_STATUS_RUNNING
					(*tasks)[ntask_id] = ntask

					go func(id string) {
						in_progress_queue <- id
					}(ntask_id)
				}

				s.Suffix = build_spinner_text(tasks)

				wg.Done()
				mu.Unlock()
			}(task_id)
		}
	}()

	wg.Wait()
	s.Stop()
}

func build_spinner_text(tasks *map[string]Task) string {
	var spinner_text = []string{}
	var count_done = 0
	var count_running = 0
	for _, t := range *tasks {
		if t.status == TASK_STATUS_RUNNING {
			spinner_text = append(spinner_text, t.task_name)
			count_running += 1
		} else if t.status == TASK_STATUS_SUCCESS {
			count_done += 1
		}
	}
	var postfix = ""
	if count_running > 3 {
		postfix = "..."
		spinner_text = spinner_text[:3]
	}

	return " [waiting:" + fmt.Sprint((len(*tasks) - (count_done + count_running))) + " | running:" + fmt.Sprint(count_running) + "] " + strings.Join(spinner_text, " | ") + postfix
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
