package spinner

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fatih/color"
)

type Spinner struct {
	lastOutput    atomic.Value
	lastOutputRaw atomic.Value
	nextOutputRaw atomic.Value
	writer        io.Writer
	spinnerChars  []string
	Lock          sync.Mutex
	idx           int32
	active        atomic.Value
}

func New(spinnerChars []string) *Spinner {
	var spinner = Spinner{
		lastOutput:    atomic.Value{},
		lastOutputRaw: atomic.Value{},
		nextOutputRaw: atomic.Value{},
		writer:        os.Stdout,
		idx:           0,
		spinnerChars:  spinnerChars,
		Lock:          sync.Mutex{},
		active:        atomic.Value{},
	}

	spinner.nextOutputRaw.Store("")
	spinner.lastOutputRaw.Store("")
	spinner.lastOutput.Store("")
	spinner.active.Store(false)

	return &spinner
}

func (s *Spinner) Start() {
	if s.active.Load() == true {
		return
	}

	s.active.Store(true)

	go func() {
		for s.active.Load() == true {
			s.Draw()
			time.Sleep(100 * time.Millisecond)
		}
	}()
}

func (s *Spinner) Draw() {
	var outputRaw = s.nextOutputRaw.Load().(string)
	var lastOutputRaw = s.lastOutputRaw.Load().(string)

	if outputRaw == lastOutputRaw {
		s.moveToStart(outputRaw)
	} else {
		s.Erase()
	}

	s.Lock.Lock()

	var nextSpinnerChar = s.spinnerChars[atomic.LoadInt32(&s.idx)]
	var output = strings.Replace(outputRaw, "$#$", color.HiBlackString(nextSpinnerChar), -1)

	fmt.Fprint(s.writer, output)

	atomic.SwapInt32(&s.idx, int32(int((atomic.LoadInt32(&s.idx)+1))%len(s.spinnerChars)))
	s.lastOutput.Store(output)
	s.lastOutputRaw.Store(outputRaw)

	s.Lock.Unlock()
}

func (s *Spinner) Pause() {
	s.active.Store(false)
}

func (s *Spinner) Resume() {
	s.active.Store(true)
}

func (s *Spinner) Stop() {
	if s.active.Load() == false {
		return
	}

	s.Draw()
	s.active.Store(false)
}

func (s *Spinner) Update(newOutput string) {
	s.nextOutputRaw.Store(newOutput)
}

func (s *Spinner) moveToStart(output string) {
	s.Lock.Lock()
	var numLines = len(strings.Split(output, "\n")) - 1
	for i := 0; i < numLines; i++ {
		// UP
		fmt.Fprintf(s.writer, "\033[A")
	}
	fmt.Fprintf(s.writer, "\r")
	s.Lock.Unlock()
}

func (s *Spinner) Erase() {
	s.Lock.Lock()
	var output = s.lastOutput.Load().(string)
	var numLines = len(strings.Split(output, "\n")) - 1
	for i := 0; i < numLines; i++ {
		// CLEAR
		fmt.Fprintf(s.writer, "\033[2K")
		// UP
		fmt.Fprintf(s.writer, "\033[A")
	}
	fmt.Fprintf(s.writer, "\033[2K")
	fmt.Fprintf(s.writer, "\r")
	s.lastOutput.Store("")
	s.lastOutputRaw.Store("")
	s.Lock.Unlock()
}
