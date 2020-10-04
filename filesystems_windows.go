// +build windows

package main

func isFuseFs(m Mount) bool {
	//FIXME: implement
	return false
}

func isNetworkFs(m Mount) bool {
	_, ok := m.Metadata.(*NetResource)
	return ok
}

func isSpecialFs(m Mount) bool {
	//FIXME: implement
	return false
}
