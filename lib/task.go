package lib

import (
	"evo/main/lib/cache"
	"fmt"
	"strings"
)

type Task struct {
	ws_name   string
	task_name string
	status    int
	Deps      []string
	Run       task_run
	Outputs   []string
}

type task_run = func(ctx *Context, t *Task) (string, error)

const (
	TASK_STATUS_PENDING = iota
	TASK_STATUS_RUNNING = iota
	TASK_STATUS_SUCCESS = iota
	TASK_STATUS_FAILURE = iota
)

func NewTask(ws_name string, task_name string, deps []string, outputs []string, run task_run) Task {
	return Task{
		ws_name:   ws_name,
		task_name: task_name,
		status:    TASK_STATUS_PENDING,
		Deps:      deps,
		Run:       run,
		Outputs:   outputs,
	}
}

func (t Task) GetCacheKey(ws_hash string) string {
	return ClearTaskName(t.task_name) + ":" + ws_hash
}

func (t Task) Invalidate(cc *cache.Cache, ws_hash string) bool {
	return !cc.Has(t.GetCacheKey(ws_hash))
}

func (t Task) Cache(cc *cache.Cache, ws *Workspace, ws_hash string) {
	if len(t.Outputs) > 0 {
		var ignores = cache.CacheDirIgnores{
			"node_modules": true,
		}
		cc.CacheOutputs(t.GetCacheKey(ws_hash), ws.Path, t.Outputs, ignores)
	} else {
		cc.CacheData(t.GetCacheKey(ws_hash), "")
	}
}

func (t Task) GetStateKey() string {
	return ClearTaskName(t.task_name)
}

func (t Task) CacheState(c *cache.Cache, ws_hash string) {
	c.CacheData(t.GetStateKey(), ws_hash)
}

func (t Task) CacheLog(c *cache.Cache, ws_hash string, log string) {
	c.CacheData(t.GetCacheKey(ws_hash)+":log", log)
}

func (t Task) GetCacheState(c *cache.Cache) string {
	return c.ReadData(t.GetStateKey())
}

func (t Task) GetLogCache(c *cache.Cache, ws_hash string) string {
	return c.ReadData(t.GetCacheKey(ws_hash) + ":log")
}

func (t Task) HasOutputs() bool {
	return len(t.Outputs) > 0
}

func ClearTaskName(name string) string {
	return strings.Replace(name, "/", "__", -1)
}

func GetTaskName(target string, ws_name string) string {
	return ws_name + ":" + target
}

func (t Task) String() string {
	return fmt.Sprintf("%s:%s", t.ws_name, t.task_name)
}
