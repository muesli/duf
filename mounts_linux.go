// +build linux

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"golang.org/x/sys/unix"
)

func (m *Mount) Stat() unix.Statfs_t {
	return m.Metadata.(unix.Statfs_t)
}

func mounts() ([]Mount, []string, error) {
	var warnings []string

	filename := "/proc/self/mountinfo"
	lines, err := readLines(filename)
	if err != nil {
		return nil, nil, err
	}

	ret := make([]Mount, 0, len(lines))
	for _, line := range lines {
		// a line of self/mountinfo has the following structure:
		// 36  35  98:0 /mnt1 /mnt2 rw,noatime master:1 - ext3 /dev/root rw,errors=continue
		// (1) (2) (3)   (4)   (5)      (6)      (7)   (8) (9)   (10)         (11)

		// split the mountinfo line by the separator hyphen
		parts := strings.Split(line, " - ")
		if len(parts) != 2 {
			return nil, nil, fmt.Errorf("found invalid mountinfo line in file %s: %s", filename, line)
		}

		fields := strings.Fields(parts[0])
		// blockDeviceID := fields[2]
		mountPoint := unescapeFstab(fields[4])
		mountOpts := fields[5]

		fields = strings.Fields(parts[1])
		fstype := unescapeFstab(fields[0])
		device := unescapeFstab(fields[1])

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
			Fstype:     fstype,
			Type:       fsTypeMap[int64(stat.Type)],
			Opts:       mountOpts,
			Metadata:   stat,
			Total:      (uint64(stat.Blocks) * uint64(stat.Bsize)),
			Free:       (uint64(stat.Bavail) * uint64(stat.Bsize)),
			Used:       (uint64(stat.Blocks) - uint64(stat.Bfree)) * uint64(stat.Bsize),
			Inodes:     stat.Files,
			InodesFree: stat.Ffree,
			InodesUsed: stat.Files - stat.Ffree,
			Blocks:     uint64(stat.Blocks),
			BlockSize:  uint64(stat.Bsize),
		}
		d.DeviceType = deviceType(d)

		// resolve /dev/mapper/* device names
		if strings.HasPrefix(d.Device, "/dev/mapper/") {
			re := regexp.MustCompile(`^\/dev\/mapper\/(.*)-(.*)`)
			match := re.FindAllStringSubmatch(d.Device, -1)
			if len(match) > 0 && len(match[0]) == 3 {
				d.Device = filepath.Join("/dev", match[0][1], match[0][2])
			}
		}

		ret = append(ret, d)
	}

	return ret, warnings, nil
}
