package main

import (
	"strings"
)

const (
	localDevice   = "local"
	networkDevice = "network"
	fuseDevice    = "fuse"
	specialDevice = "special"
	loopsDevice   = "loops"
	bindsMount    = "binds"
)

// FilterOptions contains all filters.
type FilterOptions struct {
	HiddenDevices map[string]struct{}
	OnlyDevices   map[string]struct{}

	HiddenFilesystems map[string]struct{}
	OnlyFilesystems   map[string]struct{}

	HiddenMountPoints map[string]struct{}
	OnlyMountPoints   map[string]struct{}
}

// renderTables renders all tables.
func renderTables(m []Mount, filters FilterOptions, opts TableOptions) {
	deviceMounts := make(map[string][]Mount)
	hasOnlyDevices := len(filters.OnlyDevices) != 0

	_, hideLocal := filters.HiddenDevices[localDevice]
	_, hideNetwork := filters.HiddenDevices[networkDevice]
	_, hideFuse := filters.HiddenDevices[fuseDevice]
	_, hideSpecial := filters.HiddenDevices[specialDevice]
	_, hideLoops := filters.HiddenDevices[loopsDevice]
	_, hideBinds := filters.HiddenDevices[bindsMount]

	_, onlyLocal := filters.OnlyDevices[localDevice]
	_, onlyNetwork := filters.OnlyDevices[networkDevice]
	_, onlyFuse := filters.OnlyDevices[fuseDevice]
	_, onlySpecial := filters.OnlyDevices[specialDevice]
	_, onlyLoops := filters.OnlyDevices[loopsDevice]
	_, onlyBinds := filters.OnlyDevices[bindsMount]

	// sort/filter devices
	for _, v := range m {
		if len(filters.OnlyFilesystems) != 0 {
			// skip not onlyFs
			if _, ok := filters.OnlyFilesystems[strings.ToLower(v.Fstype)]; !ok {
				continue
			}
		} else {
			// skip hideFs
			if _, ok := filters.HiddenFilesystems[strings.ToLower(v.Fstype)]; ok {
				continue
			}
		}

		// skip hidden devices
		if isHiddenFs(v) && !*all {
			continue
		}

		// skip bind-mounts
		if strings.Contains(v.Opts, "bind") {
			if (hasOnlyDevices && !onlyBinds) || (hideBinds && !*all) {
				continue
			}
		}

		// skip loop devices
		if strings.HasPrefix(v.Device, "/dev/loop") {
			if (hasOnlyDevices && !onlyLoops) || (hideLoops && !*all) {
				continue
			}
		}

		// skip special devices
		if v.Blocks == 0 && !*all {
			continue
		}

		// skip zero size devices
		if v.BlockSize == 0 && !*all {
			continue
		}

		// skip not only mount point
		if len(filters.OnlyMountPoints) != 0 {
			if !findInKey(v.Mountpoint, filters.OnlyMountPoints) {
				continue
			}
		}

		// skip hidden mount point
		if len(filters.HiddenMountPoints) != 0 {
			if findInKey(v.Mountpoint, filters.HiddenMountPoints) {
				continue
			}
		}

		t := deviceType(v)
		deviceMounts[t] = append(deviceMounts[t], v)
	}

	// print tables
	for _, devType := range groups {
		mounts := deviceMounts[devType]

		shouldPrint := *all
		if !shouldPrint {
			switch devType {
			case localDevice:
				shouldPrint = (hasOnlyDevices && onlyLocal) || (!hasOnlyDevices && !hideLocal)
			case networkDevice:
				shouldPrint = (hasOnlyDevices && onlyNetwork) || (!hasOnlyDevices && !hideNetwork)
			case fuseDevice:
				shouldPrint = (hasOnlyDevices && onlyFuse) || (!hasOnlyDevices && !hideFuse)
			case specialDevice:
				shouldPrint = (hasOnlyDevices && onlySpecial) || (!hasOnlyDevices && !hideSpecial)
			}
		}

		if shouldPrint {
			printTable(devType, mounts, opts)
		}
	}
}
