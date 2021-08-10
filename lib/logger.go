package lib

import (
	"fmt"
	"strings"
	"time"

	"github.com/fatih/color"
)

type Logger struct {
	verbose bool
}

func NewLogger(verbose bool) Logger {
	return Logger{
		verbose,
	}
}

func (l Logger) LogWithBadge(badge string, msg ...string) {
	fmt.Printf("%s %s: %s\n", color.CyanString("╺"), color.HiBlackString(strings.ToLower(badge)), strings.Join(msg, ""))
}

func (l Logger) Log(msg ...string) {
	fmt.Println(strings.Join(msg, ""))
}

func (l Logger) CreateGroup() LoggerGroup {
	return LoggerGroup{logger: &l}
}

//
// Logger Group
//

type LoggerGroup struct {
	logger *Logger
	start  time.Time
}

func (lg *LoggerGroup) Start(msg ...string) {
	lg.start = time.Now()
	fmt.Printf("\n%s %s\n", color.HiBlackString(strings.ToLower("┌")), strings.Join(msg, " "))
	lg.Log()
}

func (lg LoggerGroup) Log(msg ...string) {
	fmt.Printf("%s %s\n", color.HiBlackString(strings.ToLower("│")), strings.Join(msg, " "))
}

func (lg LoggerGroup) LogVerbose(msg ...string) {
	if lg.logger.verbose {
		lg.Log(msg...)
	}
}

func (lg LoggerGroup) Warn(msg ...string) {
	fmt.Printf("%s %s\n", color.HiBlackString(strings.ToLower("│")), color.CyanString(strings.Join(msg, " ")))
}

func (lg LoggerGroup) LogWithBadge(badge string, msg ...string) {
	fmt.Printf("%s %s: %s\n", color.HiBlackString(strings.ToLower("│")), color.CyanString(badge), strings.Join(msg, " "))
}

func (lg LoggerGroup) SuccessWithBadge(badge string, msg ...string) {
	fmt.Printf("%s %s: %s\n", color.HiBlackString(strings.ToLower("│")), color.GreenString(badge), strings.Join(msg, " "))
}

func (lg LoggerGroup) InfoWithBadge(badge string, msg ...string) {
	fmt.Printf("%s %s: %s\n", color.HiBlackString(strings.ToLower("│")), color.BlueString(badge), strings.Join(msg, " "))
}

func (lg LoggerGroup) LogWithBadgeVerbose(badge string, msg ...string) {
	if lg.logger.verbose {
		lg.LogWithBadge(badge, msg...)
	}
}

func (lg LoggerGroup) End() {
	lg.Log()
	fmt.Printf("%s Completed in %s\n", color.HiBlackString(strings.ToLower("└")), color.GreenString(time.Since(lg.start).String()))
}
