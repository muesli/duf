package main

import (
	"bufio"
	"os"
	"strconv"

	"golang.org/x/sys/unix"
)

type Mount struct {
	Device      string        `json:"device"`
	DeviceType  string        `json:"device_type"`
	Mountpoint  string        `json:"mount_point"`
	Fstype      string        `json:"fs_type"`
	Type        string        `json:"type"`
	Opts        string        `json:"opts"`
	Total       uint64        `json:"total"`
	Free        uint64        `json:"free"`
	Used        uint64        `json:"used"`
	InodesTotal uint64        `json:"inodes_total"`
	InodesFree  uint64        `json:"inodes_free"`
	InodesUsed  uint64        `json:"inodes_used"`
	Stat        unix.Statfs_t `json:"-"`
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
