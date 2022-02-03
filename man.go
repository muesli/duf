//go:build mango
// +build mango

package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/muesli/mango"
	"github.com/muesli/mango/mflag"
	"github.com/muesli/roff"
)

func init() {
	usage := `You can simply start duf without any command-line arguments:

  $ duf

If you supply arguments, duf will only list specific devices & mount points:

  $ duf /home /some/file

If you want to list everything (including pseudo, duplicate, inaccessible file systems):

  $ duf --all

You can show and hide specific tables:

  $ duf --only local,network,fuse,special,loops,binds
  $ duf --hide local,network,fuse,special,loops,binds

You can also show and hide specific filesystems:

  $ duf --only-fs tmpfs,vfat
  $ duf --hide-fs tmpfs,vfat

...or specific mount points:

  $ duf --only-mp /,/home,/dev
  $ duf --hide-mp /,/home,/dev

Wildcards inside quotes work:

  $ duf --only-mp '/sys/*,/dev/*'

Sort the output:

  $ duf --sort size

Valid keys are: mountpoint, size, used, avail, usage, inodes, inodes_used, inodes_avail, inodes_usage, type, filesystem.

Show or hide specific columns:

  $ duf --output mountpoint,size,usage

Valid keys are: mountpoint, size, used, avail, usage, inodes, inodes_used, inodes_avail, inodes_usage, type, filesystem.

List inode information instead of block usage:

  $ duf --inodes

If duf doesn't detect your terminal's colors correctly, you can set a theme:

  $ duf --theme light

duf highlights the availability & usage columns in red, green, or yellow, depending on how much space is still available. You can set your own thresholds:

  $ duf --avail-threshold="10G,1G"
  $ duf --usage-threshold="0.5,0.9"

If you prefer your output as JSON:

  $ duf --json
`

	manPage := mango.NewManPage(1, "duf", "Disk Usage/Free Utility").
		WithLongDescription("Simple Disk Usage/Free Utility.\n"+
			"Features:\n"+
			"* User-friendly, colorful output.\n"+
			"* Adjusts to your terminal's theme & width.\n"+
			"* Sort the results according to your needs.\n"+
			"* Groups & filters devices.\n"+
			"* Can conveniently output JSON.").
		WithSection("Usage", usage).
		WithSection("Notes", "Portions of duf's code are copied and modified from https://github.com/shirou/gopsutil.\n"+
			"gopsutil was written by WAKAYAMA Shirou and is distributed under BSD-3-Clause.").
		WithSection("Authors", "duf was written by Christian Muehlhaeuser <https://github.com/muesli/duf>").
		WithSection("Copyright", "Copyright (C) 2020-2022 Christian Muehlhaeuser <https://github.com/muesli>\n"+
			"Released under MIT license.")

	flag.VisitAll(mflag.FlagVisitor(manPage))
	fmt.Println(manPage.Build(roff.NewDocument()))
	os.Exit(0)
}
