//go:build netbsd
// +build netbsd

package main

import (
	"unsafe"
	"golang.org/x/sys/unix"
)

func (m *Mount) Stat() unix.Statfs_t {
	return m.Metadata.(unix.Statfs_t)
}

// Use unix.Getvfsstat when CL 550476 is merged
// https://go-review.googlesource.com/c/sys/+/550476
func Getvfsstat(buf []unix.Statvfs_t, flags int) (n int, err error) {
	var (
		_p0     unsafe.Pointer
		bufsize uintptr
	)
	if len(buf) > 0 {
		_p0 = unsafe.Pointer(&buf[0])
		bufsize = unsafe.Sizeof(unix.Statvfs_t{}) * uintptr(len(buf))
	}
	r0, _, e1 := unix.Syscall(unix.SYS_GETVFSSTAT, uintptr(_p0), bufsize, uintptr(flags))
	n = int(r0)
	if e1 != 0 {
		err = e1
	}
	return
}

func mounts() ([]Mount, []string, error) {
	var ret []Mount
	var warnings []string

	count, err := Getvfsstat(nil, unix.ST_WAIT)
	if err != nil {
		return nil, nil, err
	}

	fs := make([]unix.Statvfs_t, count)
	if _, err := Getvfsstat(fs, unix.ST_WAIT); err != nil {
		return nil, nil, err
	}

	for _, stat := range fs {
		opts := "rw"
		if stat.Flag&unix.MNT_RDONLY != 0 {
			opts = "ro"
		}
		if stat.Flag&unix.MNT_SYNCHRONOUS != 0 {
			opts += ",sync"
		}
		if stat.Flag&unix.MNT_NOEXEC != 0 {
			opts += ",noexec"
		}
		if stat.Flag&unix.MNT_NOSUID != 0 {
			opts += ",nosuid"
		}
		if stat.Flag&unix.MNT_NODEV != 0 {
			opts += ",nodev"
		}
		if stat.Flag&unix.MNT_ASYNC != 0 {
			opts += ",async"
		}
		if stat.Flag&unix.MNT_SOFTDEP != 0 {
			opts += ",softdep"
		}
		if stat.Flag&unix.MNT_NOATIME != 0 {
			opts += ",noatime"
		}

		device := byteToString(stat.Mntfromname[:])
		mountPoint := byteToString(stat.Mntonname[:])
		fsType := byteToString(stat.Fstypename[:])

		if len(device) == 0 {
			continue
		}

		d := Mount{
			Device:     device,
			Mountpoint: mountPoint,
			Fstype:     fsType,
			Type:       fsType,
			Opts:       opts,
			Metadata:   stat,
			Total:      (uint64(stat.Blocks) * uint64(stat.Bsize)),
			Free:       (uint64(stat.Bavail) * uint64(stat.Bsize)),
			Used:       (uint64(stat.Blocks) - uint64(stat.Bfree)) * uint64(stat.Bsize),
			Inodes:     stat.Files,
			InodesFree: uint64(stat.Ffree),
			InodesUsed: stat.Files - uint64(stat.Ffree),
			Blocks:     uint64(stat.Blocks),
			BlockSize:  uint64(stat.Bsize),
		}
		d.DeviceType = deviceType(d)

		ret = append(ret, d)
	}

	return ret, warnings, nil
}
