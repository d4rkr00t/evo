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
	fmt.Printf("%s %s: %s\n", color.CyanString("╺"), color.HiBlackString(strings.ToLower(badge)), strings.Join(msg, " "))
}

func (l Logger) LogWithBadgeVerbose(badge string, msg ...string) {
	if l.verbose {
		l.LogWithBadge(badge, msg...)
	}
}

func (l Logger) Log(msg ...string) {
	fmt.Println(strings.Join(msg, " "))
}

func (l Logger) LogVerbose(msg ...string) {
	if l.verbose {
		l.Log(msg...)
	}
}

func (l Logger) CreateGroup() LoggerGroup {
	return LoggerGroup{logger: &l}
}

//
// Logger Group
//

type LoggerGroup struct {
	logger *Logger
}

func (lg *LoggerGroup) Start(msg ...string) {
	fmt.Printf("\n%s %s\n", color.HiBlackString("┌"), strings.Join(msg, " "))
	lg.Log()
}

func (lg LoggerGroup) Log(msg ...string) {
	fmt.Printf("%s %s\n", color.HiBlackString("│"), strings.Join(msg, " "))
}

func (lg LoggerGroup) LogVerbose(msg ...string) {
	if lg.logger.verbose {
		lg.Log(msg...)
	}
}

func (lg LoggerGroup) Warn(msg ...string) {
	fmt.Printf("%s %s\n", color.HiBlackString("│"), color.CyanString(strings.Join(msg, " ")))
}

func (lg LoggerGroup) LogWithBadge(badge string, msg ...string) {
	fmt.Printf("%s %s: %s\n", color.HiBlackString("│"), color.CyanString(badge), strings.Join(msg, " "))
}

func (lg LoggerGroup) SuccessWithBadge(badge string, msg ...string) {
	fmt.Printf("%s %s: %s\n", color.HiBlackString("│"), color.GreenString(badge), strings.Join(msg, " "))
}

func (lg LoggerGroup) InfoWithBadge(badge string, msg ...string) {
	fmt.Printf("%s %s: %s\n", color.HiBlackString("│"), color.BlueString(badge), strings.Join(msg, " "))
}

func (lg LoggerGroup) LogWithBadgeVerbose(badge string, msg ...string) {
	if lg.logger.verbose {
		lg.LogWithBadge(badge, msg...)
	}
}

func (lg LoggerGroup) End(dur time.Duration) {
	lg.Log()
	if dur != 0 {
		fmt.Printf("%s Completed in %s\n", color.HiBlackString(strings.ToLower("└")), color.GreenString(dur.String()))
	} else {
		fmt.Printf("%s Completed\n", color.HiBlackString(strings.ToLower("└")))
	}
}
