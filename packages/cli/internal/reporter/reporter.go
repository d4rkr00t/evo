package reporter

import (
	"evo/internal/logger"
	"evo/internal/spinner"
	"evo/internal/stats"
	"evo/internal/task_graph"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
)

const (
	ReporterOutputStreamAll      = iota
	ReporterOutputStreamTopLevel = iota
	ReporterOutputCombine        = iota
	ReporterOutputOnlyErrors     = iota
)

type Reporter struct {
	logger         logger.Logger
	spinner        *spinner.Spinner
	lock           sync.Mutex
	Output         int
	spinnerEnabled bool
}

func New(logger logger.Logger) Reporter {
	return Reporter{
		logger:         logger,
		spinner:        spinner.New([]string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}),
		lock:           sync.Mutex{},
		Output:         ReporterOutputOnlyErrors,
		spinnerEnabled: false,
	}
}

func (rr *Reporter) EnableSpinner() {
	rr.spinnerEnabled = true
}

func (rr *Reporter) SetOutput(output int) {
	rr.Output = output
}

func (rr *Reporter) Start() {
	rr.spinnerStart()
}

func (rr *Reporter) shouldStreamOutput(task *task_graph.Task) bool {
	return rr.Output == ReporterOutputStreamAll || (rr.Output == ReporterOutputStreamTopLevel && task.TopLevel)
}

func (rr *Reporter) spinnerStart() {
	if !rr.spinnerEnabled {
		return
	}
	rr.spinner.Start()
}
func (rr *Reporter) spinnerStop() {
	if !rr.spinnerEnabled {
		return
	}
	rr.spinner.Stop()
}
func (rr *Reporter) spinnerPause() {
	if !rr.spinnerEnabled {
		return
	}
	rr.spinner.Pause()
}
func (rr *Reporter) spinnerErase() {
	if !rr.spinnerEnabled {
		return
	}
	rr.spinner.Erase()
}
func (rr *Reporter) spinnerResume() {
	if !rr.spinnerEnabled {
		return
	}
	rr.spinner.Resume()
}
func (rr *Reporter) spinnerUpdate(newOutput string) {
	if !rr.spinnerEnabled {
		return
	}
	rr.spinner.Update(newOutput)
}

func (rr *Reporter) StreamLog(task *task_graph.Task, lines ...string) {
	if !rr.shouldStreamOutput(task) {
		return
	}

	rr.lock.Lock()
	rr.spinnerPause()
	rr.spinnerErase()
	for _, line := range lines {
		rr.logger.Log(
			fmt.Sprintf(
				"  %s %s %s",
				logger.ColorMessage(task.Color, task.Name()),
				color.HiBlackString("←"),
				line,
			),
		)
	}
	rr.spinnerResume()
	rr.lock.Unlock()
}

func (rr *Reporter) StreamError(task *task_graph.Task, lines ...string) {
	if !rr.shouldStreamOutput(task) {
		return
	}

	rr.lock.Lock()
	rr.spinnerPause()
	rr.spinnerErase()
	for _, line := range lines {
		rr.logger.Log(
			fmt.Sprintf(
				"  %s %s %s",
				logger.ColorMessage(task.Color, task.Name()),
				color.HiBlackString("←"),
				color.RedString(line),
			),
		)
	}
	rr.spinnerResume()
	rr.lock.Unlock()
}

func (rr *Reporter) SuccessTask(task *task_graph.Task) {
	if rr.Output != ReporterOutputStreamAll && !(rr.Output == ReporterOutputStreamTopLevel && task.TopLevel) {
		return
	}

	rr.lock.Lock()
	rr.spinnerPause()
	rr.spinnerErase()
	var fromCache = ""
	if task.RestoredFromCache == task_graph.TaskCacheHitCopy {
		fromCache = color.HiBlackString("cache hit [%s] outputs don't match, restoring…", task.WsHash[0:6])
	} else if task.RestoredFromCache == task_graph.TaskCacheHitSkip || task.RestoredFromCache == task_graph.TaskCacheHit {
		fromCache = color.HiBlackString("cache hit [%s]", task.WsHash[0:6])
	}

	rr.logger.Log(
		fmt.Sprintf(
			"%s %s %s %s",
			color.HiGreenString("✓"),
			task.Name(),
			color.HiBlackString("(%s)", task.Duration.Truncate(time.Millisecond)),
			fromCache,
		),
	)

	if rr.Output == ReporterOutputCombine {
		for _, line := range strings.Split(task.Output, "\n") {
			rr.logger.Log(
				fmt.Sprintf(
					"  %s %s %s",
					logger.ColorMessage(task.Color, task.Name()),
					color.HiBlackString("←"),
					line,
				),
			)
		}
	}

	rr.spinnerResume()
	rr.lock.Unlock()
}

func (rr *Reporter) FailTask(task *task_graph.Task) {
	rr.lock.Lock()
	rr.spinnerPause()
	rr.spinnerErase()
	var fromCache = ""
	if task.RestoredFromCache == task_graph.TaskCacheHitCopy {
		fromCache = color.HiBlackString("cache hit [%s] restoring outputs…", task.WsHash[0:6])
	} else if task.RestoredFromCache == task_graph.TaskCacheHitSkip || task.RestoredFromCache == task_graph.TaskCacheHit {
		fromCache = color.HiBlackString("cache hit [%s]", task.WsHash[0:6])
	}
	rr.logger.Log(
		color.HiRedString(
			"⨯ %s %s %s",
			task.Name(),
			color.HiBlackString("(%s)", task.Duration.Truncate(time.Millisecond)),
			fromCache,
		),
	)

	if rr.Output == ReporterOutputCombine {
		for _, line := range strings.Split(task.Output, "\n") {
			rr.logger.Clone().Log(
				fmt.Sprintf(
					"  %s %s %s",
					logger.ColorMessage(task.Color, task.Name()),
					color.HiBlackString("←"),
					line,
				),
			)
		}

		if task.Error != nil {
			for _, line := range strings.Split(task.Error.Error(), "\n") {
				rr.logger.Log(
					fmt.Sprintf(
						"  %s %s %s",
						logger.ColorMessage(task.Color, task.Name()),
						color.HiBlackString("←"),
						color.RedString(line),
					),
				)
			}
		}
	}
	rr.spinnerResume()
	rr.lock.Unlock()
}

func (rr *Reporter) SuccessRun(st *stats.Stats, taskGraph *task_graph.TaskGraph) {
	rr.spinnerStop()
	rr.spinnerErase()

	var taskParallelTime = st.Get("runtasks").Duration
	var totalTime = st.Get("total").Duration
	var taskSeqTime = st.GetTasksSumDuration()
	var diff = taskSeqTime - taskParallelTime
	var restoredFromCache = 0

	for _, taskName := range taskGraph.TasksNamesList {
		var task, _ = taskGraph.Load(taskName)
		if task.RestoredFromCache != task_graph.TaskCacheMiss {
			restoredFromCache += 1
		}
	}

	rr.logger.Log()
	rr.logger.Badge("Tasks time").Log(
		color.BlueString(
			"%s %s | %s %s |",
			"total",
			taskSeqTime.Truncate(time.Millisecond).String(),
			"concurent",
			taskParallelTime.Truncate(time.Millisecond).String(),
		),
		color.GreenString("saved %s", diff.Truncate(time.Millisecond).String()),
	)

	if restoredFromCache > 0 {
		rr.logger.Badge("Restored from cache").Log(
			color.BlueString(
				"%d of %d",
				restoredFromCache,
				len(taskGraph.TasksNamesList),
			),
		)
	}

	rr.logger.Log()
	rr.logger.Log(color.GreenString("✓ Completed in %s", totalTime.Truncate(time.Millisecond).String()))
}

func (rr *Reporter) FailRun(dur time.Duration, taskGraph *task_graph.TaskGraph) {
	rr.spinnerStop()
	rr.spinnerErase()

	rr.logger.Log()
	rr.logger.Log("Errors:")
	rr.logger.Log()

	for _, taskName := range taskGraph.TasksNamesList {
		var task, _ = taskGraph.Load(taskName)
		if task.Status != task_graph.TaskStatsuFailure {
			continue
		}
		if len(task.Output) > 0 {
			rr.logger.Badge(taskName).Log(task.Output)
		}
		rr.logger.Badge(taskName).Log(task.Error.Error())
		rr.logger.Log()
	}

	rr.logger.Log(color.RedString("⨯ Failed in: %s", dur.Truncate(time.Millisecond).String()))
}

func (rr *Reporter) UpdateFromTaskGraph(taskGraph *task_graph.TaskGraph) {
	rr.lock.Lock()
	var newOutput = []string{}
	var countPending = 0
	var countRunning = 0
	var countSuccess = 0
	var countFailed = 0

	for _, taskName := range taskGraph.TasksNamesList {
		var task, _ = taskGraph.Load(taskName)
		if task.Status == task_graph.TaskStatsuPending {
			countPending += 1
		} else if task.Status == task_graph.TaskStatsuRunning {
			countRunning += 1
		} else if task.Status == task_graph.TaskStatsuFailure {
			countFailed += 1
		} else if task.Status == task_graph.TaskStatsuSuccess {
			countSuccess += 1
		}
	}

	newOutput = append(
		newOutput,
		fmt.Sprintf(
			"%s %s %s %s",
			color.YellowString("▸"),
			color.HiBlackString("status:"),
			color.GreenString("%d succeeded", countSuccess),
			color.RedString("%d failed", countFailed),
		),
		fmt.Sprintf(
			"%s %s %s %s",
			color.YellowString("▸"),
			color.HiBlackString("targets:"),
			color.GreenString("%d running", countRunning),
			color.YellowString("%d queued", countPending),
		),
	)
	newOutput = append(newOutput, "")

	for _, taskName := range taskGraph.TasksNamesList {
		var task, _ = taskGraph.Load(taskName)
		if task.Status == task_graph.TaskStatsuRunning {
			newOutput = append(newOutput, fmt.Sprintf("$#$ %s %s", color.HiBlackString("evo run"), task.Name()))
		}
	}
	newOutput = append(newOutput, "")
	rr.spinnerUpdate(strings.Join(newOutput, "\n"))
	rr.lock.Unlock()
}
