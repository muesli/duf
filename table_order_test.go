package main

import (
	"bytes"
	"io"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/jedib0t/go-pretty/v6/table"
)

func TestTableColumnOrder(t *testing.T) {
	for _, tc := range []struct {
		name   string
		opts   TableOptions
		prefix []int
	}{
		{"default", TableOptions{Columns: []int{11, 1, 5}}, []int{1, 2, 3, 4, 5}},
		{"custom", TableOptions{Columns: []int{11, 1, 5}, CustomOrder: true}, []int{11, 1, 5}},
		{"duplicates", TableOptions{Columns: []int{11, 1, 11}, CustomOrder: true}, []int{11, 1}},
		{"empty", TableOptions{CustomOrder: true}, []int{1, 2, 3}},
		{"inodes", TableOptions{Columns: []int{9, 6, 1}, CustomOrder: true}, []int{9, 6, 1}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			order := tableColumnOrder(tc.opts)
			if !reflect.DeepEqual(order[:len(tc.prefix)], tc.prefix) {
				t.Fatalf("order=%v", order)
			}
			if len(order) != 19 {
				t.Fatalf("len=%d", len(order))
			}
			for ci := 1; ci <= 19; ci++ {
				if order[columnPosition(order, ci)-1] != ci {
					t.Fatalf("lost column %d", ci)
				}
			}
		})
	}
}

func captureTable(t *testing.T, mounts []Mount, opts TableOptions) string {
	t.Helper()
	savedOut, savedWidth, savedTheme := os.Stdout, *width, theme
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	*width = 120
	theme, _ = loadTheme("ansi")
	defer func() { os.Stdout = savedOut; *width = savedWidth; theme = savedTheme }()
	printTable("local", mounts, opts)
	w.Close()
	var b bytes.Buffer
	_, err = io.Copy(&b, r)
	r.Close()
	if err != nil {
		t.Fatal(err)
	}
	return b.String()
}

func TestCustomColumnRendering(t *testing.T) {
	mounts := []Mount{
		{Mountpoint: "/alpha", Device: "/dev/z", Fstype: "ext4", Total: 200, Used: 20, Free: 180, Inodes: 100, InodesUsed: 25, InodesFree: 75},
		{Mountpoint: "/beta", Device: "/dev/a", Fstype: "xfs", Total: 100, Used: 90, Free: 10, Inodes: 200, InodesUsed: 100, InodesFree: 100},
	}
	for _, style := range []struct {
		name  string
		value table.Style
	}{{"ascii", table.StyleDefault}, {"unicode", table.StyleRounded}} {
		t.Run(style.name, func(t *testing.T) {
			opts := TableOptions{Columns: []int{11, 1, 5}, Style: style.value, StyleName: style.name, SortBy: 12}
			original := captureTable(t, mounts, opts)
			opts.CustomOrder = true
			custom := captureTable(t, mounts, opts)
			if strings.Index(custom, "FILESYSTEM") > strings.Index(custom, "MOUNTED ON") {
				t.Fatal(custom)
			}
			if strings.Index(original, "MOUNTED ON") > strings.Index(original, "FILESYSTEM") {
				t.Fatal(original)
			}
			if strings.Index(custom, "/beta") > strings.Index(custom, "/alpha") {
				t.Fatal("numeric hidden-column sort failed", custom)
			}
			if strings.Contains(custom, "SIZE") || strings.Contains(custom, "IUSED") {
				t.Fatal("hidden column visible", custom)
			}
			if !strings.Contains(custom, "90.0%") || !strings.Contains(custom, "10.0%") {
				t.Fatal("usage transformer lost", custom)
			}
			opts.Reverse = true
			reversed := captureTable(t, mounts, opts)
			if strings.Index(reversed, "/alpha") > strings.Index(reversed, "/beta") {
				t.Fatal("reverse sort failed", reversed)
			}
			opts.Columns = []int{9, 6, 1}
			opts.SortBy = 16
			opts.Reverse = false
			inode := captureTable(t, mounts, opts)
			if strings.Index(inode, "IUSE%") > strings.Index(inode, "INODES") || !strings.Contains(inode, "25.0%") {
				t.Fatal(inode)
			}
		})
	}
}

func TestCustomColumnConfigs(t *testing.T) {
	for first := 1; first <= 11; first++ {
		opts := TableOptions{Columns: []int{first, 11, 1}, CustomOrder: true}
		order := tableColumnOrder(opts)
		tab := table.NewWriter()
		appendHeaders(tab, order)
		setColumnConfigs(tab, map[int]int{}, map[int]int{}, opts, func(interface{}) string { return "usage" })
		// All original sort columns, including numeric helpers, remain addressable.
		for _, column := range columns {
			pos := columnPosition(order, column.SortIndex)
			if pos < 1 || pos > 19 || order[pos-1] != column.SortIndex {
				t.Fatalf("column %s: %v", column.ID, order)
			}
		}
	}
}
