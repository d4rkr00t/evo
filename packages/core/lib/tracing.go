package lib

import (
	"path"

	"github.com/google/chrometracing"
	"github.com/otiai10/copy"
)

type Tracing struct {
	is_enabled bool
	out        string
}

func NewTracing() Tracing {
	return Tracing{
		is_enabled: false,
	}
}

func (t *Tracing) SetOut(name string) {
	t.out = name
}

func (t *Tracing) Event(name string) *chrometracing.PendingEvent {
	return chrometracing.Event(name)
}

func (t *Tracing) Enable() {
	t.is_enabled = true
	chrometracing.EnableTracing()
}

func (t *Tracing) Path() string {
	return chrometracing.Path()
}

func (t *Tracing) Write(lg *Logger, cwd string) {
	if !t.is_enabled {
		return
	}
	var out = t.out
	if out == "" {
		out = "evo-tracing-output.trace"
	}
	lg.LogWithBadge("trace", path.Join(cwd, out))
	copy.Copy(chrometracing.Path(), path.Join(cwd, out))
}
