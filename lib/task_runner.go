package lib

import (
	"errors"
	"fmt"
	"math"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/fatih/color"
)

func CreateTasksFromWorkspaces(
	cmd string,
	workspaces *WorkspacesMap,
	updated_ws *map[string]string,
	affected_ws *map[string]string,
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

		var ws = (*workspaces)[ws_name]
		var rule, has_rule = ws.GetRule(cmd)

		if !has_rule {
			return
		}

		var deps = []string{}

		for _, dep := range rule.Deps {
			if dep[0] == '@' {
				dep = dep[1:]
				for dep_name := range ws.Deps {
					if _, ok := (*affected_ws)[dep_name]; ok {
						deps = append(deps, dep_name+":"+dep)
						__create_tasks(dep, dep_name)
					}
				}
			} else {
				deps = append(deps, ws_name+":"+dep)
				__create_tasks(dep, ws_name)
			}
		}

		tasks[task_name] = NewTask(ws_name, task_name, deps, func(ctx *Context, t *Task) error {
			var ws = (*workspaces)[ws.Name]
			var ws_hash = ws.Hash(workspaces)

			for _, dep := range t.Deps {
				if tasks[dep].status == TASK_STATUS_FAILURE {
					var msg = fmt.Sprintf("cannot continue, dependency \"%s\" has failed", color.CyanString(tasks[dep].task_name))
					lg.ErrorWithBadge(task_name, "error →", msg)
					return errors.New(msg)
				}
			}

			var run = func() (string, error) {
				var cmd = NewCmd(task_name, ws.Path, rule.Cmd, func(msg string) {
					lg.InfoWithBadge(task_name, "→ "+msg)
				})
				return cmd.Run()
			}

			if !t.Invalidate(&ctx.cache, ws_hash) {
				lg.SuccessWithBadge(task_name, "cache hit:", color.HiBlackString(ws_hash))
				var out = t.GetLogCache(&ctx.cache, ws_hash)
				if len(out) > 0 {
					lg.InfoWithBadge(task_name, "→ replaying output...")
					for _, line := range strings.Split(out, "\n") {
						lg.InfoWithBadge(task_name, "→ "+line)
					}
				}
				if t.CacheOutput {
					ctx.cache.RestoreDir(t.GetCacheKey(ws_hash), ws.Path)
				}
			} else {
				lg.InfoWithBadge(task_name, "running →", color.HiBlackString(rule.Cmd))
				var out, err = run()
				if err != nil {
					lg.ErrorWithBadge(task_name, "error →", err.Error())
					return err
				}

				if t.CacheOutput {
					ws_hash = ws.Hash(workspaces)
				}

				t.CacheLog(&ctx.cache, ws_hash, out)
				t.Cache(&ctx.cache, &ws, ws_hash)
			}

			ws.CacheState(&ctx.cache, ws_hash)
			t.CacheState(&ctx.cache, ws_hash)

			return nil
		}, rule.CacheOutput)
	}

	for ws := range *affected_ws {
		__create_tasks(cmd, ws)
	}

	return tasks
}

type TaskResult struct {
	task_id string
	err     error
}

func RunTasks(ctx *Context, tasks *map[string]Task, lg *LoggerGroup) {
	var wg sync.WaitGroup
	var mu sync.RWMutex
	var num_goroutines = int(math.Min(float64(runtime.NumCPU())*0.8, float64(len(*tasks))))
	var queue_size = num_goroutines * 2
	var pqueue = make(chan string, queue_size)
	var dqueue = make(chan TaskResult)
	var count_done = 0
	var in_progress int64

	ctx.stats.StartMeasure("runtasks", MEASURE_KIND_STAGE)

	lg.LogWithBadge("threads", fmt.Sprint(num_goroutines))
	lg.LogWithBadge("tasks", fmt.Sprint(len(*tasks)))
	lg.LogVerbose("queue size ->", fmt.Sprint(queue_size), "| num cpus ->", fmt.Sprint(runtime.NumCPU()))
	lg.Log()

	wg.Add(num_goroutines)

	lg.Log("Running tasks...")
	lg.Log()

	lg.LogVerbose("Creating go routines...")

	for i := 0; i < num_goroutines; i++ {
		go func() {
			defer wg.Done()
			for task_id := range pqueue {
				mu.RLock()
				var task = (*tasks)[task_id]
				mu.RUnlock()

				atomic.AddInt64(&in_progress, 1)
				mu.Lock()
				ctx.stats.StartMeasure(task_id, MEASURE_KIND_TASK)
				mu.Unlock()

				var err = task.Run(ctx, &task)

				// TODO: Fix this lock later
				mu.Lock()
				ctx.stats.StopMeasure(task_id)
				mu.Unlock()
				atomic.AddInt64(&in_progress, -1)

				dqueue <- TaskResult{task_id, err}
			}
		}()
	}

	lg.LogVerbose("Starting done routine...")
	lg.LogVerbose()
	go func() {
		for task_result := range dqueue {
			count_done += 1

			var task_id = task_result.task_id
			var err = task_result.err

			mu.Lock()
			var task = (*tasks)[task_id]
			if err == nil {
				lg.SuccessWithBadge(task_id, "done in "+color.HiBlackString(ctx.stats.GetMeasure(task_id).duration.String()))
				task.status = TASK_STATUS_SUCCESS
			} else {
				task.status = TASK_STATUS_FAILURE
			}
			(*tasks)[task_id] = task

			var next_tasks = find_unblocked_tasks(tasks)

			for _, ntask_id := range next_tasks {
				var ntask = (*tasks)[ntask_id]
				ntask.status = TASK_STATUS_RUNNING
				(*tasks)[ntask_id] = ntask
				go func(tid string) {
					pqueue <- tid
					lg.LogWithBadgeVerbose(tid, "added to the queue")
				}(ntask_id)
			}
			mu.Unlock()

			if count_done == len(*tasks) {
				close(pqueue)
			}
		}
	}()

	for task_id, task := range *tasks {
		if len(task.Deps) == 0 {
			task.status = TASK_STATUS_RUNNING
			(*tasks)[task_id] = task
			pqueue <- task_id
		}
	}

	wg.Wait()
	ctx.stats.StopMeasure("runtasks")
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
