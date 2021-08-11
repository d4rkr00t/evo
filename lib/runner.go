package lib

import (
	"fmt"
	"math"
	"os"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/fatih/color"
)

func Run(ctx Context) {
	ctx.stats.StartMeasure("total", MEASURE_KIND_STAGE)

	os.Setenv("PATH", GetNodeModulesBinPath(ctx.root)+":"+os.ExpandEnv("$PATH"))

	ctx.logger.Log()
	ctx.logger.LogWithBadge("cwd", ctx.cwd)
	ctx.logger.LogWithBadge("scu", fmt.Sprintf("Target → %s", color.CyanString(ctx.target)))

	if ctx.root_pkg_json.Invalidate(&ctx.cache) || !IsNodeModulesExist(ctx.root) {
		ctx.stats.StartMeasure("install", MEASURE_KIND_STAGE)
		var install_lg = ctx.logger.CreateGroup()
		install_lg.Start("Installing dependencies...")

		InstallNodeDeps(ctx.root, &install_lg)
		ctx.root_pkg_json.CacheState(&ctx.cache)

		install_lg.End(ctx.stats.StopMeasure("install"))
	}

	ctx.stats.StartMeasure("invalidate", MEASURE_KIND_STAGE)
	var invalidate_lg = ctx.logger.CreateGroup()
	invalidate_lg.Start("Invalidating workspaces...")

	var workspaces = GetWorkspaces(ctx.root, &ctx.config)
	var dep_graph = NewDepGraph(&workspaces)
	var updated_ws = InvalidateWorkspaces(&workspaces, ctx.target, &ctx.cache)

	if len(updated_ws) > 0 {
		invalidate_lg.LogWithBadge(
			"updated",
			color.CyanString(fmt.Sprint((len(updated_ws)))),
			"of",
			color.CyanString(fmt.Sprint((len(workspaces)))),
			"workspaces",
		)

		invalidate_lg.LogVerbose("Calculating affected workspaces...")
		var affected_ws = dep_graph.GetAffected(&workspaces, &updated_ws)
		invalidate_lg.LogWithBadge(
			"affected",
			color.CyanString(fmt.Sprint((len(affected_ws)))),
			"of",
			color.CyanString(fmt.Sprint((len(workspaces)))),
			"workspaces",
		)
		invalidate_lg.End(ctx.stats.StopMeasure("invalidate"))

		ctx.stats.StartMeasure("linking", MEASURE_KIND_STAGE)
		var linking_lg = ctx.logger.CreateGroup()
		linking_lg.Start("Linking workspaces...")

		LinkWorkspaces(ctx.root, &workspaces, &updated_ws)

		linking_lg.End(ctx.stats.StopMeasure("linking"))

		ctx.stats.StartMeasure("run", MEASURE_KIND_STAGE)
		var run_lg = ctx.logger.CreateGroup()

		run_lg.Start(fmt.Sprintf("Running target → %s", color.CyanString(ctx.target)))
		run_lg.LogVerbose("Creating tasks...")

		var tasks = create_tasks(
			ctx.target,
			&workspaces,
			&updated_ws,
			&affected_ws,
			&ctx.config,
			&run_lg,
		)

		if len(tasks) > 0 {
			run_lg.LogVerbose("Executing tasks...")
			// spew.Dump(tasks)
			run_tasks(&ctx, &tasks, &run_lg)
		}

		run_lg.End(ctx.stats.StopMeasure("run"))
	} else {
		invalidate_lg.Log("Everything is up-to-date.")
		invalidate_lg.End(ctx.stats.StopMeasure("invalidate"))
	}

	ctx.logger.Log()
	ctx.logger.LogWithBadge("Total time", color.GreenString(ctx.stats.StopMeasure("total").String()))
}

func create_tasks(
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
		var task_name = ws_name + ":" + cmd

		if _, ok := tasks[task_name]; ok {
			return
		}

		var ws = (*workspaces)[ws_name]
		var rule = config.GetRule(cmd, ws.Path)
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

		tasks[task_name] = NewTask(ws_name, task_name, deps, func(ctx *Context) {
			var ws_hash = (*affected_ws)[ws.Name]
			lg.InfoWithBadge(task_name, "starting...")

			var run = func() {
				var args = strings.Split(rule.Cmd, " ")
				var cmd_name = args[0]
				var cmd_args = args[1:]

				var cmd = NewCmd(task_name, ws.Path, cmd_name, cmd_args, func(msg string) {
					lg.InfoWithBadge(task_name, msg)
				})
				cmd.Run()
			}

			var cache_key = cmd + ":" + ws_hash

			if ctx.cache.Has(cache_key) {
				lg.SuccessWithBadge(task_name, "cache hit:", ws_hash)
				if rule.CacheOutput {
					ctx.cache.RestoreDir(cache_key, ws.Path)
				}
			} else {
				run()
				if rule.CacheOutput {
					ws.Cache(&ctx.cache, cache_key)
				}
			}

			ws.CacheState(&ctx.cache, cmd, ws_hash)
		}, false)

		// spew.Dump(ws_name, rule, deps)
	}

	for ws := range *affected_ws {
		__create_tasks(cmd, ws)
	}

	return tasks
}

func run_tasks(ctx *Context, tasks *map[string]Task, lg *LoggerGroup) {
	var wg sync.WaitGroup
	var mu sync.RWMutex
	var num_goroutines = int(math.Min(float64(runtime.NumCPU())*0.8, float64(len(*tasks))))
	var queue_size = num_goroutines * 2
	var pqueue = make(chan string, queue_size)
	var dqueue = make(chan string)
	var count_done = 0
	var in_progress int64

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
				ctx.stats.StartMeasure(task_id, MEASURE_KIND_TASK)
				task.Run(ctx)
				ctx.stats.StopMeasure(task_id)
				atomic.AddInt64(&in_progress, -1)

				dqueue <- task_id
			}
		}()
	}

	lg.LogVerbose("Starting done routine...")
	go func() {
		for task_id := range dqueue {
			count_done += 1
			lg.SuccessWithBadge(task_id, "done in "+ctx.stats.GetMeasure(task_id).duration.String())

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
