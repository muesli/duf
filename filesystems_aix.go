//go:build aix
// +build aix

package main

// isFuseFs reports whether the mount is a FUSE filesystem.
func isFuseFs(m Mount) bool {
	return false // AIX doesn't have FUSE
}

// isNetworkFs reports whether the mount is a network filesystem.
func isNetworkFs(m Mount) bool {
	switch m.Fstype {
	case "nfs", "nfs3", "nfs4", "cifs", "smbfs", "afs":
		return true
	}
	return false
}

// isSpecialFs reports whether the mount is a special filesystem.
func isSpecialFs(m Mount) bool {
	switch m.Fstype {
	case "procfs", "namefs", "sfs", "cdrom":
		return true
	}
	return false
}

// isHiddenFs reports whether the mount is a hidden filesystem.
func isHiddenFs(m Mount) bool {
	return false
}
