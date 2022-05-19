package tracer

import (
	"evo/internal/logger"
	"path"

	"github.com/google/chrometracing"
	"github.com/otiai10/copy"
)

type Tracer struct {
	is_enabled bool
	out        string
}

func New() Tracer {
	return Tracer{
		is_enabled: false,
	}
}

func (t *Tracer) SetOut(name string) {
	t.out = name
}

func (t *Tracer) Event(name string) *chrometracing.PendingEvent {
	return chrometracing.Event(name)
}

func (t *Tracer) Enable() {
	t.is_enabled = true
	chrometracing.EnableTracing()
}

func (t *Tracer) Path() string {
	return chrometracing.Path()
}

func (t *Tracer) Write(lg *logger.Logger, cwd string) {
	if !t.is_enabled {
		return
	}
	var out = t.out
	if out == "" {
		out = "evo-tracing-output.trace"
	}
	lg.Badge("trace").Log(path.Join(cwd, out))
	copy.Copy(chrometracing.Path(), path.Join(cwd, out))
}
