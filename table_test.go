package main

import (
	"math"
	"strings"
	"testing"

	flag "github.com/spf13/pflag"
)

func TestSizeToString(t *testing.T) {
	var tt = []struct {
		input    uint64
		si       bool
		expected string
	}{
		// IEC prefixes, powers of 1024
		{input: 0, expected: "0B"},
		{input: 1000, expected: "1000B"},
		{input: 1023, expected: "1023B"},
		{input: 1024, expected: "1.0KiB"},
		{input: 1536, expected: "1.5KiB"},
		{input: 1048524, expected: "1023.9KiB"},
		{input: 1048525, expected: "1.0MiB"},
		{input: 1<<20 - 1, expected: "1.0MiB"},
		{input: 1 << 20, expected: "1.0MiB"},
		{input: 1<<30 - 1, expected: "1.0GiB"},
		{input: 1 << 30, expected: "1.0GiB"},
		{input: 15421 << 20, expected: "15.1GiB"},
		{input: 1 << 40, expected: "1.0TiB"},
		{input: 1 << 50, expected: "1.0PiB"},
		{input: 1 << 60, expected: "1.0EiB"},
		{input: math.MaxUint64, expected: "16.0EiB"},

		// SI prefixes, powers of 1000
		{input: 0, si: true, expected: "0B"},
		{input: 999, si: true, expected: "999B"},
		{input: 1000, si: true, expected: "1.0kB"},
		{input: 1023, si: true, expected: "1.0kB"},
		{input: 1024, si: true, expected: "1.0kB"},
		{input: 1500, si: true, expected: "1.5kB"},
		{input: 999949, si: true, expected: "999.9kB"},
		{input: 999950, si: true, expected: "1.0MB"},
		{input: 1e6, si: true, expected: "1.0MB"},
		{input: 1e9 - 1, si: true, expected: "1.0GB"},
		{input: 1e9, si: true, expected: "1.0GB"},
		{input: 15421 << 20, si: true, expected: "16.2GB"},
		{input: 1e12 - 1, si: true, expected: "1.0TB"},
		{input: 1e12, si: true, expected: "1.0TB"},
		{input: 1e15 - 1, si: true, expected: "1.0PB"},
		{input: 1e15, si: true, expected: "1.0PB"},
		{input: 1e18, si: true, expected: "1.0EB"},
		{input: math.MaxUint64, si: true, expected: "18.4EB"},
	}

	for _, tc := range tt {
		actual := sizeToString(tc.input, tc.si)
		if actual != tc.expected {
			t.Errorf("sizeToString(%d, %t) == %q, expected %q", tc.input, tc.si, actual, tc.expected)
		}
	}
}

func TestStringToSize(t *testing.T) {
	var tt = []struct {
		input    string
		si       bool
		expected uint64
	}{
		// IEC prefixes, powers of 1024
		{input: "0", expected: 0},
		{input: "1000", expected: 1000},
		{input: "1024", expected: 1024},
		{input: "1k", expected: 1 << 10},
		{input: "1K", expected: 1 << 10},
		{input: "10K", expected: 10 << 10},
		{input: "1M", expected: 1 << 20},
		{input: "1G", expected: 1 << 30},
		{input: "10G", expected: 10 << 30},
		{input: "1T", expected: 1 << 40},
		{input: "1P", expected: 1 << 50},
		{input: "2E", expected: 2 << 60},
		{input: "15E", expected: 15 << 60},
		{input: "18446744073709551615", expected: math.MaxUint64},

		// SI prefixes, powers of 1000
		{input: "0", si: true, expected: 0},
		{input: "1000", si: true, expected: 1000},
		{input: "1024", si: true, expected: 1024},
		{input: "1k", si: true, expected: 1e3},
		{input: "1K", si: true, expected: 1e3},
		{input: "10K", si: true, expected: 10e3},
		{input: "1M", si: true, expected: 1e6},
		{input: "1G", si: true, expected: 1e9},
		{input: "10G", si: true, expected: 10e9},
		{input: "1T", si: true, expected: 1e12},
		{input: "1P", si: true, expected: 1e15},
		{input: "2E", si: true, expected: 2e18},
		{input: "18E", si: true, expected: 18e18},
		{input: "18446744073709551615", si: true, expected: math.MaxUint64},
	}

	for _, tc := range tt {
		actual, err := stringToSize(tc.input, tc.si)
		if err != nil {
			t.Errorf("stringToSize(%q, %t) returned error: %v", tc.input, tc.si, err)
			continue
		}
		if actual != tc.expected {
			t.Errorf("stringToSize(%q, %t) == %d, expected %d", tc.input, tc.si, actual, tc.expected)
		}
	}
}

func TestStringToSizeInvalid(t *testing.T) {
	var tt = []string{
		"",
		"G",
		"-1G",
		"1.5G",
		"1e3",
		" 10G",
		"10G ",
		"10 G",
		"10g",
		"10m",
		"10X",
		"1Z",
		"1Y",
		"10Gi",
		"10GB",
		"10GiB",
		"10G,1G",
		"18446744073709551616",
		"19E",
		"18447P",
		"99999999999999999G",
		"18446744073709551615K",
	}

	for _, input := range tt {
		for _, si := range []bool{false, true} {
			actual, err := stringToSize(input, si)
			if err == nil {
				t.Errorf("stringToSize(%q, %t) == %d, expected an error", input, si, actual)
			}
		}
	}
}

func TestStringToSizeOverflowIEC(t *testing.T) {
	// Valid with SI prefixes, too large with IEC ones.
	var tt = []string{
		"16E",
		"17E",
		"18E",
		"16384P",
	}

	for _, input := range tt {
		actual, err := stringToSize(input, false)
		if err == nil {
			t.Errorf("stringToSize(%q, false) == %d, expected an error", input, actual)
		}
	}
}

func TestDefaultAvailThreshold(t *testing.T) {
	var tt = []struct {
		si       bool
		expected [2]uint64
	}{
		{si: false, expected: [2]uint64{10 << 30, 1 << 30}},
		{si: true, expected: [2]uint64{10e9, 1e9}},
	}

	def := flag.CommandLine.Lookup("avail-threshold").DefValue
	thresholds := strings.Split(def, ",")
	if len(thresholds) != 2 {
		t.Fatalf("default avail-threshold %q does not have two values", def)
	}

	for _, tc := range tt {
		for i, threshold := range thresholds {
			actual, err := stringToSize(threshold, tc.si)
			if err != nil {
				t.Errorf("stringToSize(%q, %t) returned error: %v", threshold, tc.si, err)
				continue
			}
			if actual != tc.expected[i] {
				t.Errorf("stringToSize(%q, %t) == %d, expected %d", threshold, tc.si, actual, tc.expected[i])
			}
		}
	}
}
