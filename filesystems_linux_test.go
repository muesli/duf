//go:build linux
// +build linux

package main

import (
	"github.com/jedib0t/go-pretty/v6/table"
	"golang.org/x/sys/unix"
	"io"
	"os"
	"strings"
	"testing"
)

func TestWekaFSNetworkClassification(t *testing.T) {
	for _, tc := range []struct {
		name, fstype string
		magic        int64
		want         string
	}{
		{"weka unknown magic", "wekafs", 0, networkDevice},
		{"weka native mount", "wekafs", 123456, networkDevice},
		{"other unknown", "other", 0, localDevice},
		{"similar name", "notwekafs", 0, localDevice},
		{"nfs unchanged", "nfs", NFS_SUPER_MAGIC, networkDevice},
		{"ext4 unchanged", "ext4", EXT4_SUPER_MAGIC, localDevice},
		{"tmpfs unchanged", "tmpfs", TMPFS_MAGIC, specialDevice},
		{"fuse unchanged", "fuse.sshfs", FUSE_SUPER_MAGIC, fuseDevice},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := Mount{Fstype: tc.fstype, Metadata: unix.Statfs_t{Type: tc.magic}}
			if got := deviceType(m); got != tc.want {
				t.Fatalf("deviceType=%q want%q", got, tc.want)
			}
		})
	}
}

func TestWekaFSNetworkFiltering(t *testing.T) {
	oldWidth, oldTheme := *width, theme
	*width = 100
	theme, _ = loadTheme("ansi")
	defer func() { *width = oldWidth; theme = oldTheme }()
	m := Mount{Device: "backend/fs", Mountpoint: "/mnt/weka", Fstype: "wekafs", Metadata: unix.Statfs_t{}, Total: 100, Used: 40, Free: 60, Blocks: 100, BlockSize: 4096}
	m.DeviceType = deviceType(m)
	for _, tc := range []struct {
		name    string
		filters FilterOptions
		shown   bool
	}{
		{"only network", FilterOptions{OnlyDevices: map[string]struct{}{networkDevice: {}}}, true},
		{"only local", FilterOptions{OnlyDevices: map[string]struct{}{localDevice: {}}}, false},
		{"hide network", FilterOptions{HiddenDevices: map[string]struct{}{networkDevice: {}}}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			original := os.Stdout
			r, w, err := os.Pipe()
			if err != nil {
				t.Fatal(err)
			}
			os.Stdout = w
			renderTables([]Mount{m}, tc.filters, TableOptions{Columns: []int{1, 11}, SortBy: 1, Style: table.StyleDefault, StyleName: "ascii"})
			w.Close()
			os.Stdout = original
			b, err := io.ReadAll(r)
			r.Close()
			if err != nil {
				t.Fatal(err)
			}
			if got := strings.Contains(string(b), "/mnt/weka"); got != tc.shown {
				t.Fatalf("shown=%v want%v: %s", got, tc.shown, b)
			}
			if tc.shown && !strings.Contains(string(b), "network") {
				t.Fatal(string(b))
			}
		})
	}
}
