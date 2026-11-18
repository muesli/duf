//go:build linux
// +build linux

package main

import (
	"testing"

	"golang.org/x/sys/unix"
)

// zero-block fuse mount, e.g. a document portal: hidden by default,
// visible on explicit request.
func fusePortalMount() Mount {
	return Mount{
		Device:     "portal",
		Fstype:     "fuse.portal",
		Mountpoint: "/run/user/1001/doc",
		Opts:       "rw",
		Blocks:     0,
		BlockSize:  0,
		Metadata:   unix.Statfs_t{Type: FUSE_SUPER_MAGIC},
	}
}

func localMount() Mount {
	return Mount{
		Device:     "/dev/sda1",
		Fstype:     "ext4",
		Mountpoint: "/",
		Opts:       "rw",
		Blocks:     100,
		BlockSize:  4096,
		Metadata:   unix.Statfs_t{},
	}
}

func onlyFilter(names ...string) FilterOptions {
	m := map[string]struct{}{}
	for _, n := range names {
		m[n] = struct{}{}
	}
	return FilterOptions{OnlyDevices: m}
}

func TestFusePortalClassifiedFuse(t *testing.T) {
	if got := deviceType(fusePortalMount()); got != fuseDevice {
		t.Fatalf("deviceType = %q, want %q", got, fuseDevice)
	}
}

func TestZeroBlockFuseSkippedByDefault(t *testing.T) {
	if !shouldSkipMount(fusePortalMount(), FilterOptions{}) {
		t.Fatal("zero-block fuse mount should stay hidden without filters")
	}
}

func TestZeroBlockFuseShownWhenRequested(t *testing.T) {
	if shouldSkipMount(fusePortalMount(), onlyFilter(fuseDevice)) {
		t.Fatal("-only fuse should keep zero-block fuse mounts")
	}
}

func TestZeroBlockFuseSkippedWhenOtherGroupRequested(t *testing.T) {
	if !shouldSkipMount(fusePortalMount(), onlyFilter(localDevice)) {
		t.Fatal("zero-block fuse mount should stay hidden under -only local")
	}
}

func TestFilterMountsGroupMembership(t *testing.T) {
	m := []Mount{localMount()}

	if got := filterMounts(m, onlyFilter(specialDevice)); len(got) != 0 {
		t.Fatalf("-only special kept %d local mounts", len(got))
	}

	hide := FilterOptions{HiddenDevices: map[string]struct{}{localDevice: {}}}
	if got := filterMounts(m, hide); len(got) != 0 {
		t.Fatalf("-hide local kept %d local mounts", len(got))
	}

	both := FilterOptions{
		OnlyDevices:   map[string]struct{}{localDevice: {}},
		HiddenDevices: map[string]struct{}{localDevice: {}},
	}
	if got := filterMounts(m, both); len(got) != 1 {
		t.Fatalf("-only wins over -hide: kept %d, want 1", len(got))
	}
}
