package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/muesli/termenv"
)

type Column struct {
	ID        string
	Name      string
	SortIndex int
	Width     int
}

var (
	// "Mounted on", "Size", "Used", "Avail", "Use%", "Inodes", "Used", "Avail", "Use%", "Type", "Filesystem"
	// mountpoint, size, used, avail, usage, inodes, inodes_used, inodes_avail, inodes_usage, type, filesystem
	columns = []Column{
		{ID: "mountpoint", Name: "Mounted on", SortIndex: 1},
		{ID: "size", Name: "Size", SortIndex: 12, Width: 7},
		{ID: "used", Name: "Used", SortIndex: 13, Width: 7},
		{ID: "avail", Name: "Avail", SortIndex: 14, Width: 7},
		{ID: "usage", Name: "Use%", SortIndex: 15, Width: 6},
		{ID: "inodes", Name: "Inodes", SortIndex: 16, Width: 7},
		{ID: "inodes_used", Name: "Used", SortIndex: 17, Width: 7},
		{ID: "inodes_avail", Name: "Avail", SortIndex: 18, Width: 7},
		{ID: "inodes_usage", Name: "Use%", SortIndex: 19, Width: 6},
		{ID: "type", Name: "Type", SortIndex: 10},
		{ID: "filesystem", Name: "Filesystem", SortIndex: 11},
	}

	colorRed     = term.Color("#E88388")
	colorYellow  = term.Color("#DBAB79")
	colorGreen   = term.Color("#A8CC8C")
	colorBlue    = term.Color("#71BEF2")
	colorGray    = term.Color("#B9BFCA")
	colorMagenta = term.Color("#D290E4")
	colorCyan    = term.Color("#66C2CD")
)

// printTable prints an individual table of mounts.
func printTable(title string, m []Mount, sortBy int, cols []int) {
	tab := table.NewWriter()
	tab.SetAllowedRowLength(int(*width))
	tab.SetOutputMirror(os.Stdout)
	tab.SetStyle(table.StyleRounded)
	tab.Style().Options.SeparateColumns = true

	if barWidth() > 0 {
		columns[4].Width = barWidth() + 7
		columns[8].Width = barWidth() + 7
	}
	twidth := tableWidth(cols, tab.Style().Options.SeparateColumns)

	tab.SetColumnConfigs([]table.ColumnConfig{
		{Number: 1, Hidden: !inColumns(cols, 1), WidthMax: int(float64(twidth) * 0.4)},
		{Number: 2, Hidden: !inColumns(cols, 2), Transformer: sizeTransformer, Align: text.AlignRight, AlignHeader: text.AlignRight},
		{Number: 3, Hidden: !inColumns(cols, 3), Transformer: sizeTransformer, Align: text.AlignRight, AlignHeader: text.AlignRight},
		{Number: 4, Hidden: !inColumns(cols, 4), Transformer: spaceTransformer, Align: text.AlignRight, AlignHeader: text.AlignRight},
		{Number: 5, Hidden: !inColumns(cols, 5), Transformer: barTransformer, AlignHeader: text.AlignCenter},
		{Number: 6, Hidden: !inColumns(cols, 6), Align: text.AlignRight, AlignHeader: text.AlignRight},
		{Number: 7, Hidden: !inColumns(cols, 7), Align: text.AlignRight, AlignHeader: text.AlignRight},
		{Number: 8, Hidden: !inColumns(cols, 8), Align: text.AlignRight, AlignHeader: text.AlignRight},
		{Number: 9, Hidden: !inColumns(cols, 9), Transformer: barTransformer, AlignHeader: text.AlignCenter},
		{Number: 10, Hidden: !inColumns(cols, 10), WidthMax: int(float64(twidth) * 0.2)},
		{Number: 11, Hidden: !inColumns(cols, 11), WidthMax: int(float64(twidth) * 0.4)},
		{Number: 12, Hidden: true}, // sortBy helper for size
		{Number: 13, Hidden: true}, // sortBy helper for used
		{Number: 14, Hidden: true}, // sortBy helper for avail
		{Number: 15, Hidden: true}, // sortBy helper for usage
		{Number: 16, Hidden: true}, // sortBy helper for inodes size
		{Number: 17, Hidden: true}, // sortBy helper for inodes used
		{Number: 18, Hidden: true}, // sortBy helper for inodes avail
		{Number: 19, Hidden: true}, // sortBy helper for inodes usage
	})

	headers := table.Row{}
	for _, v := range columns {
		headers = append(headers, v.Name)
	}
	tab.AppendHeader(headers)

	for _, v := range m {
		// spew.Dump(v)

		var usage, inodeUsage float64
		if v.Total > 0 {
			usage = float64(v.Used) / float64(v.Total)
			if usage > 1.0 {
				usage = 1.0
			}
		}
		if v.Inodes > 0 {
			inodeUsage = float64(v.InodesUsed) / float64(v.Inodes)
			if inodeUsage > 1.0 {
				inodeUsage = 1.0
			}
		}

		tab.AppendRow([]interface{}{
			termenv.String(v.Mountpoint).Foreground(colorBlue), // mounted on
			v.Total,      // size
			v.Used,       // used
			v.Free,       // avail
			usage,        // use%
			v.Inodes,     // inodes
			v.InodesUsed, // inodes used
			v.InodesFree, // inodes avail
			inodeUsage,   // inodes use%
			termenv.String(v.Fstype).Foreground(colorGray), // type
			termenv.String(v.Device).Foreground(colorGray), // filesystem
			v.Total,      // size sorting helper
			v.Used,       // used sorting helper
			v.Free,       // avail sorting helper
			usage,        // use% sorting helper
			v.Inodes,     // inodes sorting helper
			v.InodesUsed, // inodes used sorting helper
			v.InodesFree, // inodes avail sorting helper
			inodeUsage,   // inodes use% sorting helper
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
	if sortBy >= 12 {
		sortMode = table.AscNumeric
	}

	tab.SortBy([]table.SortBy{{Number: sortBy, Mode: sortMode}})
	tab.Render()
}

// sizeTransformer makes a size human-readable.
func sizeTransformer(val interface{}) string {
	return sizeToString(val.(uint64))
}

// spaceTransformer makes a size human-readable and applies a color coding.
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

// barTransformer transforms a percentage into a progress-bar.
func barTransformer(val interface{}) string {
	usage := val.(float64)
	s := termenv.String()
	if usage > 0 {
		if barWidth() > 0 {
			bw := barWidth() - 2
			s = termenv.String(fmt.Sprintf("[%s%s] %5.1f%%",
				strings.Repeat("#", int(usage*float64(bw))),
				strings.Repeat(".", bw-int(usage*float64(bw))),
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

// inColumns return true if the column with index i is in the slice of visible
// columns cols.
func inColumns(cols []int, i int) bool {
	for _, v := range cols {
		if v == i {
			return true
		}
	}

	return false
}

// barWidth returns the width of progress-bars for the given render width.
func barWidth() int {
	switch {
	case *width < 100:
		return 0
	case *width < 120:
		return 12
	default:
		return 22
	}
}

// tableWidth returns the required minimum table width for the given columns.
func tableWidth(cols []int, separators bool) int {
	var sw int
	if separators {
		sw = 1
	}

	twidth := int(*width)
	for i := 0; i < len(columns); i++ {
		if inColumns(cols, i+1) {
			twidth -= 2 + sw + columns[i].Width
		}
	}

	return twidth
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

// stringToColumn converts a column name to its index.
func stringToColumn(s string) (int, error) {
	s = strings.ToLower(s)

	for i, v := range columns {
		if v.ID == s {
			return i + 1, nil
		}
	}

	return 0, fmt.Errorf("unknown column: %s (valid: %s)", s, strings.Join(columnIDs(), ", "))
}

// stringToSortIndex converts a column name to its sort index.
func stringToSortIndex(s string) (int, error) {
	s = strings.ToLower(s)

	for _, v := range columns {
		if v.ID == s {
			return v.SortIndex, nil
		}
	}

	return 0, fmt.Errorf("unknown column: %s (valid: %s)", s, strings.Join(columnIDs(), ", "))
}

// columnsIDs returns a slice of all column IDs.
func columnIDs() []string {
	s := make([]string, len(columns))
	for i, v := range columns {
		s[i] = v.ID
	}

	return s
}
