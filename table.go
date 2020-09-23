package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/muesli/termenv"
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

func printTable(title string, m []Mount) {
	tab := table.NewWriter()
	tab.SetAllowedRowLength(int(*width))
	tab.SetOutputMirror(os.Stdout)
	tab.SetStyle(table.StyleRounded)

	barWidth := 20.0
	switch {
	case *width < 100:
		barWidth = 0
	case *width < 120:
		barWidth = 10
	}
	cols := *width -
		7*3 - // size columns
		uint(barWidth) - // bar
		6 - // percentage
		2 - // frame
		7*2 - // spacers
		7 // seperators
	if barWidth > 0 {
		cols -= 2
	}

	tab.SetColumnConfigs([]table.ColumnConfig{
		{Number: 1, WidthMax: int(float64(cols) * 0.4)},
		{Number: 2, Align: text.AlignRight, AlignHeader: text.AlignRight},
		{Number: 3, Align: text.AlignRight, AlignHeader: text.AlignRight},
		{Number: 4, Align: text.AlignRight, AlignHeader: text.AlignRight},
		{Number: 5, AlignHeader: text.AlignCenter},
		{Number: 6, WidthMax: int(float64(cols) * 0.2)},
		{Number: 7, WidthMax: int(float64(cols) * 0.4)},
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
		// skip loopback devices
		if *hideLoopback && strings.HasPrefix(v.Device, "/dev/loop") {
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
		var usage = float64(v.Used) / float64(v.Total)
		usepct := termenv.String()
		if v.Total > 0 {
			if barWidth > 0 {
				usepct = termenv.String(fmt.Sprintf("[%s%s] %5.1f%%",
					strings.Repeat("#", int(usage*barWidth)),
					strings.Repeat(".", int(barWidth)-int(usage*barWidth)),
					usage*100,
				))
			} else {
				usepct = termenv.String(fmt.Sprintf("%5.1f%%", usage*100))
			}
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
