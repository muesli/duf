package main

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// stripVerbatimPrefix removes the "\\?\" or "\\?\UNC\" prefix from a
// Windows path. Such verbatim paths bypass path normalization, so
// filepath.Abs and filepath.EvalSymlinks preserve the prefix - while
// mount points are reported in their canonical (non-verbatim) form and
// would never match. On other platforms the path is returned unchanged.
func stripVerbatimPrefix(path string) string {
	if runtime.GOOS != "windows" {
		return path
	}

	const prefix = `\\?\`
	if !strings.HasPrefix(path, prefix) {
		return path
	}

	rest := path[len(prefix):]
	// \\?\UNC\host\share -> \\host\share
	if strings.HasPrefix(rest, `UNC\`) {
		return `\\` + rest[len(`UNC\`):]
	}
	// \\?\D:\foo -> D:\foo, but only for drive-letter paths: other
	// verbatim forms (e.g. \\?\Volume{GUID}\) have no plain equivalent.
	if len(rest) > 1 && rest[1] == ':' &&
		(rest[0] >= 'a' && rest[0] <= 'z' || rest[0] >= 'A' && rest[0] <= 'Z') {
		return rest
	}
	return path
}

func findMounts(mounts []Mount, path string) ([]Mount, error) {
	var err error
	path = stripVerbatimPrefix(path)
	path, err = filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	path, err = filepath.EvalSymlinks(path)
	if err != nil {
		return nil, err
	}
	// a symlink may resolve to a verbatim path
	path = stripVerbatimPrefix(path)

	_, err = os.Stat(path)
	if err != nil {
		return nil, err
	}

	var m []Mount
	for _, v := range mounts {
		if path == v.Device {
			return []Mount{v}, nil
		}

		if strings.HasPrefix(path, v.Mountpoint) {
			var nm []Mount

			// keep all entries that are as close or closer to the target
			for _, mv := range m {
				if len(mv.Mountpoint) >= len(v.Mountpoint) {
					nm = append(nm, mv)
				}
			}
			m = nm

			// add entry only if we didn't already find something closer
			if len(nm) == 0 || len(v.Mountpoint) >= len(nm[0].Mountpoint) {
				m = append(m, v)
			}
		}
	}

	return m, nil
}

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
