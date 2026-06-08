package ui

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
)

var asciiBanner = []string{
	"██████╗░██████╗░░█████╗░██╗░░██╗██╗░░░██╗",
	"██╔══██╗██╔══██╗██╔══██╗╚██╗██╔╝╚██╗░██╔╝",
	"██████╔╝██████╔╝██║░░██║░╚███╔╝░░╚████╔╝░",
	"██╔═══╝░██╔══██╗██║░░██║░██╔██╗░░░╚██╔╝░░",
	"██║░░░░░██║░░██║╚█████╔╝██╔╝╚██╗░░░██║░░░",
	"╚═╝░░░░░╚═╝░░╚═╝░╚════╝░╚═╝░░╚═╝░░░╚═╝░░░",
}

var simpleBanner = []string{
	"██████╗ ██████╗  ██████╗ ██╗  ██╗██╗   ██╗",
	"██╔══██╗██╔══██╗██╔═══██╗╚██╗██╔╝╚██╗ ██╔╝",
	"██████╔╝██████╔╝██║   ██║ ╚███╔╝  ╚████╔╝ ",
	"██╔═══╝ ██╔══██╗██║   ██║ ██╔██╗   ╚██╔╝  ",
	"██║     ██║  ██║╚██████╔╝██╔╝╚██╗   ██║   ",
	"╚═╝     ╚═╝  ╚═╝ ╚═════╝ ╚═╝  ╚═╝   ╚═╝   ",
}

var bannerEmitted = false

func FormatBannerArt() string {
	rich := IsRich()
	if !rich {
		return strings.Join(simpleBanner, "\n")
	}

	accent := color.New(color.FgHiRed, color.Bold)
	accentDim := color.New(color.FgRed)

	var lines []string
	for _, line := range simpleBanner {
		var coloredLine strings.Builder
		for _, ch := range line {
			switch ch {
			case '█', '╗', '╔', '╚', '╝', '║':
				coloredLine.WriteString(accent.Sprint(string(ch)))
			case '░', '═', '╣', '╠':
				coloredLine.WriteString(accentDim.Sprint(string(ch)))
			default:
				coloredLine.WriteString(Muted(string(ch)))
			}
		}
		lines = append(lines, coloredLine.String())
	}
	return strings.Join(lines, "\n")
}

func FormatBannerLine(version, tagline string) string {
	rich := IsRich()
	title := "◆ MULTI-PROTOCOL PROXY"

	if rich {
		return fmt.Sprintf("%s %s %s %s",
			Heading(title),
			Info(version),
			Muted("—"),
			AccentDim(tagline))
	}
	return fmt.Sprintf("%s %s — %s", title, version, tagline)
}

func EmitBanner(version, tagline string) {
	if bannerEmitted {
		return
	}
	if !isTTY() {
		return
	}
	for _, arg := range os.Args {
		if arg == "--json" || arg == "--version" || arg == "-v" {
			return
		}
	}

	fmt.Println()
	fmt.Println(FormatBannerArt())
	fmt.Println()
	fmt.Println(FormatBannerLine(version, tagline))
	fmt.Println()
	bannerEmitted = true
}

func EmitSimpleBanner(version, tagline string) {
	if bannerEmitted {
		return
	}
	if !isTTY() {
		return
	}

	fmt.Println()

	badge := color.New(color.BgMagenta, color.FgWhite, color.Bold).Sprint(" ◆ PROXY ")
	ver := Muted(version)

	topBorder := Muted(boxTopLeft + strings.Repeat(boxHorizontal, 60) + boxTopRight)
	fmt.Println(topBorder)

	titleLine := fmt.Sprintf("%s  %s %s  %s",
		Muted(boxVertical),
		badge,
		ver,
		Muted(strings.Repeat(" ", 38)+boxVertical))
	fmt.Println(titleLine)

	subtitle := Subtle(tagline)
	subtitleLine := fmt.Sprintf("%s  %s%s",
		Muted(boxVertical),
		subtitle,
		Muted(strings.Repeat(" ", 60-2-len(tagline))+boxVertical))
	fmt.Println(subtitleLine)

	bottomBorder := Muted(boxBottomLeft + strings.Repeat(boxHorizontal, 60) + boxBottomRight)
	fmt.Println(bottomBorder)
	fmt.Println()

	bannerEmitted = true
}

func isTTY() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}

func ResetBanner() {
	bannerEmitted = false
}
