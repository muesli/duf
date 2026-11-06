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

// shouldSkipMount reports whether a mount is filtered out before display,
// independent of table grouping, so JSON output can share the same filtering.
func shouldSkipMount(v Mount, filters FilterOptions) bool {
	hasOnlyDevices := len(filters.OnlyDevices) != 0

	_, hideLoops := filters.HiddenDevices[loopsDevice]
	_, hideBinds := filters.HiddenDevices[bindsMount]

	_, onlyLoops := filters.OnlyDevices[loopsDevice]
	_, onlyBinds := filters.OnlyDevices[bindsMount]

	if len(filters.OnlyFilesystems) != 0 {
		// skip not onlyFs
		if _, ok := filters.OnlyFilesystems[strings.ToLower(v.Fstype)]; !ok {
			return true
		}
	} else {
		// skip hideFs
		if _, ok := filters.HiddenFilesystems[strings.ToLower(v.Fstype)]; ok {
			return true
		}
	}

	// skip hidden devices
	if isHiddenFs(v) && !*all {
		return true
	}

	// skip bind-mounts
	if strings.Contains(v.Opts, "bind") {
		if (hasOnlyDevices && !onlyBinds) || (hideBinds && !*all) {
			return true
		}
	}

	// skip loop devices
	if strings.HasPrefix(v.Device, "/dev/loop") {
		if (hasOnlyDevices && !onlyLoops) || (hideLoops && !*all) {
			return true
		}
	}

	// An explicitly requested device group stays visible even without size
	// info, so `-only fuse` finds zero-block fuse filesystems.
	_, onlyRequested := filters.OnlyDevices[deviceType(v)]
	showAnyway := hasOnlyDevices && onlyRequested

	// skip special devices
	if v.Blocks == 0 && !*all && !showAnyway {
		return true
	}

	// skip zero size devices
	if v.BlockSize == 0 && !*all && !showAnyway {
		return true
	}

	// skip not only mount point
	if len(filters.OnlyMountPoints) != 0 {
		if !findInKey(v.Mountpoint, filters.OnlyMountPoints) {
			return true
		}
	}

	// skip hidden mount point
	if len(filters.HiddenMountPoints) != 0 {
		if findInKey(v.Mountpoint, filters.HiddenMountPoints) {
			return true
		}
	}

	return false
}

// filterMounts drops every mount the filters skip, keeping only explicitly
// requested device groups, so JSON output matches the tables.
func filterMounts(m []Mount, filters FilterOptions) []Mount {
	mounts := m[:0]
	for _, v := range m {
		if shouldSkipMount(v, filters) {
			continue
		}
		// match the table group gate: an explicit `-only` wins over `-hide`
		if len(filters.OnlyDevices) != 0 {
			if _, ok := filters.OnlyDevices[deviceType(v)]; !ok {
				continue
			}
		} else if _, ok := filters.HiddenDevices[deviceType(v)]; ok {
			continue
		}
		mounts = append(mounts, v)
	}
	return mounts
}

// renderTables renders all tables.
func renderTables(m []Mount, filters FilterOptions, opts TableOptions) {
	deviceMounts := make(map[string][]Mount)
	hasOnlyDevices := len(filters.OnlyDevices) != 0

	_, hideLocal := filters.HiddenDevices[localDevice]
	_, hideNetwork := filters.HiddenDevices[networkDevice]
	_, hideFuse := filters.HiddenDevices[fuseDevice]
	_, hideSpecial := filters.HiddenDevices[specialDevice]

	_, onlyLocal := filters.OnlyDevices[localDevice]
	_, onlyNetwork := filters.OnlyDevices[networkDevice]
	_, onlyFuse := filters.OnlyDevices[fuseDevice]
	_, onlySpecial := filters.OnlyDevices[specialDevice]

	// sort/filter devices
	for _, v := range m {
		if shouldSkipMount(v, filters) {
			continue
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
