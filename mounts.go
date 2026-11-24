package main

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

// Mount contains all metadata for a single filesystem mount.
type Mount struct {
	Device     string      `json:"device"`
	DeviceType string      `json:"device_type"`
	Mountpoint string      `json:"mount_point"`
	Fstype     string      `json:"fs_type"`
	Type       string      `json:"type"`
	Opts       string      `json:"opts"`
	Total      uint64      `json:"total"`
	Free       uint64      `json:"free"`
	Used       uint64      `json:"used"`
	Inodes     uint64      `json:"inodes"`
	InodesFree uint64      `json:"inodes_free"`
	InodesUsed uint64      `json:"inodes_used"`
	Blocks     uint64      `json:"blocks"`
	BlockSize  uint64      `json:"block_size"`
	Metadata   interface{} `json:"-"`
}

func readLines(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close() //nolint:errcheck // ignore error

	scanner := bufio.NewScanner(file)
	var s []string
	for scanner.Scan() {
		s = append(s, scanner.Text())
	}

	return s, scanner.Err()
}

func unescapeFstab(path string) string {
	if !strings.Contains(path, `\`) {
		return path
	}

	// Mountinfo escapes are three octal digits. Copy all other bytes verbatim.
	var decoded strings.Builder
	decoded.Grow(len(path))
	for i := 0; i < len(path); i++ {
		if path[i] == '\\' && i+3 < len(path) {
			if b, err := strconv.ParseUint(path[i+1:i+4], 8, 8); err == nil {
				decoded.WriteByte(byte(b))
				i += 3
				continue
			}
		}
		decoded.WriteByte(path[i])
	}
	return decoded.String()
}

//nolint:unused // used on BSD
func byteToString(orig []byte) string {
	n := -1
	l := -1
	for i, b := range orig {
		// skip left side null
		if l == -1 && b == 0 {
			continue
		}
		if l == -1 {
			l = i
		}

		if b == 0 {
			break
		}
		n = i + 1
	}
	if n == -1 {
		return string(orig)
	}
	return string(orig[l:n])
}

//nolint:unused // used on OpenBSD
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
