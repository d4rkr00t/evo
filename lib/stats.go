package lib

import (
	"sync"
	"time"
)

const (
	MEASURE_KIND_STAGE = iota
	MEASURE_KIND_TASK  = iota
)

type StatsMeasure struct {
	name     string
	start    time.Time
	duration time.Duration
	kind     int
}

type Stats struct {
	measures sync.Map
}

func NewStats() Stats {
	var measures sync.Map
	return Stats{
		measures,
	}
}

func (s *Stats) StartMeasure(name string, kind int) {
	s.measures.Store(name, StatsMeasure{
		name:  name,
		kind:  kind,
		start: time.Now(),
	})
}

func (s *Stats) StopMeasure(name string) time.Duration {
	if _m, ok := s.measures.Load(name); ok {
		var m = _m.(StatsMeasure)
		m.duration = time.Since(m.start)
		s.measures.Store(name, m)
		return m.duration.Truncate(time.Millisecond)
	}

	return 0
}

func (s Stats) GetMeasure(name string) StatsMeasure {
	var m, ok = s.measures.Load(name)
	if ok {
		return m.(StatsMeasure)
	}
	var nil_measure StatsMeasure
	return nil_measure
}

func (s Stats) GetTasksSumDuration() time.Duration {
	var res time.Duration = 0

	s.measures.Range(func(key, value interface{}) bool {
		var m = value.(StatsMeasure)
		if m.kind == MEASURE_KIND_TASK {
			res += m.duration
		}
		return true
	})

	return res
}
