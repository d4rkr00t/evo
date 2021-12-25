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
	return LoggerGroup{
		logger: &l,
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

func (lg LoggerGroup) Verbose() LoggerGroup {
	lg.condition = func() bool {
		return lg.logger.verbose
	}
	return lg
}

func (lg LoggerGroup) Badge(badge string) LoggerGroup {
	lg.badge = badge
	return lg
}

func (lg LoggerGroup) BadgeColor(color string) LoggerGroup {
	lg.badgeColor = color
	return lg
}

func (lg LoggerGroup) __color_badge__(color_name string, color_fn func(string, ...interface{}) string) string {
	switch color_name {
	case "red":
		return color.RedString(lg.badge)
	case "cyan":
		return color.CyanString(lg.badge)
	case "green":
		return color.GreenString(lg.badge)
	case "magenta":
		return color.MagentaString(lg.badge)
	case "yellow":
		return color.YellowString(lg.badge)
	case "blue":
		return color.BlueString(lg.badge)
	}
	return color_fn(lg.badge)
}

func (lg *LoggerGroup) __reset__() {
	lg.condition = func() bool {
		return true
	}
	lg.badge = ""
}

func (lg *LoggerGroup) end(dur time.Duration) {
	if dur != 0 {
		fmt.Printf("%s Completed in %s\n", color.HiBlackString("└"), color.GreenString(dur.Truncate(time.Millisecond).String()))
	} else {
		fmt.Printf("%s Completed\n", color.HiBlackString("└"))
	}
}

func (lg *LoggerGroup) Start(msg ...string) {
	fmt.Printf("\n%s %s\n", color.HiBlackString("┌"), strings.Join(msg, " "))
	lg.Log()
}

func (lg LoggerGroup) Log(msg ...string) {
	if !lg.condition() {
		lg.__reset__()
		return
	}

	var processed_msg = strings.Join(msg, " ")

	for _, line := range strings.Split(processed_msg, "\n") {
		if len(lg.badge) > 0 {
			fmt.Printf("%s %s: %s\n", color.HiBlackString("│"), lg.badge, line)
		} else {

			fmt.Printf("%s %s\n", color.HiBlackString("│"), line)
		}
	}

	lg.__reset__()
}

func (lg LoggerGroup) Info(msg ...string) {
	if len(lg.badge) > 0 {
		lg.badge = lg.__color_badge__(lg.badgeColor, color.BlueString)
	}
	lg.Log(msg...)
}

func (lg LoggerGroup) Success(msg ...string) {
	if len(lg.badge) > 0 {
		lg.badge = lg.__color_badge__(lg.badgeColor, color.GreenString)
	}
	lg.Log(msg...)
}

func (lg LoggerGroup) Warn(msg ...string) {
	if len(lg.badge) > 0 {
		lg.badge = lg.__color_badge__(lg.badgeColor, color.CyanString)
	}
	if len(msg) > 0 {
		msg[0] = color.CyanString(msg[0])
		lg.Log(msg...)
	}
}

func (lg LoggerGroup) Error(msg ...string) {
	if len(lg.badge) > 0 {
		lg.badge = lg.__color_badge__(lg.badgeColor, color.RedString)
	}
	lg.Log(msg...)
}

func (lg *LoggerGroup) End(dur time.Duration) {
	lg.Log()
	lg.end(dur)
}

func (lg *LoggerGroup) EndPlain() {
	lg.Log()
	fmt.Printf("%s ●\n", color.HiBlackString(strings.ToLower("└")))
}
