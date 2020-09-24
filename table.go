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
	columns = table.Row{"Mounted on", "Size", "Used", "Avail", "Use%", "Type", "Filesystem"}

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

// stringToColumn converts a column name to its index.
func stringToColumn(s string) (int, error) {
	switch strings.ToLower(s) {
	case "mountpoint":
		return 1, nil
	case "size":
		return 8, nil
	case "used":
		return 9, nil
	case "avail":
		return 10, nil
	case "usage":
		return 11, nil
	case "type":
		return 6, nil
	case "filesystem":
		return 7, nil

	default:
		return 0, fmt.Errorf("unknown column identifier: %s", s)
	}
}

func sizeTransformer(val interface{}) string {
	return sizeToString(val.(uint64))
}

func spaceTransformer(val interface{}) string {
	free := val.(uint64)

	var s = termenv.String(sizeToString(free))
	switch {
	case free < 1<<30:
		s = s.Foreground(colorRed)
	case free < 10*1<<30:
		s = s.Foreground(colorYellow)
	default:
		s = s.Foreground(colorGreen)
	}

	return s.String()
}

func barTransformer(val interface{}) string {
	barWidth := 20
	switch {
	case *width < 100:
		barWidth = 0
	case *width < 120:
		barWidth = 10
	}

	usage := val.(float64)
	s := termenv.String()
	if usage > 0 {
		if barWidth > 0 {
			s = termenv.String(fmt.Sprintf("[%s%s] %5.1f%%",
				strings.Repeat("#", int(usage*float64(barWidth))),
				strings.Repeat(".", barWidth-int(usage*float64(barWidth))),
				usage*100,
			))
		} else {
			s = termenv.String(fmt.Sprintf("%5.1f%%", usage*100))
		}
	}

	// apply color to progress-bar
	switch {
	case usage >= 0.9:
		s = s.Foreground(colorRed)
	case usage >= 0.5:
		s = s.Foreground(colorYellow)
	default:
		s = s.Foreground(colorGreen)
	}

	return s.String()
}

func printTable(title string, m []Mount, sortBy int) {
	tab := table.NewWriter()
	tab.SetAllowedRowLength(int(*width))
	tab.SetOutputMirror(os.Stdout)
	tab.SetStyle(table.StyleRounded)
	// tab.Style().Options.SeparateColumns = false

	barWidth := 20
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
		{Number: 2, Transformer: sizeTransformer, Align: text.AlignRight, AlignHeader: text.AlignRight},
		{Number: 3, Transformer: sizeTransformer, Align: text.AlignRight, AlignHeader: text.AlignRight},
		{Number: 4, Transformer: spaceTransformer, Align: text.AlignRight, AlignHeader: text.AlignRight},
		{Number: 5, Transformer: barTransformer, AlignHeader: text.AlignCenter},
		{Number: 6, WidthMax: int(float64(cols) * 0.2)},
		{Number: 7, WidthMax: int(float64(cols) * 0.4)},
		{Number: 8, Hidden: true},  // sortBy helper for size
		{Number: 9, Hidden: true},  // sortBy helper for used
		{Number: 10, Hidden: true}, // sortBy helper for avail
		{Number: 11, Hidden: true}, // sortBy helper for usage
	})

	tab.AppendHeader(columns)

	for _, v := range m {
		// spew.Dump(v)

		// render progress-bar
		var usage float64
		if v.Total > 0 {
			usage = float64(v.Used) / float64(v.Total)
		}

		tab.AppendRow([]interface{}{
			termenv.String(v.Mountpoint).Foreground(colorBlue), // mounted on
			v.Total, // size
			v.Used,  // used
			v.Free,  // avail
			usage,   // use%
			termenv.String(v.Fstype).Foreground(colorGray), // type
			termenv.String(v.Device).Foreground(colorGray), // filesystem
			v.Total, // size sorting helper
			v.Used,  // used sorting helper
			v.Free,  // avail sorting helper
			usage,   // use% sorting helper
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

	//tab.AppendFooter(table.Row{fmt.Sprintf("%d %s", tab.Length(), title)})
	sortMode := table.Asc
	if sortBy >= 8 && sortBy <= 11 {
		sortMode = table.AscNumeric
	}

	tab.SortBy([]table.SortBy{{Number: sortBy, Mode: sortMode}})
	tab.Render()
}
