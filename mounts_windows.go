// +build windows

package main

import (
	"fmt"
	"golang.org/x/sys/windows"
	"syscall"
)

// Local devices
const (
	guidBufLen       = windows.MAX_PATH + 1
	volumeNameBufLen = windows.MAX_PATH + 1
	rootPathBufLen   = windows.MAX_PATH + 1
	fileSystemBufLen = windows.MAX_PATH + 1
)

func getMountPoint(guidBuf []uint16) (mountPoint string, err error) {
	var rootPathLen uint32
	rootPathBuf := make([]uint16, rootPathBufLen)

	err = windows.GetVolumePathNamesForVolumeName(&guidBuf[0], &rootPathBuf[0], rootPathBufLen*2, &rootPathLen)
	if err != nil && err.(windows.Errno) == windows.ERROR_MORE_DATA {
		// Retry if buffer size is too small
		rootPathBuf = make([]uint16, (rootPathLen+1)/2)
		err = windows.GetVolumePathNamesForVolumeName(
			&guidBuf[0], &rootPathBuf[0], rootPathLen, &rootPathLen)
	}
	return windows.UTF16ToString(rootPathBuf), err
}

func getVolumeInfo(guidOrMountPointBuf []uint16) (volumeName string, fsType string, err error) {
	volumeNameBuf := make([]uint16, volumeNameBufLen)
	fsTypeBuf := make([]uint16, fileSystemBufLen)

	err = windows.GetVolumeInformation(&guidOrMountPointBuf[0], &volumeNameBuf[0], volumeNameBufLen*2,
		nil, nil, nil,
		&fsTypeBuf[0], fileSystemBufLen*2)

	return windows.UTF16ToString(volumeNameBuf), windows.UTF16ToString(fsTypeBuf), err
}

func getSpaceInfo(guidOrMountPointBuf []uint16) (totalBytes uint64, freeBytes uint64, err error) {
	err = windows.GetDiskFreeSpaceEx(&guidOrMountPointBuf[0], nil, &totalBytes, &freeBytes)
	return
}

func getMountFromGUID(guidBuf []uint16) (m Mount, skip bool, warnings []string) {
	var err error
	guid := windows.UTF16ToString(guidBuf)

	mountPoint, err := getMountPoint(guidBuf)
	if err != nil {
		warnings = append(warnings, fmt.Sprintf("%s: %s", guid, err))
	}
	// Skip unmounted volumes
	if len(mountPoint) == 0 {
		skip = true
		return
	}

	// Get volume name & filesystem type
	volumeName, fsType, err := getVolumeInfo(guidBuf)
	if err != nil {
		warnings = append(warnings, fmt.Sprintf("%s: %s", guid, err))
	}

	// Get space info
	totalBytes, freeBytes, err := getSpaceInfo(guidBuf)
	if err != nil {
		warnings = append(warnings, fmt.Sprintf("%s: %s", guid, err))
	}

	// Use GUID as volume name if no label was set
	if len(volumeName) == 0 {
		volumeName = guid
	}

	m = Mount{
		Device:     volumeName,
		Mountpoint: mountPoint,
		Fstype:     fsType,
		Type:       fsType,
		Opts:       "",
		Total:      totalBytes,
		Free:       freeBytes,
		Used:       totalBytes - freeBytes,
	}
	m.DeviceType = deviceType(m)
	return
}

func appendLocalMounts(mounts []Mount, warnings []string) ([]Mount, []string, error) {
	guidBuf := make([]uint16, guidBufLen)

	hFindVolume, err := windows.FindFirstVolume(&guidBuf[0], guidBufLen*2)
	if err != nil {
		return mounts, warnings, err
	}

VolumeLoop:
	for ; ; err = windows.FindNextVolume(hFindVolume, &guidBuf[0], guidBufLen*2) {
		if err != nil {
			switch err.(windows.Errno) {
			case windows.ERROR_NO_MORE_FILES:
				break VolumeLoop
			default:
				warnings = append(warnings, fmt.Sprintf("%s: %s", windows.UTF16ToString(guidBuf), err))
				continue VolumeLoop
			}
		}

		if m, skip, w := getMountFromGUID(guidBuf); !skip {
			mounts = append(mounts, m)
			warnings = append(warnings, w...)
		}
	}

	if err = windows.FindVolumeClose(hFindVolume); err != nil {
		warnings = append(warnings, fmt.Sprintf("%s", err))
	}
	return mounts, warnings, nil
}

func mounts() (ret []Mount, warnings []string, err error) {
	ret = make([]Mount, 0)

	// Local devices
	if ret, warnings, err = appendLocalMounts(ret, warnings); err != nil {
		return
	}

	return ret, warnings, nil
}
