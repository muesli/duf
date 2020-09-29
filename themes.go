package main

import (
	"github.com/muesli/termenv"
)

type Theme struct {
	colorRed     termenv.Color
	colorYellow  termenv.Color
	colorGreen   termenv.Color
	colorBlue    termenv.Color
	colorGray    termenv.Color
	colorMagenta termenv.Color
	colorCyan    termenv.Color
}

func loadThemes(theme string) Theme {
	themes := make(map[string]Theme)

	themes["dark"] = Theme{
		colorRed:     term.Color("#E88388"),
		colorYellow:  term.Color("#DBAB79"),
		colorGreen:   term.Color("#A8CC8C"),
		colorBlue:    term.Color("#71BEF2"),
		colorGray:    term.Color("#B9BFCA"),
		colorMagenta: term.Color("#D290E4"),
		colorCyan:    term.Color("#66C2CD"),
	}

	themes["light"] = Theme{
		colorRed:     term.Color("#D70000"),
		colorYellow:  term.Color("#FFAF00"),
		colorGreen:   term.Color("#005F00"),
		colorBlue:    term.Color("#000087"),
		colorGray:    term.Color("#303030"),
		colorMagenta: term.Color("#AF00FF"),
		colorCyan:    term.Color("#0087FF"),
	}

	if _, ok := themes[theme]; !ok {
		return themes["dark"]
	}

	return themes[theme]
}
