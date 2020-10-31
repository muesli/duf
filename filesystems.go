package main

func deviceType(m Mount) string {
	if isNetworkFs(m) {
		return networkDevice
	}
	if isSpecialFs(m) {
		return specialDevice
	}
	if isFuseFs(m) {
		return fuseDevice
	}

	return localDevice
}

// remote: [ "nfs", "smbfs", "cifs", "ncpfs", "afs", "coda", "ftpfs", "mfs", "sshfs", "fuse.sshfs", "nfs4" ]
// special: [ "tmpfs", "devpts", "devtmpfs", "proc", "sysfs", "usbfs", "devfs", "fdescfs", "linprocfs" ]
