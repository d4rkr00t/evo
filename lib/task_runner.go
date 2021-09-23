package lib

import (
	"errors"
	"fmt"
	"math"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/briandowns/spinner"
	"github.com/fatih/color"
)

func CreateTasksFromWorkspaces(
	targets []string,
	wm *WorkspacesMap,
	config *Config,
	lg *LoggerGroup,
) map[string]Task {
	var tasks = map[string]Task{}

	var __create_tasks func(cmd string, ws_name string)
	__create_tasks = func(cmd string, ws_name string) {
		var task_name = GetTaskName(cmd, ws_name)

		if _, ok := tasks[task_name]; ok {
			return
		}

		var ws = wm.workspaces[ws_name]
		var rule, has_rule = ws.GetRule(cmd)

		if !has_rule {
			return
		}

		var deps = []string{}

		for _, dep := range rule.Deps {
			if dep[0] == '@' {
				dep = dep[1:]
				for dep_name := range ws.Deps {
					if _, ok := wm.updated[dep_name]; ok {
						var _, has_rule = wm.workspaces[dep_name].GetRule(dep)
						if has_rule {
							deps = append(deps, dep_name+":"+dep)
							__create_tasks(dep, dep_name)
						}
					}
				}
			} else {
				deps = append(deps, ws_name+":"+dep)
				__create_tasks(dep, ws_name)
			}
		}

		tasks[task_name] = NewTask(ws_name, task_name, deps, rule.Outputs, func(ctx *Context, t *Task) (string, error) {
			var ws = wm.workspaces[ws.Name]

			for _, dep := range t.Deps {
				if tasks[dep].status == TASK_STATUS_FAILURE {
					var msg = fmt.Sprintf("cannot continue, dependency \"%s\" has failed", color.CyanString(tasks[dep].task_name))
					lg.Verbose().Badge(task_name).Error("error →", msg)
					return "", errors.New(msg)
				}
			}

			var run = func() (string, error) {
				var cmd = NewCmd(task_name, ws.Path, rule.Cmd, func(msg string) {
					lg.Verbose().Badge(task_name).Info("→ " + msg)
				}, func(msg string) {
					lg.Verbose().Badge(task_name).Error("→ ", msg)
				})
				return cmd.Run()
			}

			if !t.Invalidate(&ctx.cache, ws.hash) {
				lg.Verbose().Badge(task_name).Success("cache hit:", color.HiBlackString(ws.hash))
				var out = t.GetLogCache(&ctx.cache, ws.hash)
				if len(out) > 0 {
					lg.Verbose().Badge(task_name).Info("→ replaying output...")
					for _, line := range strings.Split(out, "\n") {
						lg.Verbose().Badge(task_name).Info("→ " + line)
					}
				}
				if t.HasOutputs() {
					ctx.cache.RestoreOutputs(t.GetCacheKey(ws.hash), ws.Path, rule.Outputs)
				}
				t.CacheState(&ctx.cache, ws.hash)
			} else {
				lg.Verbose().Badge(task_name).Info("running →", color.HiBlackString(rule.Cmd))
				var out, err = run()
				if err != nil {
					lg.Verbose().Badge(task_name).Error("error →", err.Error())
					return out, err
				}

				if t.HasOutputs() {
					var err = t.ValidateOutputs(ws.Path)
					if err != nil {
						return out, err
					}
					t.Cache(&ctx.cache, &ws, ws.hash)
				}
				t.CacheLog(&ctx.cache, ws.hash, out)
				t.CacheState(&ctx.cache, ws.hash)
			}

			return "", nil
		})
	}

	for ws := range wm.updated {
		for _, target := range targets {
			__create_tasks(target, ws)
		}
	}

	return tasks
}

type TaskResult struct {
	task_id string
	err     error
	out     string
}

func RunTasks(ctx *Context, tasks *map[string]Task, wm *WorkspacesMap, lg *LoggerGroup) error {
	var wg sync.WaitGroup
	var mu sync.RWMutex
	var mesure_mu sync.Mutex
	var num_goroutines = int(math.Min(float64(runtime.NumCPU())*0.5, float64(len(*tasks))))
	var queue_size = num_goroutines * 2
	var pqueue = make(chan string, queue_size)
	var dqueue = make(chan TaskResult)
	var in_progress int64
	var closed = false
	var task_errors = []TaskResult{}

	ctx.stats.StartMeasure("runtasks", MEASURE_KIND_STAGE)

	lg.Badge("threads").Info(fmt.Sprint(num_goroutines))
	lg.Badge("tasks").Info(" ", fmt.Sprint(len(*tasks)))
	lg.Verbose().Log("queue size ->", fmt.Sprint(queue_size), "| num cpus ->", fmt.Sprint(runtime.NumCPU()))
	lg.Log()

	wg.Add(num_goroutines)

	var progress_spinner = spinner.New(spinner.CharSets[14], 100*time.Millisecond)

	lg.Log("Running tasks...")
	lg.Log()

	if !lg.logger.verbose {
		progress_spinner.Start()
		progress_spinner.Prefix = color.HiBlackString("│ ")
		progress_spinner.Suffix = fmt.Sprintf(" done: %s / %s", "0", fmt.Sprint(len(*tasks)))
	}

	lg.Verbose().Log("Creating go routines...")

	for i := 0; i < num_goroutines; i++ {
		go func() {
			defer wg.Done()
			for task_id := range pqueue {
				mu.Lock()
				var task = (*tasks)[task_id]
				mu.Unlock()

				atomic.AddInt64(&in_progress, 1)

				mesure_mu.Lock()
				ctx.stats.StartMeasure(task_id, MEASURE_KIND_TASK)
				mesure_mu.Unlock()

				var out, err = task.Run(ctx, &task)

				mesure_mu.Lock()
				ctx.stats.StopMeasure(task_id)
				mesure_mu.Unlock()

				atomic.AddInt64(&in_progress, -1)

				dqueue <- TaskResult{task_id, err, out}
			}
		}()
	}

	lg.Verbose().Log("Starting done routine...")
	lg.Verbose().Log()
	go func() {
		for task_result := range dqueue {
			var task_id = task_result.task_id
			var err = task_result.err

			mu.Lock()
			var task = (*tasks)[task_id]
			if err == nil {
				mesure_mu.Lock()
				lg.Verbose().Badge(task_id).Success("done in " + color.HiBlackString(ctx.stats.GetMeasure(task_id).duration.String()))
				mesure_mu.Unlock()
				task.status = TASK_STATUS_SUCCESS
			} else {
				task.status = TASK_STATUS_FAILURE
				task_errors = append(task_errors, task_result)
			}
			(*tasks)[task_id] = task

			var next_tasks = find_unblocked_tasks(tasks)

			for _, ntask_id := range next_tasks {
				var ntask = (*tasks)[ntask_id]
				if ntask.status == TASK_STATUS_PENDING {
					ntask.status = TASK_STATUS_RUNNING
					(*tasks)[ntask_id] = ntask
					go func(tid string) {
						pqueue <- tid
						lg.Verbose().Badge(tid).Log("added to the queue")
					}(ntask_id)
				}
			}

			var all_done = true
			var done_count = 0
			for _, task := range *tasks {
				if task.status == TASK_STATUS_PENDING || task.status == TASK_STATUS_RUNNING {
					all_done = false
				} else {
					done_count += 1
				}
			}

			progress_spinner.Suffix = fmt.Sprintf("   done: %s / %s", fmt.Sprint(done_count), fmt.Sprint(len(*tasks)))

			if all_done && !closed {
				close(pqueue)
				closed = true
			}

			mu.Unlock()
		}
	}()

	var pushed = 0
	mu.Lock()
	for task_id, task := range *tasks {
		if pushed >= queue_size {
			break
		}

		if len(task.Deps) == 0 {
			task.status = TASK_STATUS_RUNNING
			(*tasks)[task_id] = task
			pushed += 1
			pqueue <- task_id
		}
	}
	mu.Unlock()
	wg.Wait()

	lg.Verbose().Log()
	lg.Verbose().Badge("start").Info("   Updating state of workspaces...")
	ctx.stats.StartMeasure("wsstate", MEASURE_KIND_STAGE)
	for ws_name := range wm.updated {
		var ws = wm.workspaces[ws_name]
		ws.CacheState(&ctx.cache, ws.hash)
	}
	lg.Verbose().Badge("done").Info("    in", ctx.stats.StopMeasure("wsstate").String())

	ctx.stats.StopMeasure("runtasks")
	progress_spinner.Stop()

	if !lg.logger.verbose {
		fmt.Println()
	}

	if len(task_errors) > 0 {
		lg.Log()
		lg.Log("Errors:")
		lg.Log()
		for _, task_result := range task_errors {
			lg.Badge(task_result.task_id).Error(task_result.err.Error(), task_result.out)
		}

		return errors.New("")
	}

	return nil
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
