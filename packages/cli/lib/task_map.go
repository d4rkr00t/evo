package lib

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"

	mapset "github.com/deckarep/golang-set"
	"github.com/fatih/color"
	"github.com/pyr-sh/dag"
)

type TasksMap struct {
	tasks  sync.Map
	graph  dag.AcyclicGraph
	length int32
}

func NewTaskMap(targets []string,
	wm *WorkspacesMap,
	config *Config,
	lg *LoggerGroup) TasksMap {
	var graph, tasks, length = CreateTasksFromWorkspaces(targets, wm, config, lg)
	return TasksMap{
		tasks, graph, length,
	}
}

func (tm *TasksMap) Validate() error {
	var cycles = tm.graph.Cycles()

	if len(cycles) == 0 {
		return nil
	}

	var error_msg = []string{}

	for _, cycle := range cycles {
		var path = ""
		for _, item := range cycle {
			if len(path) > 0 {
				path += " → "
			}
			path = path + fmt.Sprintf("%s", item)
		}
		path += fmt.Sprintf(" → %s", cycle[0])

		error_msg = append(error_msg, path)
	}

	return fmt.Errorf("cycle in the tasks graph\n%s", strings.Join(error_msg[:], "\n"))
}

func (tm *TasksMap) Walk(f func(tid string) error) {
	tm.graph.Walk(func(v dag.Vertex) error {
		var task_id = v.(string)
		return f(task_id)
	})
}

func (tm *TasksMap) Load(tid string) (Task, bool) {
	var value, ok = tm.tasks.Load(tid)
	if ok {
		return value.(Task), ok
	}
	var t Task
	return t, ok
}

func (tm *TasksMap) Store(task Task) {
	tm.tasks.Store(task.task_name, task)
}

func CreateTasksFromWorkspaces(targets []string,
	wm *WorkspacesMap,
	config *Config,
	lg *LoggerGroup) (dag.AcyclicGraph, sync.Map, int32) {
	var graph dag.AcyclicGraph
	var tasks sync.Map
	var visited = mapset.NewSet()
	var length int32 = 0

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

		tasks.Store(task_name, create_executable_task(ws_name, task_name, deps, rule, wm, &ws, &tasks, lg))
		atomic.AddInt32(&length, 1)
	}

	wm.updated.Each(func(ws_name interface{}) bool {
		for _, target := range targets {
			create_tasks(target, ws_name.(string))
		}
		return false
	})

	return graph, tasks, length
}

func create_executable_task(ws_name string, task_name string, deps []string, rule Rule, wm *WorkspacesMap, ws *Workspace, tasks *sync.Map, lg *LoggerGroup) Task {
	return NewTask(ws_name, task_name, rule.String(), deps, rule.Outputs, func(ctx *Context, t *Task) (string, error) {
		var ws, _ = wm.Load(ws.Name)
		var ws_partial_hash = ws.GetHashForTask()

		for _, dep := range t.Deps {
			var _sub_task, _ = tasks.Load(dep)
			var sub_task = _sub_task.(Task)
			if sub_task.status == TASK_STATUS_FAILURE {
				var msg = fmt.Sprintf("cannot continue, dependency \"%s\" has failed", color.CyanString(sub_task.task_name))
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

			var cmd_title = rule.Cmd
			if len(cmd_title) == 0 {
				cmd_title = "..."
			}
			lg.Badge(task_name).BadgeColor(t.color).Info("running → ", color.HiBlackString(cmd_title))

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
