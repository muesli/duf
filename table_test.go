package main

import (
	"fmt"
	"io"
	"math"
	"os"
	"strings"
	"testing"
)

func TestFormatSizeFixedUnit(t *testing.T) {
	for _, tt := range []struct {
		size       uint64
		unit, want string
	}{
		{0, "G", "0.0G"}, {1 << 20, "G", "0.0G"}, {1 << 30, "G", "1.0G"},
		{3 << 29, "G", "1.5G"}, {1 << 40, "G", "1024.0G"},
		{1536, "K", "1.5K"}, {1536, "B", "1536B"},
		{math.MaxUint64, "B", "18446744073709551615B"},
		{1 << 60, "E", "1.0E"}, {math.MaxUint64, "E", "16.0E"},
	} {
		t.Run(fmt.Sprintf("%d/%s", tt.size, tt.unit), func(t *testing.T) {
			if got := formatSize(tt.size, tt.unit); got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
	for i, unit := range []string{"K", "M", "G", "T", "P", "E"} {
		t.Run(unit, func(t *testing.T) {
			n := uint64(1) << uint(10*(i+1))
			if got := formatSize(n, unit); got != "1.0"+unit {
				t.Errorf("got %q", got)
			}
		})
	}
}

func TestFormatSizeAutomaticUnchanged(t *testing.T) {
	for _, tt := range []struct {
		size uint64
		want string
	}{
		{0, "0B"}, {1023, "1023B"}, {1024, "1.0K"}, {1 << 20, "1.0M"},
		{1 << 30, "1.0G"}, {1 << 40, "1.0T"}, {1 << 50, "1.0P"}, {1 << 60, "1.0E"},
	} {
		if got := formatSize(tt.size, ""); got != tt.want {
			t.Errorf("size %d: got %q, want %q", tt.size, got, tt.want)
		}
	}
}

func TestParseSizeUnit(t *testing.T) {
	for _, unit := range []string{"", "B", "K", "M", "G", "T", "P", "E"} {
		got, err := parseSizeUnit(unit)
		if err != nil || got != unit {
			t.Errorf("%q: %q, %v", unit, got, err)
		}
	}
	if got, err := parseSizeUnit("g"); err != nil || got != "G" {
		t.Errorf("lowercase: %q, %v", got, err)
	}
	for _, unit := range []string{"GB", "GiB", "auto", "-1", " G"} {
		if _, err := parseSizeUnit(unit); err == nil {
			t.Errorf("accepted %q", unit)
		}
	}
}

func captureTableOutput(t *testing.T, f func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	original := os.Stdout
	os.Stdout = w
	defer func() { os.Stdout = original }()
	f()
	w.Close()
	data, err := io.ReadAll(r)
	r.Close()
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func TestFixedUnitTableAndNumericSorting(t *testing.T) {
	oldUnits, oldWidth := *units, *width
	defer func() { *units = oldUnits; *width = oldWidth }()
	*units = "M"
	*width = 120
	mounts := []Mount{
		{Mountpoint: "/large", Total: 10 << 30, Used: 5 << 30, Free: 5 << 30},
		{Mountpoint: "/small", Total: 2 << 30, Used: 1 << 30, Free: 1 << 30},
	}
	style, err := parseStyle("ascii")
	if err != nil {
		t.Fatal(err)
	}
	output := captureTableOutput(t, func() {
		printTable("local", mounts, TableOptions{Columns: []int{1, 2, 3, 4}, SortBy: 12, Style: style, StyleName: "ascii"})
	})
	for _, want := range []string{"10240.0M", "5120.0M", "2048.0M", "1024.0M"} {
		if !strings.Contains(output, want) {
			t.Errorf("missing %q in %s", want, output)
		}
	}
	if strings.Index(output, "/small") > strings.Index(output, "/large") {
		t.Errorf("sorting should use raw sizes: %s", output)
	}
	fixed := captureTableOutput(t, func() {
		if err := renderJSON(mounts); err != nil {
			t.Fatal(err)
		}
	})
	*units = ""
	automatic := captureTableOutput(t, func() {
		if err := renderJSON(mounts); err != nil {
			t.Fatal(err)
		}
	})
	if fixed != automatic {
		t.Fatal("fixed display units changed JSON")
	}
}
