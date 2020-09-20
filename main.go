package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
)

var (
	all         = flag.Bool("all", false, "show all devices")
	hideLocal   = flag.Bool("hide-local", false, "hides local devices")
	hideNetwork = flag.Bool("hide-network", false, "hides network devices")
	hideBinds   = flag.Bool("hide-binds", true, "hides bind mounts")
	hideVirtual = flag.Bool("hide-virtual", true, "hides virtual devices")
)

func printTable(m []Mount) {
	if len(m) == 0 {
		return
	}

	tab := table.NewWriter()
	tab.SetOutputMirror(os.Stdout)
	tab.SetStyle(table.StyleRounded)
	tab.AppendHeader(table.Row{"Filesystem", "Size", "Used", "Avail", "Use%", "Type", "Mounted on"})

	for _, v := range m {
		// fmt.Println(v)
		// fmt.Println(stat.Type)
		// fmt.Println(fsTypeMap[stat.Type])

		// skip autofs
		if v.Fstype == "autofs" {
			continue
		}

		// skip bind-mounts
		if *hideBinds && strings.Contains(v.Opts, "bind") {
			continue
		}
		if v.Stat.Blocks == 0 && !*all {
			continue
		}

		var usepct string
		if v.Total > 0 {
			usepct = fmt.Sprintf("%.1f%%", float64(v.Used)/float64(v.Total)*100)
		}

		tab.AppendRow([]interface{}{
			v.Device, // filesystem
			sizeToString(v.Total),
			sizeToString(v.Used),
			sizeToString(v.Free),
			usepct,
			fsTypeMap[v.Stat.Type], // type
			v.Mountpoint,           // mounted on
		})
	}

	tab.Render()
}

// sizeToString prettifies sizes.
func sizeToString(size uint64) (str string) {
	b := float64(size)

	switch {
	case size >= 1<<60:
		str = fmt.Sprintf("%.2f EiB", b/(1<<60))
	case size >= 1<<50:
		str = fmt.Sprintf("%.2f PiB", b/(1<<50))
	case size >= 1<<40:
		str = fmt.Sprintf("%.2f TiB", b/(1<<40))
	case size >= 1<<30:
		str = fmt.Sprintf("%.2f GiB", b/(1<<30))
	case size >= 1<<20:
		str = fmt.Sprintf("%.2f MiB", b/(1<<20))
	case size >= 1<<10:
		str = fmt.Sprintf("%.2f KiB", b/(1<<10))
	default:
		str = fmt.Sprintf("%dB", size)
	}

	return
}

func main() {
	flag.Parse()

	m, err := mounts()
	if err != nil {
		panic(err)
	}

	var local []Mount
	var network []Mount
	var special []Mount
	for _, v := range m {
		if isLocalFs(v.Stat) {
			local = append(local, v)
		}
		if isNetworkFs(v.Stat) {
			network = append(network, v)
		}
		if isVirtualFs(v.Stat) {
			special = append(special, v)
		}
	}

	if !*hideLocal || *all {
		printTable(local)
	}
	if !*hideNetwork || *all {
		printTable(network)
	}
	if !*hideVirtual || *all {
		printTable(special)
	}
}
