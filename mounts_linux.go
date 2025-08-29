//go:build linux
// +build linux

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/sys/unix"
)

const (
	// A line of self/mountinfo has the following structure:
	// 36  35  98:0 /mnt1 /mnt2 rw,noatime master:1 - ext3 /dev/root rw,errors=continue
	// (0) (1) (2)   (3)   (4)      (5)      (6)   (7) (8)    (9)           (10)
	//
	// (0) mount ID: unique identifier of the mount (may be reused after umount).
	//mountinfoMountID = 0
	// (1) parent ID: ID of parent (or of self for the top of the mount tree).
	//mountinfoParentID = 1
	// (2) major:minor: value of st_dev for files on filesystem.
	//mountinfoMajorMinor = 2
	// (3) root: root of the mount within the filesystem.
	//mountinfoRoot = 3
	// (4) mount point: mount point relative to the process's root.
	mountinfoMountPoint = 4
	// (5) mount options: per mount options.
	mountinfoMountOpts = 5
	// (6) optional fields: zero or more fields terminated by "-".
	mountinfoOptionalFields = 6
	// (7) separator between optional fields.
	//mountinfoSeparator = 7
	// (8) filesystem type: name of filesystem of the form.
	mountinfoFsType = 8
	// (9) mount source: filesystem specific information or "none".
	mountinfoMountSource = 9
	// (10) super options: per super block options.
	mountinfoSuperOptions = 10
)

// Stat returns the mountpoint's stat information.
func (m *Mount) Stat() unix.Statfs_t {
	return m.Metadata.(unix.Statfs_t)
}

func mounts() ([]Mount, []string, error) {
	var warnings []string

	filename := "/proc/self/mountinfo"
	lines, err := readLines(filename)
	if err != nil {
		// wrapcheck: add context to the error.
		return nil, nil, fmt.Errorf("reading mountinfo %q: %w", filename, err)
	}

	ret := make([]Mount, 0, len(lines))
	for _, line := range lines {
		nb, fields := parseMountInfoLine(line)
		if nb == 0 {
			continue
		}

		// if the number of fields does not match the structure of mountinfo,
		// emit a warning and ignore the line.
		if nb < 10 || nb > 11 {
			warnings = append(warnings, fmt.Sprintf("found invalid mountinfo line: %s", line))
			continue
		}

		// blockDeviceID := fields[mountinfoMountID]
		mountPoint := fields[mountinfoMountPoint]
		mountOpts := fields[mountinfoMountOpts]
		fstype := fields[mountinfoFsType]
		device := fields[mountinfoMountSource]

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

		// Resolve /dev/mapper/* device names.
		if strings.HasPrefix(d.Device, "/dev/mapper/") {
			re := regexp.MustCompile(`^/dev/mapper/(.*)-(.*)`)
			match := re.FindAllStringSubmatch(d.Device, -1)
			if len(match) > 0 && len(match[0]) == 3 {
				d.Device = filepath.Join("/dev", match[0][1], match[0][2])
			}
		}

		ret = append(ret, d)
	}

	return ret, warnings, nil
}

// splitMountInfoFields splits a mountinfo line into its fields.
// It treats spaces and tabs as field separators and decodes certain octal escapes.
func splitMountInfoFields(line string) []string {
	var fields []string
	var buf strings.Builder

	for i := 0; i < len(line); i++ {
		c := line[i]

		// Treat both space and tab as separators
		if c == ' ' || c == '\t' {
			if buf.Len() > 0 {
				fields = append(fields, buf.String())
				buf.Reset()
			}
			continue
		}

		if c == '\\' && i+3 < len(line) {
			oct := line[i+1 : i+4]
			if v, err := strconv.ParseInt(oct, 8, 0); err == nil {
				switch byte(v) {
				case ' ', '\t', '\n':
					buf.WriteByte(byte(v))
					i += 3
					continue
				default:
					// keep unknown escapes as-is
					buf.WriteString("\\" + oct)
					i += 3
					continue
				}
			}
		}

		buf.WriteByte(c)
	}

	if buf.Len() > 0 {
		fields = append(fields, buf.String())
	}

	return fields
}

func parseMountInfoLine(line string) (int, [11]string) {
	var fields [11]string

	if len(line) == 0 || line[0] == '#' {
		// ignore comments and empty lines
		return 0, fields
	}

	all := splitMountInfoFields(line)

	var i int
	sawSep := false
	sawSup := false

	for _, f := range all {
		if i >= len(fields) {
			break
		}

		if i == mountinfoOptionalFields {
			// (6)  optional fields: zero or more fields of the form "tag[:value]"; see below.
			// (7)  separator: the end of the optional fields is marked by a single hyphen.
			if f != "-" {
				// Join tokens with spaces for mountinfoOptionalFields.
				fields[i] = strings.TrimSpace(fields[i] + " " + f)
				continue
			}
			// Found separator.
			sawSep = true
			i++
			fields[i] = f
			i++
			continue
		}

		if i == mountinfoSuperOptions {
			// join tokens with spaces for WSL2 path=... they are splitted around spaces.
			fields[i] = strings.TrimSpace(fields[i] + " " + f)
			sawSup = true
			continue
		}

		// Default case: copy with unescape for certain fields
		switch i {
		case mountinfoMountPoint, mountinfoMountSource, mountinfoFsType:
			fields[i] = unescapeFstab(f)
		default:
			fields[i] = f
		}
		i++
	}

	// Handle malformed line (no "-" found).
	if !sawSep && len(all) > mountinfoOptionalFields {
		i = mountinfoOptionalFields
	}

	// When super options are present, the index is one less than 11.
	if sawSup {
		i++
	}

	// clear trailing empties.
	for j := i + 1; j < len(fields); j++ {
		fields[j] = ""
	}

	return i, fields
}
