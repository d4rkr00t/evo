package lib

type Task struct {
	ws_name   string
	task_name string
	status    int
	Deps      []string
	Run       task_run
	Force     bool
}

type task_run = func(r *Runner)

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

// func NewTaskFromRule(scope string, rule_name string, rule *Rule) Task {
// 	var task_name = scope + ":" + rule_name
// }
