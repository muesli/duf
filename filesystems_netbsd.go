//go:build netbsd
// +build netbsd

package main

func isFuseFs(m Mount) bool {
	//FIXME: implement
	return false
}

func isNetworkFs(m Mount) bool {
	fs := []string{"nfs", "smbfs"}

	for _, v := range fs {
		if m.Fstype == v {
			return true
		}
	}

	return false
}

func isSpecialFs(m Mount) bool {
	fs := []string{"devfs", "kernfs", "procfs", "ptyfs", "tmpfs"}

	for _, v := range fs {
		if m.Fstype == v {
			return true
		}
	}

	return false
}

func isHiddenFs(m Mount) bool {
	return false
}
