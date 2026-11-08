package main

import (
	"reflect"
	"strings"
	"testing"

	"github.com/jedib0t/go-pretty/v6/table"
)

func TestMountMetadataControls(t *testing.T) {
	for _, value := range []string{"name\x1b[31mred", "name\n\r\t\x00\x07\x7f", "name\u009b31m\u009dtitle\u009c", "/mnt/日本語-é"} {
		t.Run(value, func(t *testing.T) {
			mounts := []Mount{{Mountpoint: value, Fstype: value, Device: value}}
			original := append([]Mount(nil), mounts...)
			tab := table.NewWriter()
			appendRows(tab, mounts)
			output := tab.Render()
			expected := strings.Map(func(r rune) rune {
				if r < 0x20 || r >= 0x7f && r <= 0x9f {
					return '\ufffd'
				}
				return r
			}, value)
			if strings.Count(output, expected) != 3 {
				t.Errorf("metadata not safely rendered in all three columns: %q", output)
			}
			if !reflect.DeepEqual(mounts, original) {
				t.Fatal("presentation changed original mount metadata")
			}
			widths := computeMaxContentWidths(mounts, TableOptions{Columns: []int{1, 10, 11}})
			safeWidths := computeMaxContentWidths([]Mount{{Mountpoint: expected, Fstype: expected, Device: expected}}, TableOptions{Columns: []int{1, 10, 11}})
			if !reflect.DeepEqual(widths, safeWidths) {
				t.Errorf("widths %v do not match displayed widths %v", widths, safeWidths)
			}
		})
	}
}
