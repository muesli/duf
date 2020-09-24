package main

func deviceType(m Mount) string {
	if isNetworkFs(m) {
		return "network"
	}
	if isSpecialFs(m) {
		return "special"
	}
	if isFuseFs(m) {
		return "fuse"
	}

	return "local"
}

// remote: [ "nfs", "smbfs", "cifs", "ncpfs", "afs", "coda", "ftpfs", "mfs", "sshfs", "fuse.sshfs", "nfs4" ]
// special: [ "tmpfs", "devpts", "devtmpfs", "proc", "sysfs", "usbfs", "devfs", "fdescfs", "linprocfs" ]
