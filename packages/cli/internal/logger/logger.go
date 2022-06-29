package logger

import (
	"fmt"
	"strings"
	"time"

	"github.com/fatih/color"
)

type Logger struct {
	verbose   bool
	debug     bool
	badge     string
	condition func() bool
}

func NewLogger(verbose bool, debug bool) Logger {
	return Logger{
		verbose:   verbose,
		debug:     debug,
		badge:     "",
		condition: func() bool { return true },
	}
}

func (l *Logger) reset() {
	l.condition = func() bool { return true }
	l.badge = ""
}

func (l Logger) Clone() *Logger {
	return &l
}

func (l *Logger) Verbose() *Logger {
	l.condition = func() bool {
		return l.verbose || l.debug
	}
	return l
}

func (l *Logger) Debug() *Logger {
	l.condition = func() bool {
		return l.debug
	}
	return l
}

func (l *Logger) Badge(badge string) *Logger {
	l.badge = badge
	return l
}

func (l *Logger) Log(msg ...string) {
	if !l.condition() {
		l.reset()
		return
	}

	var processed_msg = strings.Join(msg, " ")

	for _, line := range strings.Split(processed_msg, "\n") {
		if len(l.badge) > 0 {
			fmt.Printf("%s %s: %s\n", color.YellowString("▸"), color.HiBlackString(strings.ToLower(l.badge)), line)
		} else {
			fmt.Println(line)
		}
	}

	l.reset()
}

func (l *Logger) CreateGroup() LoggerGroup {
	return LoggerGroup{
		logger: l,
		badge:  "",
		condition: func() bool {
			return true
		},
	}
}

//
// Logger Group
//

type LoggerGroup struct {
	logger     *Logger
	condition  func() bool
	badge      string
	badgeColor string
}

func (lg LoggerGroup) Clone() *LoggerGroup {
	return &lg
}

func (lg *LoggerGroup) Verbose() *LoggerGroup {
	lg.condition = func() bool {
		return lg.logger.verbose || lg.logger.debug
	}
	return lg
}

func (lg *LoggerGroup) Debug() *LoggerGroup {
	lg.condition = func() bool {
		return lg.logger.debug
	}
	return lg
}

func (lg *LoggerGroup) Badge(badge string) *LoggerGroup {
	lg.badge = badge
	return lg
}

func (lg *LoggerGroup) BadgeColor(color string) *LoggerGroup {
	lg.badgeColor = color
	return lg
}

func (lg *LoggerGroup) __color_badge__(colorName string) string {
	return ColorMessage(colorName, lg.badge)
}

func (lg *LoggerGroup) __reset__() {
	lg.condition = func() bool {
		return true
	}
	lg.badge = ""
	lg.badgeColor = ""
}

func (lg *LoggerGroup) end(dur time.Duration) {
	if dur != 0 {
		fmt.Printf("%s Completed in %s\n", color.HiBlackString("└"), color.GreenString(dur.Truncate(time.Millisecond).String()))
	} else {
		fmt.Printf("%s Completed\n", color.HiBlackString("└"))
	}
}

func (lg *LoggerGroup) fail(dur time.Duration) {
	if dur != 0 {
		fmt.Printf("%s %s %s\n", color.HiBlackString("└"), color.RedString("Failed in"), color.GreenString(dur.Truncate(time.Millisecond).String()))
	} else {
		fmt.Printf("%s %s\n", color.HiBlackString("└"), color.RedString("Failed"))
	}
}

func (lg *LoggerGroup) Start(msg ...string) {
	if !lg.condition() {
		lg.__reset__()
		return
	}
	fmt.Printf("\n%s %s\n", color.HiBlackString("┌"), strings.Join(msg, " "))
	lg.Log()
}

func (lg *LoggerGroup) Log(msg ...string) {
	if !lg.condition() {
		lg.__reset__()
		return
	}

	var processed_msg = strings.Join(msg, " ")

	for _, line := range strings.Split(processed_msg, "\n") {
		if len(lg.badge) > 0 {
			fmt.Printf("%s %s: %s\n", color.HiBlackString("│"), lg.__color_badge__(lg.badgeColor), line)
		} else {
			fmt.Printf("%s %s\n", color.HiBlackString("│"), line)
		}
	}

	lg.__reset__()
}

func (lg *LoggerGroup) Info(msg ...string) {
	if len(lg.badge) > 0 {
		lg.badgeColor = "blue"
		lg.badge = lg.__color_badge__(lg.badgeColor)
	}
	lg.Log(msg...)
}

func (lg *LoggerGroup) Success(msg ...string) {
	if len(lg.badge) > 0 {
		lg.badgeColor = "green"
		lg.badge = lg.__color_badge__(lg.badgeColor)
	}
	lg.Log(msg...)
}

func (lg *LoggerGroup) Warn(msg ...string) {
	if len(lg.badge) > 0 {
		lg.badgeColor = "cyan"
		lg.badge = lg.__color_badge__(lg.badgeColor)
	}
	if len(msg) > 0 {
		msg[0] = color.CyanString(msg[0])
		lg.Log(msg...)
	}
}

func (lg *LoggerGroup) Error(msg ...string) {
	if len(lg.badge) > 0 {
		lg.badgeColor = "red"
		lg.badge = lg.__color_badge__(lg.badgeColor)
	}
	lg.Log(msg...)
}

func (lg *LoggerGroup) EndEmpty(dur time.Duration) {
	if !lg.condition() {
		lg.__reset__()
		return
	}
	lg.end(dur)
}

func (lg *LoggerGroup) End(dur time.Duration) {
	if !lg.condition() {
		lg.__reset__()
		return
	}
	lg.Log()
	lg.end(dur)
}

func (lg *LoggerGroup) Fail(dur time.Duration) {
	if !lg.condition() {
		lg.__reset__()
		return
	}
	lg.Log()
	lg.fail(dur)
}

func (lg *LoggerGroup) EndPlain() {
	if !lg.condition() {
		lg.__reset__()
		return
	}
	lg.Log()
	fmt.Printf("%s ●\n", color.HiBlackString(strings.ToLower("└")))
}

func (lg *LoggerGroup) EndPlainEmpty() {
	if !lg.condition() {
		lg.__reset__()
		return
	}
	fmt.Printf("%s ●\n", color.HiBlackString(strings.ToLower("└")))
}

func ColorMessage(colorName string, msg string) string {
	switch colorName {
	case "red":
		return color.RedString(msg)
	case "cyan":
		return color.CyanString(msg)
	case "green":
		return color.GreenString(msg)
	case "magenta":
		return color.MagentaString(msg)
	case "yellow":
		return color.YellowString(msg)
	case "blue":
		return color.BlueString(msg)
	}
	return msg
}
