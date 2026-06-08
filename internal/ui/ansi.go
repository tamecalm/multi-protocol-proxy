package ui

import (
	"regexp"
	"unicode/utf8"
)

var (
	ansiSGRPattern = regexp.MustCompile(`\x1b\[[0-9;]*m`)

	osc8Pattern = regexp.MustCompile(`\x1b\]8;;[^\x1b]*\x1b\\|\x1b\]8;;\x1b\\`)

	allAnsiPattern = regexp.MustCompile(`\x1b\[[0-9;]*m|\x1b\]8;;[^\x1b]*\x1b\\`)
)

func StripAnsi(input string) string {
	result := osc8Pattern.ReplaceAllString(input, "")
	result = ansiSGRPattern.ReplaceAllString(result, "")
	return result
}


func VisibleWidth(input string) int {
	stripped := StripAnsi(input)
	return utf8.RuneCountInString(stripped)
}


func TruncateVisible(input string, maxWidth int) string {
	stripped := StripAnsi(input)
	if utf8.RuneCountInString(stripped) <= maxWidth {
		return input
	}

	runes := []rune(stripped)
	if len(runes) > maxWidth-3 {
		return string(runes[:maxWidth-3]) + "..."
	}
	return string(runes[:maxWidth])
}

func PadRight(input string, width int) string {
	visible := VisibleWidth(input)
	if visible >= width {
		return input
	}
	padding := width - visible
	return input + spaces(padding)
}

func PadLeft(input string, width int) string {
	visible := VisibleWidth(input)
	if visible >= width {
		return input
	}
	padding := width - visible
	return spaces(padding) + input
}

func PadCenter(input string, width int) string {
	visible := VisibleWidth(input)
	if visible >= width {
		return input
	}
	padding := width - visible
	left := padding / 2
	right := padding - left
	return spaces(left) + input + spaces(right)
}

func spaces(n int) string {
	if n <= 0 {
		return ""
	}
	b := make([]byte, n)
	for i := range b {
		b[i] = ' '
	}
	return string(b)
}
