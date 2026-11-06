package main

import (
	"io"
	"os"
	"strings"
	"testing"

	"github.com/jedib0t/go-pretty/v6/table"
)

func TestPrintTableDeviceCount(t *testing.T) {
	for _, style := range []table.Style{table.StyleDefault, table.StyleRounded} {
		for _, count := range []int{1, 2} {
			for _, hide := range []bool{false, true} {
				mounts := make([]Mount, count)
				for i := range mounts {
					mounts[i] = Mount{Mountpoint: "/a-long-mountpoint-for-device-count-tests", Device: "test", Fstype: "test", Total: 1024, Used: 256, Free: 768, Blocks: 1, BlockSize: 1024}
				}
				old := os.Stdout
				r, w, err := os.Pipe()
				if err != nil {
					t.Fatal(err)
				}
				os.Stdout = w
				printTable("local", mounts, TableOptions{Columns: []int{1}, SortBy: 1, Style: style, HideDeviceCount: hide})
				os.Stdout = old
				if err := w.Close(); err != nil {
					t.Fatal(err)
				}
				out, err := io.ReadAll(r)
				_ = r.Close()
				if err != nil {
					t.Fatal(err)
				}
				title := "local device"
				if count == 2 {
					title += "s"
				}
				if !hide {
					if count == 1 {
						title = "1 " + title
					} else {
						title = "2 " + title
					}
				}
				if !strings.Contains(string(out), title) {
					t.Errorf("count=%d hide=%v: missing title %q in %q", count, hide, title, out)
				}
				if hide && (strings.Contains(string(out), "1 local") || strings.Contains(string(out), "2 local")) {
					t.Errorf("hidden count in output: %q", out)
				}
			}
		}
	}
}
