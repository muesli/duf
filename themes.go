package main

import (
	"fmt"

	"github.com/muesli/termenv"
)

// Theme defines a color theme used for printing tables.
type Theme struct {
	colorRed     termenv.Color
	colorYellow  termenv.Color
	colorGreen   termenv.Color
	colorBlue    termenv.Color
	colorGray    termenv.Color
	colorMagenta termenv.Color
	colorCyan    termenv.Color
}

func defaultThemeName() string {
	if !termenv.HasDarkBackground() {
		return "light"
	}
	return "dark"
}

func loadTheme(theme string) (Theme, error) {
	themes := make(map[string]Theme)

	themes["dark"] = Theme{
		colorRed:     env.Color("#E88388"),
		colorYellow:  env.Color("#DBAB79"),
		colorGreen:   env.Color("#A8CC8C"),
		colorBlue:    env.Color("#71BEF2"),
		colorGray:    env.Color("#B9BFCA"),
		colorMagenta: env.Color("#D290E4"),
		colorCyan:    env.Color("#66C2CD"),
	}

	themes["light"] = Theme{
		colorRed:     env.Color("#D70000"),
		colorYellow:  env.Color("#FFAF00"),
		colorGreen:   env.Color("#005F00"),
		colorBlue:    env.Color("#000087"),
		colorGray:    env.Color("#303030"),
		colorMagenta: env.Color("#AF00FF"),
		colorCyan:    env.Color("#0087FF"),
	}

	themes["ansi"] = Theme{
		colorRed:     env.Color("9"),
		colorYellow:  env.Color("11"),
		colorGreen:   env.Color("10"),
		colorBlue:    env.Color("12"),
		colorGray:    env.Color("7"),
		colorMagenta: env.Color("13"),
		colorCyan:    env.Color("8"),
	}

	if _, ok := themes[theme]; !ok {
		return Theme{}, fmt.Errorf("unknown theme: %s", theme)
	}

	return themes[theme], nil
}
