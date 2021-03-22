package main

import "github.com/mattn/go-runewidth"

func defaultStyleName() string {
	/*
		Due to a bug in github.com/mattn/go-runewidth v0.0.9, the width of unicode rune(such as '╭') could not be correctly
		calculated.	Degrade to ascii to prevent broken table structure. Remove this once the bug is fixed.
	*/
	if runewidth.RuneWidth('╭') > 1 {
		return "ascii"
	}

	return "unicode"
}
