package ui

import (
	"math/rand"
	"strings"
	"time"
)

const defaultTagline = "Trusted Proxy Service"

var taglines = []string{
	"Trusted Proxy Service",
	"Your Packets, Relayed Securely",
	"Privacy-first proxy routing",
	"Connecting you securely, anywhere",
	"Secure relay for the modern era",
	"Routing trust, one packet at a time",
	"Making censorship obsolete since 2024",
	"Where privacy meets performance",
	"Silent guardian of your privacy",
	"Tunneling through barriers",
}

var holidayTaglines = map[string][]taglineRule{
	"christmas": {
		{month: 12, day: 25, tagline: "🎄 Ho ho ho—relaying holiday cheer!"},
		{month: 12, day: 24, tagline: "🎄 Santa's favorite proxy service"},
	},
	"halloween": {
		{month: 10, day: 31, tagline: "🎃 Boo! Your packets are haunted"},
		{month: 10, day: 30, tagline: "🎃 Spooky secure connections"},
	},
	"valentine": {
		{month: 2, day: 14, tagline: "💘 Sending love through encrypted channels"},
	},
	"newyear": {
		{month: 1, day: 1, tagline: "🎉 Happy New Year! Fresh connections await"},
	},
}

type taglineRule struct {
	month   int
	day     int
	tagline string
}

func PickTagline() string {
	now := time.Now()
	month := int(now.Month())
	day := now.Day()

	for _, rules := range holidayTaglines {
		for _, rule := range rules {
			if rule.month == month && rule.day == day {
				return rule.tagline
			}
		}
	}

	if len(taglines) == 0 {
		return defaultTagline
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return taglines[r.Intn(len(taglines))]
}

func GetAllTaglines() []string {
	return append([]string{}, taglines...)
}

func FormatTagline(tagline string) string {
	if !IsRich() {
		return tagline
	}
	if strings.HasPrefix(tagline, "🎄") ||
		strings.HasPrefix(tagline, "🎃") ||
		strings.HasPrefix(tagline, "💘") ||
		strings.HasPrefix(tagline, "🎉") {
		return tagline 
	}
	return AccentDim(tagline)
}
