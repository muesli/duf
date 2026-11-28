//go:build darwin
// +build darwin

package main

import (
	"testing"

	"golang.org/x/sys/unix"
)

func TestDarwinDeviceType(t *testing.T) {
	var tt = []struct {
		fstype   string
		local    bool
		expected string
	}{
		// The bug: an nfs mount carries no MNT_LOCAL, and the fstype alone was
		// never consulted, so every share was reported under local devices.
		{fstype: "nfs", local: false, expected: networkDevice},
		{fstype: "nfs4", local: false, expected: networkDevice},
		{fstype: "smbfs", local: false, expected: networkDevice},
		{fstype: "afpfs", local: false, expected: networkDevice},
		// autofs lacks the flag too: the kernel automounter serves it, but the
		// map it triggers can point at a local disk.
		{fstype: "autofs", local: false, expected: localDevice},
		{fstype: "apfs", local: true, expected: localDevice},
		{fstype: "hfs", local: true, expected: localDevice},
		{fstype: "devfs", local: true, expected: specialDevice},
	}

	for _, v := range tt {
		var flags uint32
		if v.local {
			flags |= unix.MNT_LOCAL
		}
		stat := unix.Statfs_t{Flags: flags}
		m := Mount{
			Fstype:   v.fstype,
			Type:     v.fstype,
			Blocks:   1,
			Metadata: stat,
		}

		if got := deviceType(m); got != v.expected {
			t.Errorf("%s (MNT_LOCAL=%t): expected %s, got %s", v.fstype, v.local, v.expected, got)
		}
	}
}
