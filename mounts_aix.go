//go:build aix
// +build aix

package main

import (
	"bufio"
	"fmt"
	"os/exec"
	"strings"

	"golang.org/x/sys/unix"
)

// Stat returns the mountpoint's stat information.
func (m *Mount) Stat() unix.Statfs_t {
	return m.Metadata.(unix.Statfs_t)
}

// mounts returns all mounted filesystems from AIX.
// AIX mount output format (after header):
//   /dev/hd4  /  jfs2  Dec 30 22:15 rw,log=/dev/hd8
// Fields: device mountpoint fstype month day time options
func mounts() ([]Mount, []string, error) {
	var ret []Mount
	var warnings []string

	cmd := exec.Command("mount")
	out, err := cmd.Output()
	if err != nil {
		return nil, nil, fmt.Errorf("running mount command: %w", err)
	}

	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// Skip header lines (first 2 lines)
		if lineNum <= 2 {
			continue
		}

		// Skip empty lines
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 6 {
			continue
		}

		// Fields: device mountpoint fstype month day time [options]
		device := fields[0]
		mountPoint := fields[1]
		fstype := fields[2]
		// fields[3], fields[4], fields[5] are date/time
		opts := ""
		if len(fields) >= 7 {
			opts = fields[6]
		}

		var stat unix.Statfs_t
		err := unix.Statfs(mountPoint, &stat)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("%s: %s", mountPoint, err))
			continue
		}

		d := Mount{
			Device:     device,
			Mountpoint: mountPoint,
			Fstype:     fstype,
			Type:       fstype,
			Opts:       opts,
			Metadata:   stat,
			Total:      uint64(stat.Blocks) * uint64(stat.Bsize),
			Free:       uint64(stat.Bavail) * uint64(stat.Bsize),
			Used:       (uint64(stat.Blocks) - uint64(stat.Bfree)) * uint64(stat.Bsize),
			Inodes:     stat.Files,
			InodesFree: stat.Ffree,
			InodesUsed: stat.Files - stat.Ffree,
			Blocks:     uint64(stat.Blocks),
			BlockSize:  uint64(stat.Bsize),
		}
		d.DeviceType = deviceType(d)

		ret = append(ret, d)
	}

	return ret, warnings, nil
}
