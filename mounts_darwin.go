// +build darwin

package main

import (
	"fmt"
	"os"

	"golang.org/x/sys/unix"
)

func mounts() ([]Mount, []string, error) {
	var ret []Mount
	var warnings []string

	count, err := unix.Getfsstat(nil, unix.MNT_WAIT)
	if err != nil {
		return nil, nil, err
	}
	fs := make([]unix.Statfs_t, count)
	if _, err = unix.Getfsstat(fs, unix.MNT_WAIT); err != nil {
		return nil, nil, err
	}

	for _, fstat := range fs {
		opts := "rw"
		if fstat.Flags&unix.MNT_RDONLY != 0 {
			opts = "ro"
		}
		if fstat.Flags&unix.MNT_SYNCHRONOUS != 0 {
			opts += ",sync"
		}
		if fstat.Flags&unix.MNT_NOEXEC != 0 {
			opts += ",noexec"
		}
		if fstat.Flags&unix.MNT_NOSUID != 0 {
			opts += ",nosuid"
		}
		if fstat.Flags&unix.MNT_UNION != 0 {
			opts += ",union"
		}
		if fstat.Flags&unix.MNT_ASYNC != 0 {
			opts += ",async"
		}
		if fstat.Flags&unix.MNT_DONTBROWSE != 0 {
			opts += ",nobrowse"
		}
		if fstat.Flags&unix.MNT_AUTOMOUNTED != 0 {
			opts += ",automounted"
		}
		if fstat.Flags&unix.MNT_JOURNALED != 0 {
			opts += ",journaled"
		}
		if fstat.Flags&unix.MNT_MULTILABEL != 0 {
			opts += ",multilabel"
		}
		if fstat.Flags&unix.MNT_NOATIME != 0 {
			opts += ",noatime"
		}
		if fstat.Flags&unix.MNT_NODEV != 0 {
			opts += ",nodev"
		}

		device := intToString(fstat.Mntfromname[:])
		mountPoint := intToString(fstat.Mntonname[:])
		fsType := intToString(fstat.Fstypename[:])

		if len(device) == 0 {
			continue
		}

		var stat unix.Statfs_t
		err := unix.Statfs(mountPoint, &stat)
		if err != nil {
			if err != os.ErrPermission {
				warnings = append(warnings, fmt.Sprintf("%s: %s", mountPoint, err))
				continue
			}

			stat = unix.Statfs_t{}
		}

		d := Mount{
			Device:     device,
			Mountpoint: mountPoint,
			Fstype:     fsType,
			Type:       fsType,
			Opts:       opts,
			Stat:       stat,
			Total:      (uint64(stat.Blocks) * uint64(stat.Bsize)),
			Free:       (uint64(stat.Bavail) * uint64(stat.Bsize)),
			Used:       (uint64(stat.Blocks) - uint64(stat.Bfree)) * uint64(stat.Bsize),
		}
		d.DeviceType = deviceType(d)

		ret = append(ret, d)
	}

	return ret, warnings, nil
}

func intToString(orig []int8) string {
	ret := make([]byte, len(orig))
	size := -1
	for i, o := range orig {
		if o == 0 {
			size = i
			break
		}
		ret[i] = byte(o)
	}
	if size == -1 {
		size = len(orig)
	}

	return string(ret[0:size])
}
