package lib

import (
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
	finished []*StatsMeasure
	measures map[string]StatsMeasure
}

func NewStats() Stats {
	return Stats{
		finished: []*StatsMeasure{},
		measures: map[string]StatsMeasure{},
	}
}

func (s *Stats) StartMeasure(name string, kind int) {
	s.measures[name] = StatsMeasure{
		name:  name,
		kind:  kind,
		start: time.Now(),
	}
}

func (s *Stats) StopMeasure(name string) time.Duration {
	if m, ok := s.measures[name]; ok {
		m.duration = time.Since(m.start)
		s.measures[name] = m
		s.finished = append(s.finished, &m)

		return m.duration.Truncate(time.Millisecond)
	}

	return 0
}

func (s Stats) GetMeasure(name string) StatsMeasure {
	return s.measures[name]
}

func (s Stats) GetTasksSumDuration() time.Duration {
	var res time.Duration = 0

	for _, m := range s.measures {
		if m.kind == MEASURE_KIND_TASK {
			res += m.duration
		}
	}

	return res
}
