//go:build linux
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

const (
	// A line of self/mountinfo has the following structure:
	// 36  35  98:0 /mnt1 /mnt2 rw,noatime master:1 - ext3 /dev/root rw,errors=continue
	// (1) (2) (3)   (4)   (5)      (6)      (7)   (8) (9)   (10)         (11)
	//
	// (1) mount ID: unique identifier of the mount (may be reused after umount).
	//mountinfoMountID = 1
	// (2) parent ID: ID of parent (or of self for the top of the mount tree).
	//mountinfoParentID = 2
	// (3) major:minor: value of st_dev for files on filesystem.
	//mountinfoMajorMinor = 3
	// (4) root: root of the mount within the filesystem.
	//mountinfoRoot = 4
	// (5) mount point: mount point relative to the process's root.
	mountinfoMountPoint = 5
	// (6) mount options: per mount options.
	mountinfoMountOpts = 6
	// (7) optional fields: zero or more fields terminated by "-".
	mountinfoOptionalFields = 7
	// (8) separator between optional fields.
	//mountinfoSeparator = 8
	// (9) filesystem type: name of filesystem of the form.
	mountinfoFsType = 9
	// (10) mount source: filesystem specific information or "none".
	mountinfoMountSource = 10
	// (11) super options: per super block options.
	//mountinfoSuperOptions = 11
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
		// Get fields from line
		nb, fields := getFields(line)

		// If no field finded, skip this line
		if nb == 0 {
			continue
		}

		if nb < 11 {
			return nil, nil, fmt.Errorf("found invalid mountinfo line in file %s: %s", filename, line)
		}

		// blockDeviceID := fields[mountinfoMountID]
		mountPoint := unescapeFstab(fields[mountinfoMountPoint])
		mountOpts := fields[mountinfoMountOpts]
		fstype := unescapeFstab(fields[mountinfoFsType])
		device := unescapeFstab(fields[mountinfoMountSource])

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
			Type:       fsTypeMap[int64(stat.Type)], //nolint:unconvert
			Opts:       mountOpts,
			Metadata:   stat,
			Total:      (uint64(stat.Blocks) * uint64(stat.Bsize)),                      //nolint:unconvert
			Free:       (uint64(stat.Bavail) * uint64(stat.Bsize)),                      //nolint:unconvert
			Used:       (uint64(stat.Blocks) - uint64(stat.Bfree)) * uint64(stat.Bsize), //nolint:unconvert
			Inodes:     stat.Files,
			InodesFree: stat.Ffree,
			InodesUsed: stat.Files - stat.Ffree,
			Blocks:     uint64(stat.Blocks), //nolint:unconvert
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

// getFields reads a row to extract the fields.
// it returns the number of fields found and the fields.
func getFields(line string) (nb int, fields [12]string) {
	// Ignore commented or empty line
	if line == "" || line[0] == '#' {
		return
	}

	nb = 1
	for _, f := range strings.Fields(line) {
		if nb == mountinfoOptionalFields {
			// (7)  optional fields: zero or more fields of the form
			//        "tag[:value]"; see below.
			// (8)  separator: the end of the optional fields is marked
			//        by a single hyphen.
			if f != "-" {
				fields[nb] += " " + f
				continue
			}

			nb++
		}

		// Assign the value of the field to the corresponding index
		fields[nb] = f
		nb++
	}

	fields[mountinfoMountPoint] = decodeName(fields[mountinfoMountPoint])
	fields[mountinfoMountSource] = decodeName(fields[mountinfoMountSource])

	return
}

// decodeName returns the decoded name
// A name cannot contain spaces, tabs, new lines or backslashes.
// Therefore, some programs encode them by "\040", "\011", "\012" and "\134".
func decodeName(n string) string {
	l := len(n)
	for i := 0; i < l; i++ {
		// if there is no thing to decode
		if i+3 >= l {
			break
		}

		// if rune is not a backslash
		if n[i] != '\\' {
			continue
		}

		// reading 3 bytes to decode the rune
		switch {
		case n[i+1] == '0' && n[i+2] == '4' && n[i+3] == '0':
			n = n[:i] + " " + n[i+4:]
		case n[i+1] == '0' && n[i+2] == '1' && n[i+3] == '1':
			n = n[:i] + "\t" + n[i+4:]
		case n[i+1] == '0' && n[i+2] == '1' && n[i+3] == '2':
			n = n[:i] + "\n" + n[i+4:]
		case n[i+1] == '1' && n[i+2] == '3' && n[i+3] == '4':
			n = n[:i] + "\\" + n[i+4:]
		default:
			continue
		}

		l -= 3
	}

	return n
}
