package stats

import (
	"sync"
	"time"
)

const (
	MeasureKindStage = iota
	MeasureKindTask  = iota
)

type StatsMeasure struct {
	Name     string
	Start    time.Time
	Duration time.Duration
	kind     int
}

type Stats struct {
	measures sync.Map
}

func New() Stats {
	return Stats{
		measures: sync.Map{},
	}
}

func (s *Stats) Start(name string, kind int) {
	s.measures.Store(name, StatsMeasure{
		Name:  name,
		kind:  kind,
		Start: time.Now(),
	})
}

func (s *Stats) Stop(name string) time.Duration {
	if _m, ok := s.measures.Load(name); ok {
		var m = _m.(StatsMeasure)
		m.Duration = time.Since(m.Start)
		s.measures.Store(name, m)
		return m.Duration.Truncate(time.Millisecond)
	}

	return 0
}

func (s *Stats) Get(name string) StatsMeasure {
	var m, ok = s.measures.Load(name)

	if ok {
		return m.(StatsMeasure)
	}

	var nil_measure StatsMeasure
	return nil_measure
}

func (s *Stats) GetTasksSumDuration() time.Duration {
	var res time.Duration = 0

	s.measures.Range(func(key, value any) bool {
		var m = value.(StatsMeasure)
		if m.kind == MeasureKindTask {
			res += m.Duration
		}
		return true
	})

	return res
}
