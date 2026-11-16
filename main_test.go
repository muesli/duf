package main

import (
	"reflect"
	"testing"
)

func TestParseCommaSeparatedValues(t *testing.T) {
	var tt = []struct {
		input    string
		expected map[string]struct{}
	}{
		// Empty values are skipped
		{input: "", expected: map[string]struct{}{}},
		{input: " , ,", expected: map[string]struct{}{}},

		// Values are trimmed and deduplicated
		{input: "local,network", expected: map[string]struct{}{"local": {}, "network": {}}},
		{input: " local , network ", expected: map[string]struct{}{"local": {}, "network": {}}},
		{input: "/home,/home", expected: map[string]struct{}{"/home": {}}},

		// The case of each value is kept
		{input: "/mnt/PrimLine", expected: map[string]struct{}{"/mnt/PrimLine": {}}},
		{input: "/Volumes/*,/volumes/*", expected: map[string]struct{}{"/Volumes/*": {}, "/volumes/*": {}}},
	}

	for _, tc := range tt {
		if got := parseCommaSeparatedValues(tc.input); !reflect.DeepEqual(got, tc.expected) {
			t.Errorf("parseCommaSeparatedValues(%q) = %v, expected %v",
				tc.input, got, tc.expected)
		}
	}
}

func TestFindInKey(t *testing.T) {
	var tt = []struct {
		patterns []string
		key      string
		expected bool
	}{
		// No pattern
		{patterns: nil, key: "/home", expected: false},

		// Exact match
		{patterns: []string{"/home"}, key: "/home", expected: true},
		{patterns: []string{"/home"}, key: "/home/user", expected: false},
		{patterns: []string{"/boot", "/home"}, key: "/home", expected: true},

		// * matches any number of characters
		{patterns: []string{"*"}, key: "/anything", expected: true},
		{patterns: []string{"/sys/*"}, key: "/sys/kernel/debug", expected: true},
		{patterns: []string{"/sys/*"}, key: "/sys/", expected: true},
		{patterns: []string{"/sys/*"}, key: "/system", expected: false},
		{patterns: []string{"*/docker/*"}, key: "/var/lib/docker/overlay2", expected: true},

		// ? matches zero or one character
		{patterns: []string{"/dev/sd?"}, key: "/dev/sda", expected: true},
		{patterns: []string{"/dev/sd?"}, key: "/dev/sd", expected: true},
		{patterns: []string{"/dev/sd?"}, key: "/dev/sdab", expected: false},

		// . matches exactly one character, a dot included
		{patterns: []string{"/dev/sd."}, key: "/dev/sda", expected: true},
		{patterns: []string{"/dev/sd."}, key: "/dev/sd", expected: false},
		{patterns: []string{"/dev/sd."}, key: "/dev/sdab", expected: false},
		{patterns: []string{"/mnt/a.b"}, key: "/mnt/a.b", expected: true},
		{patterns: []string{"/mnt/a.b"}, key: "/mnt/axb", expected: true},

		// Operators apply to characters, not bytes
		{patterns: []string{"/mnt/caf."}, key: "/mnt/café", expected: true},
		{patterns: []string{"/mnt/caf?"}, key: "/mnt/café", expected: true},

		// Matching is case-sensitive
		{patterns: []string{"/mnt/prim*"}, key: "/mnt/PrimLine", expected: false},
		{patterns: []string{"/mnt/Prim*"}, key: "/mnt/PrimLine", expected: true},
	}

	for _, tc := range tt {
		km := make(map[string]struct{}, len(tc.patterns))
		for _, p := range tc.patterns {
			km[p] = struct{}{}
		}

		if got := findInKey(tc.key, km); got != tc.expected {
			t.Errorf("findInKey(%q, %q) = %v, expected %v",
				tc.key, tc.patterns, got, tc.expected)
		}
	}
}
