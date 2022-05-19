package task_graph

import (
	gocontext "context"
	"evo/internal/errors"
	"fmt"
	"strings"
	"sync"

	"github.com/pyr-sh/dag"
	"golang.org/x/sync/semaphore"
)

type TasksMap = sync.Map

type TaskGraph struct {
	graph    dag.AcyclicGraph
	tasksMap TasksMap
}

func New() TaskGraph {
	return TaskGraph{
		graph:    dag.AcyclicGraph{},
		tasksMap: sync.Map{},
	}
}

func (tg *TaskGraph) Add(task *Task) {
	tg.graph.Add(task.Name())
	tg.Store(task)
}

func (tg *TaskGraph) Connect(from string, to string) {
	tg.graph.Connect(dag.BasicEdge(from, to))
}

func (tg *TaskGraph) Load(taskName string) (*Task, bool) {
	var value, ok = tg.tasksMap.Load(taskName)
	var task Task

	if ok {
		task = value.(Task)
	}

	return &task, ok
}

func (tg *TaskGraph) Store(task *Task) {
	tg.tasksMap.Store(task.Name(), *task)
}

func (tg *TaskGraph) Walk(fn func(task *Task) error, concurency int) {
	var cc = gocontext.TODO()
	var sem = semaphore.NewWeighted(int64(concurency))

	tg.graph.Walk(func(v dag.Vertex) error {
		var taskName = fmt.Sprint(v)
		if err := sem.Acquire(cc, 1); err != nil {
			panic(fmt.Sprintf("Failed to acquire semaphore: %v", err))
		}
		defer sem.Release(1)
		var task, _ = tg.Load(taskName)
		return fn(task)
	})
}

func (tg *TaskGraph) Validate() error {
	var cycles = tg.graph.Cycles()

	if len(cycles) == 0 {
		return nil
	}

	var msg = []string{"cycles in the task graph:"}
	for _, cycle := range cycles {
		msg = append(msg, fmt.Sprintf("– %s", cycle))
	}

	return errors.New(errors.ErrorTaskGraphCycle, strings.Join(msg, "\n"))
}
