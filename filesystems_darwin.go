//go:build darwin
// +build darwin

package main

import "golang.org/x/sys/unix"

func isFuseFs(m Mount) bool {
	//FIXME: implement
	return false
}

// MNT_LOCAL is the kernel's answer: darwin sets it on every mount a local
// disk backs and clears it for the network filesystem served by nfs, smbfs or
// afpfs, whatever f_fstypename says. autofs is the one local filesystem that
// lacks the bit (the automounter serves it through a kernel thread), so it is
// exempted rather than classified as a network device.
func isNetworkFs(m Mount) bool {
	if m.Fstype == "autofs" {
		return false
	}

	return m.Stat().Flags&unix.MNT_LOCAL == 0
}

func isSpecialFs(m Mount) bool {
	return m.Fstype == "devfs"
}

func isHiddenFs(m Mount) bool {
	return false
}
