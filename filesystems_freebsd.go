// +build freebsd

package main

func isLocalFs(m Mount) bool {
	//FIXME: implement
	return false
}

func isFuseFs(m Mount) bool {
	//FIXME: implement
	return false
}

func isNetworkFs(m Mount) bool {
	//FIXME: implement
	return false
}

func isSpecialFs(m Mount) bool {
	return m.Fstype == "devfs"
}
