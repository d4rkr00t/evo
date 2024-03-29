package task_graph

import (
	"encoding/json"
	"evo/internal/cache"
	"evo/internal/fsutils"
	"evo/internal/hash_utils"
	"evo/internal/label"
	"evo/internal/target"
	"evo/internal/workspace"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/otiai10/copy"
)

type Task struct {
	TopLevel          bool
	Ws                *workspace.Workspace
	DepsHash          string
	Hash              string
	TargetName        string
	Target            *target.Target
	Status            int
	Deps              []string
	Color             string
	RestoredFromCache int
	Duration          time.Duration
	Output            string
	Error             error
}

type TaskCacheDiagnostics struct {
	TaskHash        string
	TaskDepsHash    string
	TaskTarget      string
	WsHash          string
	WsFilesHash     string
	WsExtDepsHash   string
	WsLocalDepsHash string
	WsFiles         []string
}

const (
	TaskStatsuPending = iota
	TaskStatsuRunning = iota
	TaskStatsuSuccess = iota
	TaskStatsuFailure = iota
)

var taskBadgeColors = []string{
	"cyan",
	"green",
	"magenta",
	"yellow",
	"blue",
}

const (
	TaskStdoutPostfix           = "__stdout"
	TaskStderrPostfix           = "__stderr"
	TaskOutputsPostfix          = "__outputs"
	TaskOutputsHashPostfix      = "__outputs__hash"
	TaskCacheDiagnosticsPostfix = "__hash__state"
)

const (
	TaskCacheMiss    = iota
	TaskCacheHit     = iota
	TaskCacheHitCopy = iota
	TaskCacheHitSkip = iota
)

func NewTask(ws *workspace.Workspace, targetName string, target *target.Target, topLevel bool) Task {
	return Task{
		TopLevel:          topLevel,
		Ws:                ws,
		TargetName:        targetName,
		Target:            target,
		Status:            TaskStatsuPending,
		Color:             taskBadgeColors[hash_utils.StrToNum(GetTaskName(ws.Name, targetName))%len(taskBadgeColors)],
		RestoredFromCache: TaskCacheMiss,
	}
}

func (t *Task) AddDependency(depName string) {
	t.Deps = append(t.Deps, depName)
}

func (t *Task) UpdateStatus(status int) {
	t.Status = status
}

func (t *Task) Name() string {
	return GetTaskName(t.Ws.Name, t.TargetName)
}

func (t *Task) String() string {
	return t.Name()
}

func (t *Task) CleanName() string {
	return strings.Replace(t.Name(), "/", "__", -1)
}

func (t *Task) HasOutputs() bool {
	return len(t.Target.Outputs) > 0
}

func (t *Task) GetCacheKey() string {
	if t.Hash == "" {
		panic(fmt.Sprintf("Hash for a task '%s' is empty", t.Name()))
	}

	return fmt.Sprintf("%s__%s", t.CleanName(), t.Hash)
}

func (t *Task) Rehash(tg *TaskGraph) string {
	t.DepsHash = t.getDepsHash(tg)
	t.Hash = hash_utils.HashStringList([]string{
		t.Ws.Hash,
		t.Target.String(),
		t.DepsHash,
	})
	tg.Store(t)
	return t.Hash
}

func (t *Task) Invalidate(cc *cache.Cache) bool {
	return !cc.Has(t.GetCacheKey())
}

func (t *Task) getDepsHash(tg *TaskGraph) string {
	var depsList = []string{}

	for _, depName := range t.Deps {
		var dep, _ = tg.Load(depName)
		depsList = append(depsList, dep.Name()+":"+dep.Hash)
	}

	sort.Strings(depsList)
	return hash_utils.HashStringList(depsList)
}

func (t *Task) Cache(cc *cache.Cache, stdout string, stderr string) {
	var exitCode = 0
	if t.Status == TaskStatsuFailure {
		exitCode = 1
	}

	cc.CacheData(t.GetCacheKey(), fmt.Sprintf("%d", exitCode))
	cc.CacheData(t.GetCacheKey()+TaskStdoutPostfix, stdout)
	cc.CacheData(t.GetCacheKey()+TaskStderrPostfix, stderr)
	t.storeCacheDiagnostics(cc)

	if t.HasOutputs() && len(stderr) == 0 {
		t.cacheOutputs(cc)
	}
}

func (t *Task) storeCacheDiagnostics(cc *cache.Cache) {
	var cacheDiagnostics = TaskCacheDiagnostics{
		TaskHash:        t.Hash,
		TaskTarget:      t.Target.String(),
		TaskDepsHash:    t.DepsHash,
		WsHash:          t.Ws.Hash,
		WsFilesHash:     t.Ws.FilesHash,
		WsExtDepsHash:   t.Ws.ExtDepsHash,
		WsLocalDepsHash: t.Ws.LocalDepsHash,
		WsFiles:         t.Ws.Files,
	}

	var cacheDiagnosticsSerialized, err = json.Marshal(cacheDiagnostics)
	if err == nil {
		cc.CacheData(t.CleanName()+"__"+TaskCacheDiagnosticsPostfix, string(cacheDiagnosticsSerialized))
	}
}

func (t *Task) RetriveCacheDiagnostics(cc *cache.Cache) (TaskCacheDiagnostics, error) {
	var cacheKey = t.CleanName() + "__" + TaskCacheDiagnosticsPostfix
	if !cc.Has(cacheKey) {
		return TaskCacheDiagnostics{}, fmt.Errorf("no hash state found")
	}
	var state = cc.ReadData(cacheKey)
	var cachDiagnostics TaskCacheDiagnostics
	var err = json.Unmarshal([]byte(state), &cachDiagnostics)
	return cachDiagnostics, err
}

func (t *Task) GetStatusAndLogs(cc *cache.Cache) (string, string, string, error) {
	if !cc.Has(t.GetCacheKey()) {
		return "", "", "", fmt.Errorf("no cache for a task")
	}
	var exitCode = cc.ReadData(t.GetCacheKey())
	var outLog = cc.ReadData(t.GetCacheKey() + TaskStdoutPostfix)
	var errLog = cc.ReadData(t.GetCacheKey() + TaskStderrPostfix)

	return exitCode, outLog, errLog, nil
}

func (t *Task) RestoreOutputs(cc *cache.Cache) {
	if !t.HasOutputs() {
		return
	}
	var cacheKey = t.GetCacheKey() + TaskOutputsPostfix
	for _, output := range t.Target.Outputs {
		copy.Copy(path.Join(cc.GetCachePath(cacheKey), output), path.Join(t.Ws.Path, output))
	}
}

func (t *Task) ValidateOutputs() error {
	if !t.HasOutputs() {
		return nil
	}

	var missing = []string{}

	for _, output := range t.Target.Outputs {
		if !fsutils.Exist(path.Join(t.Ws.Path, output)) {
			missing = append(missing, output)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("task \"%s\" didn't produce expected outputs: %s", color.CyanString(t.Name()), color.YellowString(strings.Join(missing, ", ")))
	}

	return nil
}

func (t *Task) CleanOutputs() {
	if !t.HasOutputs() {
		return
	}

	for _, output := range t.Target.Outputs {
		os.RemoveAll(path.Join(t.Ws.Path, output))
	}
}

func (t *Task) GetOutputsHash() string {
	if !t.HasOutputs() {
		return ""
	}

	var globs = []string{}
	for _, out := range t.Target.Outputs {
		globs = append(globs, path.Join(out, "**", "*"))
		globs = append(globs, out)
	}

	var files []string = fsutils.GlobFiles(t.Ws.Path, &globs, &[]string{"node_modules/**"})
	sort.Strings(files)

	return fsutils.GetFileListHash(files)
}

func (t *Task) ShouldRestoreOutputs(cc *cache.Cache) bool {
	if !t.HasOutputs() {
		return false
	}

	var cacheKey = t.GetCacheKey() + TaskOutputsHashPostfix
	if !cc.Has(cacheKey) {
		return true
	}
	var prevOutputsHash = cc.ReadData(cacheKey)
	var curOutputsHash = t.GetOutputsHash()
	return prevOutputsHash != curOutputsHash
}

func (t *Task) cacheOutputs(cc *cache.Cache) {
	var cacheKeyOutputsDir = t.GetCacheKey() + TaskOutputsPostfix
	var cacheKeyOutputsHash = t.GetCacheKey() + TaskOutputsHashPostfix
	var ignores = cache.CacheDirIgnores{
		"node_modules": true,
	}

	for _, output := range t.Target.Outputs {
		copy.Copy(path.Join(t.Ws.Path, output), path.Join(cc.GetCachePath(cacheKeyOutputsDir), output), copy.Options{
			Skip: func(srcinfo fs.FileInfo, src string, dest string) (bool, error) {
				var relSrc, _ = filepath.Rel(t.Ws.Path, src)
				return ignores[relSrc], nil
			},
		})
	}
	cc.CacheData(cacheKeyOutputsHash, t.GetOutputsHash())
}

func GetTaskName(wsName string, targetName string) string {
	return fmt.Sprintf("%s%s%s", wsName, label.Sep, targetName)
}
