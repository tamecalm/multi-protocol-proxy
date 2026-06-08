package ui

import (
	"fmt"
	"os"
	"strings"
)

const DOCS_ROOT = "https://signal.org/docs"

func SupportsHyperlinks() bool {
	termProgram := os.Getenv("TERM_PROGRAM")
	term := os.Getenv("TERM")
	wtSession := os.Getenv("WT_SESSION") 

	if strings.Contains(termProgram, "iTerm") ||
		strings.Contains(termProgram, "WezTerm") ||
		strings.Contains(termProgram, "vscode") ||
		strings.Contains(termProgram, "Hyper") ||
		wtSession != "" {
		return true
	}

	if strings.Contains(term, "xterm-256color") {
		return true
	}

	return false
}


func FormatTerminalLink(label, url string) string {
	if !SupportsHyperlinks() {
		return fmt.Sprintf("%s (%s)", label, url)
	}


	return fmt.Sprintf("\x1b]8;;%s\x1b\\%s\x1b]8;;\x1b\\", url, label)
}

func FormatDocsLink(path, label string) string {
	url := path
	if !strings.HasPrefix(path, "http") {
		if !strings.HasPrefix(path, "/") {
			path = "/" + path
		}
		url = DOCS_ROOT + path
	}

	if label == "" {
		label = url
	}

	return FormatTerminalLink(label, url)
}

func FormatURLWithStyle(label, url string) string {
	link := FormatTerminalLink(label, url)
	if IsRich() {
		return Secondary(link)
	}
	return link
}

func FormatEmail(email string) string {
	return FormatTerminalLink(email, "mailto:"+email)
}
