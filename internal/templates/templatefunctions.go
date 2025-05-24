package templates

import (
	"strings"
	"tektmud/internal/language"
	"text/template"

	"github.com/mattn/go-runewidth"
)

var (
	functionsMap = template.FuncMap{
		"t":        language.T,
		"pad":      pad,
		"padLeft":  padLeft,
		"padright": padRight,
	}
)

// Usage:
//
//	{{ pad 10 }}
//	OUTPUT: 	"          "
//	{{ pad 10 "monkeys "."}}
//	OUTPUT: ".monkeys.."
func pad(width int, stringArgs ...string) string {
	var toPad string = ""
	var padWith string = " "

	if len(stringArgs) > 0 {
		toPad = stringArgs[0]
		if len(stringArgs) > 1 {
			padWith = stringArgs[1]
		}
	}

	toPadWidth := runewidth.StringWidth(toPad)

	if toPadWidth >= width {
		return toPad
	}

	paddingDifference := width - toPadWidth
	leftPad := paddingDifference >> 1 //Same as  num / 2

	//We want to pad extra chars to the right
	// So if width was 10, string of "monkeys" and padWith of "." - ".monkeys.."
	// Width: 10, "monkeysee" "." - "monkeysee."

	//this means diff was 1
	if leftPad < 1 {
		return toPad + padWith
	}
	return strings.Repeat(padWith, leftPad) + toPad + strings.Repeat(padWith, paddingDifference-leftPad)
}

// Usage:
//
//	{{ padLeft 10 }}
//	OUTPUT: 	"          "
//	{{ padLeft 10 "monkeys "."}}
//	OUTPUT: "...monkeys"
func padLeft(width int, stringArgs ...string) string {
	var toPad string = ""
	var padWith string = " "

	if len(stringArgs) > 0 {
		toPad = stringArgs[0]
		if len(stringArgs) > 1 {
			padWith = stringArgs[1]
		}
	}

	toPadWidth := runewidth.StringWidth(toPad)

	if toPadWidth >= width {
		return toPad
	}

	paddingDifference := width - toPadWidth

	return strings.Repeat(padWith, paddingDifference) + toPad
}

// Usage:
//
//	{{ padRight 10 }}
//	OUTPUT: 	"          "
//	{{ padRight 10 "monkeys "."}}
//	OUTPUT: "monkeys..."
func padRight(width int, stringArgs ...string) string {
	var toPad string = ""
	var padWith string = " "

	if len(stringArgs) > 0 {
		toPad = stringArgs[0]
		if len(stringArgs) > 1 {
			padWith = stringArgs[1]
		}
	}

	toPadWidth := runewidth.StringWidth(toPad)

	if toPadWidth >= width {
		return toPad
	}

	paddingDifference := width - toPadWidth

	return toPad + strings.Repeat(padWith, paddingDifference)
}
