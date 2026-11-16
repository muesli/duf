package main

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestFindMounts(t *testing.T) {
	base, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	data := filepath.Join(base, "data")
	nested := filepath.Join(data, "nested")
	sibling := filepath.Join(base, "database")
	for _, dir := range []string{nested, sibling} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	for _, dir := range []string{data, nested, sibling} {
		if err := os.WriteFile(filepath.Join(dir, "file"), nil, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	root := Mount{Device: "root", Mountpoint: filepath.VolumeName(base) + string(os.PathSeparator)}
	dataMount := Mount{Device: "data", Mountpoint: data}
	nestedMount := Mount{Device: "nested", Mountpoint: nested}
	devicePath := filepath.Join(base, "device")
	if err := os.WriteFile(devicePath, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	deviceMount := Mount{Device: devicePath, Mountpoint: filepath.Join(base, "device-mount")}
	mounts := []Mount{root, dataMount, nestedMount, deviceMount}

	for _, tc := range []struct {
		name string
		path string
		want []Mount
	}{
		{"root", root.Mountpoint, []Mount{root}},
		{"device path", devicePath, []Mount{deviceMount}},
		{"exact mount", data, []Mount{dataMount}},
		{"descendant", filepath.Join(data, "file"), []Mount{dataMount}},
		{"sibling directory", sibling, []Mount{root}},
		{"sibling file", filepath.Join(sibling, "file"), []Mount{root}},
		{"closest nested mount", filepath.Join(nested, "file"), []Mount{nestedMount}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := findMounts(mounts, tc.path)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("findMounts(%q) = %v, want %v", tc.path, got, tc.want)
			}
		})
	}

	t.Run("symlink to sibling", func(t *testing.T) {
		link := filepath.Join(base, "link")
		if err := os.Symlink(filepath.Join(sibling, "file"), link); err != nil {
			t.Skipf("symlinks unavailable: %v", err)
		}
		got, err := findMounts(mounts, link)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(got, []Mount{root}) {
			t.Errorf("findMounts(%q) = %v, want root mount", link, got)
		}
	})
}
