package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/muesli/termenv"
)

var (
	term = termenv.ColorProfile()

	all = flag.Bool("all", false, "show all devices")

	hideLocal   = flag.Bool("hide-local", false, "hides local devices")
	hideNetwork = flag.Bool("hide-network", false, "hides network devices")
	hideFuse    = flag.Bool("hide-fuse", false, "hides fuse devices")
	hideSpecial = flag.Bool("hide-special", false, "hides special devices")
	hideBinds   = flag.Bool("hide-binds", true, "hides bind mounts")

	jsonOutput = flag.Bool("json", false, "output all devices in JSON format")
)

var (
	colorRed     = term.Color("#E88388")
	colorYellow  = term.Color("#DBAB79")
	colorGreen   = term.Color("#A8CC8C")
	colorBlue    = term.Color("#71BEF2")
	colorGray    = term.Color("#B9BFCA")
	colorMagenta = term.Color("#D290E4")
	colorCyan    = term.Color("#66C2CD")
)

func printTable(title string, m []Mount) {
	tab := table.NewWriter()
	tab.SetOutputMirror(os.Stdout)
	tab.SetStyle(table.StyleRounded)

	tab.SetColumnConfigs([]table.ColumnConfig{
		{Number: 2, Align: text.AlignRight, AlignHeader: text.AlignRight},
		{Number: 3, Align: text.AlignRight, AlignHeader: text.AlignRight},
		{Number: 4, Align: text.AlignRight, AlignHeader: text.AlignRight},
		{Number: 5, AlignHeader: text.AlignCenter},
	})
	tab.AppendHeader(table.Row{"Mounted on", "Size", "Used", "Avail", "Use%", "Type", "Filesystem"})

	for _, v := range m {
		// spew.Dump(v)

		// skip autofs
		if v.Fstype == "autofs" {
			continue
		}
		// skip bind-mounts
		if *hideBinds && strings.Contains(v.Opts, "bind") {
			continue
		}
		// skip special devices
		if v.Stat.Blocks == 0 && !*all {
			continue
		}
		// skip zero size devices
		if v.Stat.Bsize == 0 && !*all {
			continue
		}

		// free space
		var free = termenv.String(sizeToString(v.Free))
		switch {
		case v.Free < 1<<30:
			free = free.Foreground(colorRed)
		case v.Free < 10*1<<30:
			free = free.Foreground(colorYellow)
		default:
			free = free.Foreground(colorGreen)
		}

		// render progress-bar
		barWidth := 20.0
		var usage = float64(v.Used) / float64(v.Total)
		usepct := termenv.String()
		if v.Total > 0 {
			usepct = termenv.String(fmt.Sprintf("[%s%s] %.1f%%",
				strings.Repeat("#", int(usage*barWidth)),
				strings.Repeat(".", int(barWidth)-int(usage*barWidth)),
				usage*100,
			))
		}

		// apply color to progress-bar
		switch {
		case usage >= 0.9:
			usepct = usepct.Foreground(colorRed)
		case usage >= 0.5:
			usepct = usepct.Foreground(colorYellow)
		default:
			usepct = usepct.Foreground(colorGreen)
		}

		tab.AppendRow([]interface{}{
			termenv.String(v.Mountpoint).Foreground(colorBlue), // mounted on
			sizeToString(v.Total),                              // size
			sizeToString(v.Used),                               // used
			free,                                               // avail
			usepct,                                             // use%
			termenv.String(v.Type).Foreground(colorGray),   // type
			termenv.String(v.Device).Foreground(colorGray), // filesystem
		})
	}

	if tab.Length() == 0 {
		return
	}

	suffix := "device"
	if tab.Length() > 1 {
		suffix = "devices"
	}
	tab.SetTitle("%d %s %s", tab.Length(), title, suffix)
	tab.Render()
}

// sizeToString prettifies sizes.
func sizeToString(size uint64) (str string) {
	b := float64(size)

	switch {
	case size >= 1<<60:
		str = fmt.Sprintf("%.1fE", b/(1<<60))
	case size >= 1<<50:
		str = fmt.Sprintf("%.1fP", b/(1<<50))
	case size >= 1<<40:
		str = fmt.Sprintf("%.1fT", b/(1<<40))
	case size >= 1<<30:
		str = fmt.Sprintf("%.1fG", b/(1<<30))
	case size >= 1<<20:
		str = fmt.Sprintf("%.1fM", b/(1<<20))
	case size >= 1<<10:
		str = fmt.Sprintf("%.1fK", b/(1<<10))
	default:
		str = fmt.Sprintf("%dB", size)
	}

	return
}

func renderTables(m []Mount) {
	var local, network, fuse, special []Mount
	for _, v := range m {
		if isNetworkFs(v.Stat) {
			network = append(network, v)
			continue
		}
		if isFuseFs(v.Stat) {
			fuse = append(fuse, v)
			continue
		}
		if isSpecialFs(v.Stat) {
			special = append(special, v)
			continue
		}

		local = append(local, v)
	}

	if !*hideLocal || *all {
		printTable("local", local)
	}
	if !*hideNetwork || *all {
		printTable("network", network)
	}
	if !*hideFuse || *all {
		printTable("FUSE", fuse)
	}
	if !*hideSpecial || *all {
		printTable("special", special)
	}
}

func renderJSON(m []Mount) error {
	output, err := json.MarshalIndent(m, "", " ")
	if err != nil {
		return fmt.Errorf("error formatting the json output: %s", err)
	}

	fmt.Println(string(output))
	return nil
}

func main() {
	flag.Parse()

	m, warnings, err := mounts()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	for _, warning := range warnings {
		fmt.Fprintln(os.Stderr, warning)
	}

	if *jsonOutput {
		err := renderJSON(m)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}

		return
	}

	renderTables(m)
}
