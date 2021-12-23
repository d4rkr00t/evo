package lib

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	mapset "github.com/deckarep/golang-set"
	"github.com/fatih/color"
	"github.com/pyr-sh/dag"
	"golang.org/x/sync/semaphore"
)

func CreateTasksFromWorkspaces(targets []string,
	wm *WorkspacesMap,
	config *Config,
	lg *LoggerGroup) (dag.AcyclicGraph, map[string]Task) {
	var graph dag.AcyclicGraph
	var tasks = make(map[string]Task)
	var visited = mapset.NewSet()

	var create_tasks func(target string, ws_name string)
	create_tasks = func(target string, ws_name string) {
		var task_name = GetTaskName(target, ws_name)
		var ws, _ = wm.Load(ws_name)
		var rule, has_rule = ws.GetRule(target)

		if visited.Contains(task_name) {
			return
		}

		visited.Add(task_name)

		if !has_rule {
			return
		}

		graph.Add(task_name)
		var deps []string

		for _, dep := range rule.Deps {
			if dep[0] == '@' {
				dep = dep[1:]
				for dep_name := range ws.Deps {
					if wm.updated.Contains(dep_name) {
						var dep_ws, _ = wm.Load(dep_name)
						var _, has_rule = dep_ws.GetRule(dep)
						if has_rule {
							create_tasks(dep, dep_name)
							var dep_task_name = GetTaskName(dep, dep_name)
							deps = append(deps, dep_task_name)
							graph.Connect(dag.BasicEdge(task_name, dep_task_name))
						}
					}
				}
			} else {
				create_tasks(dep, ws_name)
				var dep_task_name = GetTaskName(dep, ws_name)
				deps = append(deps, dep_task_name)
				graph.Connect(dag.BasicEdge(task_name, dep_task_name))
			}
		}

		tasks[task_name] = create_executable_task(ws_name, task_name, deps, rule, wm, &ws, &tasks, lg)
	}

	wm.updated.Each(func(ws_name interface{}) bool {
		for _, target := range targets {
			create_tasks(target, ws_name.(string))
		}
		return false
	})

	return graph, tasks
}

func create_executable_task(ws_name string, task_name string, deps []string, rule Rule, wm *WorkspacesMap, ws *Workspace, tasks *map[string]Task, lg *LoggerGroup) Task {
	return NewTask(ws_name, task_name, rule.String(), deps, rule.Outputs, func(ctx *Context, t *Task) (string, error) {
		var ws, _ = wm.Load(ws.Name)
		var ws_partial_hash = ws.GetHashForTask()

		for _, dep := range t.Deps {
			if (*tasks)[dep].status == TASK_STATUS_FAILURE {
				var msg = fmt.Sprintf("cannot continue, dependency \"%s\" has failed", color.CyanString((*tasks)[dep].task_name))
				lg.Badge(task_name).BadgeColor(t.color).Error("error → ", msg)
				return "", errors.New(msg)
			}
		}

		var run = func() (string, error) {
			var cmd = NewCmd(task_name, ws.Path, rule.Cmd, func(msg string) {
				lg.Badge(task_name).BadgeColor(t.color).Info(color.HiBlackString("← ") + msg)
			}, func(msg string) {
				lg.Badge(task_name).BadgeColor(t.color).Error(color.RedString("← "), msg)
			})
			return cmd.Run()
		}

		if !t.Invalidate(&ctx.cache, tasks, ws_partial_hash) {
			lg.Badge(task_name).BadgeColor(t.color).Success(color.GreenString("cache hit:"), color.HiBlackString(ws.hash))
			var out = t.GetLogCache(&ctx.cache, tasks, ws_partial_hash)
			if len(out) > 0 {
				lg.Badge(task_name).BadgeColor(t.color).Info("replaying output...")
				for _, line := range strings.Split(out, "\n") {
					lg.Badge(task_name).BadgeColor(t.color).Info(color.HiBlackString("← "), line)
				}
			}
			if t.HasOutputs() {
				t.CleanOutputs(ws.Path)
				lg.Badge(task_name).BadgeColor(t.color).Info("restoring outputs from cache...")
				ctx.cache.RestoreOutputs(t.GetCacheKey(tasks, ws_partial_hash), ws.Path, rule.Outputs)
				t.UpdateOutputsHash(ws.Path)
			}
			t.CacheState(&ctx.cache, ws.hash)
		} else {
			lg.Badge(task_name).BadgeColor(t.color).Info(color.YellowString("cache miss:"), color.HiBlackString(ws.hash))
			lg.Badge(task_name).BadgeColor(t.color).Verbose().Info("cleaning outputs...")
			t.CleanOutputs(ws.Path)
			lg.Badge(task_name).BadgeColor(t.color).Info("running → ", color.HiBlackString(rule.Cmd))
			var out, err = run()
			if err != nil {
				lg.Badge(task_name).BadgeColor(t.color).Error(color.RedString("error → "), err.Error())
				return out, err
			}

			if t.HasOutputs() {
				var err = t.ValidateOutputs(ws.Path)
				if err != nil {
					return out, err
				}
				t.UpdateOutputsHash(ws.Path)
				t.Cache(&ctx.cache, &ws, tasks, ws_partial_hash)
			}
			t.CacheLog(&ctx.cache, tasks, ws_partial_hash, out)
			t.CacheState(&ctx.cache, ws.hash)
		}

		return "", nil
	})
}

type TaskResult struct {
	task_id string
	err     error
	out     string
}

func RunTasks(ctx *Context, tasks_graph *dag.AcyclicGraph, tasks *map[string]Task, wm *WorkspacesMap, lg *LoggerGroup) error {
	var mesure_mu sync.Mutex
	var mu sync.RWMutex
	var task_errors = []TaskResult{}
	var sem = semaphore.NewWeighted(int64(ctx.concurrency))
	var cc = context.TODO()

	ctx.stats.StartMeasure("runtasks", MEASURE_KIND_STAGE)
	lg.Badge("tasks").Info(fmt.Sprint(len(*tasks)))
	lg.Log()
	lg.Log("Running tasks...")
	lg.Log()

	tasks_graph.Walk(func(vx dag.Vertex) error {
		var task_id = fmt.Sprint(vx)

		mu.Lock()
		var task = (*tasks)[task_id]
		task.UpdateStatus(TASK_STATUS_RUNNING)
		(*tasks)[task_id] = task
		mu.Unlock()

		if err := sem.Acquire(cc, 1); err != nil {
			panic(fmt.Sprintf("Failed to acquire semaphore: %v", err))
		}
		defer sem.Release(1)

		mesure_mu.Lock()
		ctx.stats.StartMeasure(task_id, MEASURE_KIND_TASK)
		mesure_mu.Unlock()

		var trace = ctx.tracing.Event(task_id)
		var out, err = task.Run(ctx, &task)
		trace.Done()

		mesure_mu.Lock()
		ctx.stats.StopMeasure(task_id)
		mesure_mu.Unlock()

		mu.Lock()
		if err == nil {
			mesure_mu.Lock()
			lg.Badge(task_id).BadgeColor(task_id).Success("done in " + color.HiBlackString(ctx.stats.GetMeasure(task_id).duration.String()))
			mesure_mu.Unlock()
			task.UpdateStatus(TASK_STATUS_SUCCESS)
		} else {
			task.UpdateStatus(TASK_STATUS_FAILURE)
			task_errors = append(task_errors, TaskResult{task_id, err, out})
		}

		(*tasks)[task_id] = task
		mu.Unlock()

		return nil
	})

	lg.Verbose().Log()
	lg.Verbose().Badge("start").Info("   Updating states of workspaces...")
	ctx.stats.StartMeasure("wsstate", MEASURE_KIND_STAGE)
	var trace = ctx.tracing.Event("updating states of workspaces")

	wm.updated.Each(func(ws_name interface{}) bool {
		var ws, _ = wm.Load(ws_name.(string))
		ws.CacheState(&ctx.cache, ws.hash)
		return false
	})

	trace.Done()
	lg.Verbose().Badge("done").Info("    in", ctx.stats.StopMeasure("wsstate").String())
	ctx.stats.StopMeasure("runtasks")

	if !lg.logger.verbose {
		lg.Log()
	}

	if len(task_errors) > 0 {
		lg.Log()
		lg.Log("Errors:")
		lg.Log()
		for _, task_result := range task_errors {
			lg.Badge(task_result.task_id).BadgeColor((*tasks)[task_result.task_id].color).Error(task_result.err.Error(), task_result.out)
		}

		return errors.New("")
	}

	return nil
}
