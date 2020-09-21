package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/sys/unix"
)

type Mount struct {
	Device     string
	Mountpoint string
	Fstype     string
	Type       string
	Opts       string
	Stat       unix.Statfs_t
	Total      uint64
	Free       uint64
	Used       uint64
}

func mounts() ([]Mount, error) {
	filename := "/proc/self/mountinfo"
	lines, err := readLines(filename)
	if err != nil {
		return nil, err
	}

	ret := make([]Mount, 0, len(lines))
	for _, line := range lines {
		// a line of self/mountinfo has the following structure:
		// 36  35  98:0 /mnt1 /mnt2 rw,noatime master:1 - ext3 /dev/root rw,errors=continue
		// (1) (2) (3)   (4)   (5)      (6)      (7)   (8) (9)   (10)         (11)

		// split the mountinfo line by the separator hyphen
		parts := strings.Split(line, " - ")
		if len(parts) != 2 {
			return nil, fmt.Errorf("found invalid mountinfo line in file %s: %s ", filename, line)
		}

		fields := strings.Fields(parts[0])
		// blockDeviceID := fields[2]
		mountPoint := fields[4]
		mountOpts := fields[5]

		fields = strings.Fields(parts[1])
		fstype := fields[0]
		device := fields[1]

		var stat unix.Statfs_t
		err := unix.Statfs(unescapeFstab(mountPoint), &stat)
		if err != nil {
			fmt.Printf("%s: %s\n", unescapeFstab(mountPoint), err)
			continue
		}

		d := Mount{
			Device:     device,
			Mountpoint: unescapeFstab(mountPoint),
			Fstype:     fstype,
			Type:       fsTypeMap[int64(stat.Type)],
			Opts:       mountOpts,
			Stat:       stat,
			Total:      (uint64(stat.Blocks) * uint64(stat.Bsize)),
			Free:       (uint64(stat.Bavail) * uint64(stat.Bsize)),
			Used:       (uint64(stat.Blocks) - uint64(stat.Bfree)) * uint64(stat.Bsize),
		}

		if strings.HasPrefix(d.Device, "/dev/mapper/") {
			re := regexp.MustCompile(`^\/dev\/mapper\/(.*)-(.*)`)
			match := re.FindAllStringSubmatch(d.Device, -1)
			if len(match) > 0 && len(match[0]) == 3 {
				d.Device = filepath.Join("/dev", match[0][1], match[0][2])
			}
		}

		ret = append(ret, d)
	}

	return ret, nil
}

func readLines(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var s []string
	for scanner.Scan() {
		s = append(s, scanner.Text())
	}

	return s, scanner.Err()
}

func unescapeFstab(path string) string {
	escaped, err := strconv.Unquote(`"` + path + `"`)
	if err != nil {
		return path
	}
	return escaped
}

func StringsHas(target []string, src string) bool {
	for _, t := range target {
		if strings.TrimSpace(t) == src {
			return true
		}
	}
	return false
}
