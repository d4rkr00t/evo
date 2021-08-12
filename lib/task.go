package lib

import (
	"scu/main/lib/cache"
	"strings"
)

type Task struct {
	ws_name   string
	task_name string
	status    int
	Deps      []string
	Run       task_run
	Force     bool
}

type task_run = func(ctx *Context, t *Task)

const (
	TASK_STATUS_PENDING = iota
	TASK_STATUS_RUNNING = iota
	TASK_STATUS_SUCCESS = iota
	TASK_STATUS_FAILURE = iota
)

func NewTask(ws_name string, task_name string, deps []string, run task_run, force bool) Task {
	return Task{
		ws_name:   ws_name,
		task_name: task_name,
		status:    TASK_STATUS_PENDING,
		Deps:      deps,
		Run:       run,
		Force:     force,
	}
}

func (t Task) GetCacheKey(ws_hash string) string {
	return ClearTaskName(t.task_name) + ":" + ws_hash
}

func (t Task) Invalidate(cc *cache.Cache, ws_hash string) bool {
	return t.GetCacheState(cc) != ws_hash
}

func (t Task) GetStateKey() string {
	return ClearTaskName(t.task_name)
}

func (t Task) CacheState(c *cache.Cache, ws_hash string) {
	c.CacheData(t.GetStateKey(), ws_hash)
}

func (t Task) GetCacheState(c *cache.Cache) string {
	return c.ReadData(t.GetStateKey())
}

func ClearTaskName(name string) string {
	return strings.Replace(name, "/", "__", -1)
}

func GetTaskName(target string, ws_name string) string {
	return ws_name + ":" + target
}
