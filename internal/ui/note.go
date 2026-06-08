package ui

import (
	"fmt"
	"os"
	"strings"
)

func Note(message string, title string) {
	wrapped := WrapNoteMessage(message, 80)
	lines := strings.Split(wrapped, "\n")

	fmt.Println()

	maxWidth := 0
	for _, line := range lines {
		w := VisibleWidth(line)
		if w > maxWidth {
			maxWidth = w
		}
	}
	boxWidth := maxWidth + 4 

	if title != "" {
		styledTitle := title
		if IsRich() {
			styledTitle = Heading(title)
		}
		top := fmt.Sprintf("%s%s %s %s%s",
			Muted(boxTopLeft),
			Muted(strings.Repeat(boxHorizontal, 2)),
			styledTitle,
			Muted(strings.Repeat(boxHorizontal, boxWidth-4-VisibleWidth(title))),
			Muted(boxTopRight))
		fmt.Println(top)
	} else {
		fmt.Println(Muted(boxTopLeft + strings.Repeat(boxHorizontal, boxWidth) + boxTopRight))
	}

	for _, line := range lines {
		padding := boxWidth - VisibleWidth(line) - 2
		if padding < 0 {
			padding = 0
		}
		fmt.Printf("%s %s%s %s\n",
			Muted(boxVertical),
			line,
			spaces(padding),
			Muted(boxVertical))
	}

	fmt.Println(Muted(boxBottomLeft + strings.Repeat(boxHorizontal, boxWidth) + boxBottomRight))
	fmt.Println()
}

func WrapNoteMessage(message string, maxWidth int) string {
	columns := 80
	if term, ok := os.LookupEnv("COLUMNS"); ok {
		if n := parseIntOr(term, 80); n > 0 {
			columns = n
		}
	}

	width := columns - 10
	if width > maxWidth {
		width = maxWidth
	}
	if width < 40 {
		width = 40
	}

	inputLines := strings.Split(message, "\n")
	var outputLines []string

	for _, line := range inputLines {
		wrapped := wrapLine(line, width)
		outputLines = append(outputLines, wrapped...)
	}

	return strings.Join(outputLines, "\n")
}

func wrapLine(line string, maxWidth int) []string {
	if strings.TrimSpace(line) == "" {
		return []string{line}
	}

	words := strings.Fields(line)
	if len(words) == 0 {
		return []string{""}
	}

	var lines []string
	current := ""

	for _, word := range words {
		candidate := current
		if current != "" {
			candidate += " "
		}
		candidate += word

		if VisibleWidth(candidate) <= maxWidth {
			current = candidate
		} else {
			if current != "" {
				lines = append(lines, current)
			}
			current = word
		}
	}

	if current != "" {
		lines = append(lines, current)
	}

	return lines
}

func parseIntOr(s string, def int) int {
	var n int
	_, err := fmt.Sscanf(s, "%d", &n)
	if err != nil {
		return def
	}
	return n
}

func InfoNote(message string) {
	Note(message, "ℹ Info")
}

func WarningNote(message string) {
	Note(message, "⚠ Warning")
}

func ErrorNote(message string) {
	Note(message, "✗ Error")
}

func SuccessNote(message string) {
	Note(message, "✓ Success")
}
