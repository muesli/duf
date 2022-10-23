//go:build darwin
// +build darwin

package main

import (
	"syscall"
	"unsafe"

	"golang.org/x/sys/unix"
)

type attrlist struct {
	bitmapcount uint16
	reserved    uint16
	commonattr  uint32
	volattr     uint32
	dirattr     uint32
	fileattr    uint32
	forkattr    uint32
}

type volAttrs struct {
	length    uint32
	spaceUsed [8]byte
}

const ATTR_BIT_MAP_COUNT = 5
const ATTR_VOL_INFO = 0x80000000
const ATTR_VOL_SPACEUSED = 0x00800000

func (m *Mount) Stat() unix.Statfs_t {
	return m.Metadata.(unix.Statfs_t)
}

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

	for _, stat := range fs {
		opts := "rw"
		if stat.Flags&unix.MNT_RDONLY != 0 {
			opts = "ro"
		}
		if stat.Flags&unix.MNT_SYNCHRONOUS != 0 {
			opts += ",sync"
		}
		if stat.Flags&unix.MNT_NOEXEC != 0 {
			opts += ",noexec"
		}
		if stat.Flags&unix.MNT_NOSUID != 0 {
			opts += ",nosuid"
		}
		if stat.Flags&unix.MNT_UNION != 0 {
			opts += ",union"
		}
		if stat.Flags&unix.MNT_ASYNC != 0 {
			opts += ",async"
		}
		if stat.Flags&unix.MNT_DONTBROWSE != 0 {
			opts += ",nobrowse"
		}
		if stat.Flags&unix.MNT_AUTOMOUNTED != 0 {
			opts += ",automounted"
		}
		if stat.Flags&unix.MNT_JOURNALED != 0 {
			opts += ",journaled"
		}
		if stat.Flags&unix.MNT_MULTILABEL != 0 {
			opts += ",multilabel"
		}
		if stat.Flags&unix.MNT_NOATIME != 0 {
			opts += ",noatime"
		}
		if stat.Flags&unix.MNT_NODEV != 0 {
			opts += ",nodev"
		}

		device := byteToString(stat.Mntfromname[:])
		mountPoint := byteToString(stat.Mntonname[:])
		fsType := byteToString(stat.Fstypename[:])

		if len(device) == 0 {
			continue
		}

		used := (stat.Blocks - stat.Bfree) * uint64(stat.Bsize)

		var mountPointPtr *byte
		mountPointPtr, err = syscall.BytePtrFromString(mountPoint)
		attrList := attrlist{
			bitmapcount: ATTR_BIT_MAP_COUNT,
			volattr:     ATTR_VOL_INFO | ATTR_VOL_SPACEUSED,
		}
		var volAttrs volAttrs
		volAttrsRes, _, _ := syscall.Syscall6(syscall.SYS_GETATTRLIST,
			uintptr(unsafe.Pointer(mountPointPtr)),
			uintptr(unsafe.Pointer(&attrList)),
			uintptr(unsafe.Pointer(&volAttrs)),
			unsafe.Sizeof(volAttrs),
			unix.FSOPT_NOFOLLOW,
			0)
		if volAttrsRes == 0 {
			used = *(*uint64)(unsafe.Pointer(&volAttrs.spaceUsed[0]))
		}

		d := Mount{
			Device:     device,
			Mountpoint: mountPoint,
			Fstype:     fsType,
			Type:       fsType,
			Opts:       opts,
			Metadata:   stat,
			Total:      stat.Blocks * uint64(stat.Bsize),
			Free:       stat.Bavail * uint64(stat.Bsize),
			Used:       used,
			Inodes:     stat.Files,
			InodesFree: stat.Ffree,
			InodesUsed: stat.Files - stat.Ffree,
			Blocks:     stat.Blocks,
			BlockSize:  uint64(stat.Bsize),
		}
		d.DeviceType = deviceType(d)

		ret = append(ret, d)
	}

	return ret, warnings, nil
}
