package ui

import (
	"os"
	"strings"
	"github.com/fatih/color"
)



var (
	noColor    = os.Getenv("NO_COLOR") != ""
	forceColor = isForceColor()
)

func isForceColor() bool {
	fc := strings.TrimSpace(os.Getenv("FORCE_COLOR"))
	return fc != "" && fc != "0"
}

func IsRich() bool {
	if noColor && !forceColor {
		return false
	}
	return color.NoColor == false
}


func Accent(msg string) string {
	return color.New(color.FgHiRed).Sprint(msg)
}

func AccentBright(msg string) string {
	return color.New(color.FgHiRed, color.Bold).Sprint(msg)
}

func AccentDim(msg string) string {
	return color.New(color.FgRed).Sprint(msg)
}

func Info(msg string) string {
	return color.New(color.FgHiYellow).Sprint(msg)
}

func Success(msg string) string {
	return color.New(color.FgGreen).Sprint(msg)
}

func Warn(msg string) string {
	return color.New(color.FgYellow).Sprint(msg)
}

func Error(msg string) string {
	return color.New(color.FgRed).Sprint(msg)
}

func Muted(msg string) string {
	return color.New(color.FgHiBlack).Sprint(msg)
}

func Heading(msg string) string {
	return color.New(color.FgHiRed, color.Bold).Sprint(msg)
}

func Command(msg string) string {
	return color.New(color.FgCyan, color.Bold).Sprint(msg)
}

func Option(msg string) string {
	return color.New(color.FgYellow).Sprint(msg)
}

func Subtle(msg string) string {
	return color.New(color.FgWhite).Sprint(msg)
}

func Bold(msg string) string {
	return color.New(color.FgWhite, color.Bold).Sprint(msg)
}

func Primary(msg string) string {
	return color.New(color.FgMagenta, color.Bold).Sprint(msg)
}

func Secondary(msg string) string {
	return color.New(color.FgCyan).Sprint(msg)
}

func Colorize(rich bool, colorFn func(string) string, msg string) string {
	if rich {
		return colorFn(msg)
	}
	return msg
}
