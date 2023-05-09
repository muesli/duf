package main

import (
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/muesli/termenv"
)

// TableOptions contains all options for the table.
type TableOptions struct {
	Columns []int
	SortBy  int
	Style   table.Style
}

// Column defines a column.
type Column struct {
	ID        string
	Name      string
	SortIndex int
	Width     int
}

// "Mounted on", "Size", "Used", "Avail", "Use%", "Inodes", "IUsed", "IAvail", "IUse%", "Type", "Filesystem"
// mountpoint, size, used, avail, usage, inodes, inodes_used, inodes_avail, inodes_usage, type, filesystem
var columns = []Column{
	{ID: "mountpoint", Name: "Mounted on", SortIndex: 1},
	{ID: "size", Name: "Size", SortIndex: 12, Width: 7},
	{ID: "used", Name: "Used", SortIndex: 13, Width: 7},
	{ID: "avail", Name: "Avail", SortIndex: 14, Width: 7},
	{ID: "usage", Name: "Use%", SortIndex: 15, Width: 6},
	{ID: "inodes", Name: "Inodes", SortIndex: 16, Width: 7},
	{ID: "inodes_used", Name: "IUsed", SortIndex: 17, Width: 7},
	{ID: "inodes_avail", Name: "IAvail", SortIndex: 18, Width: 7},
	{ID: "inodes_usage", Name: "IUse%", SortIndex: 19, Width: 6},
	{ID: "type", Name: "Type", SortIndex: 10},
	{ID: "filesystem", Name: "Filesystem", SortIndex: 11},
}

// printTable prints an individual table of mounts.
func printTable(title string, m []Mount, opts TableOptions) {
	tab := table.NewWriter()
	tab.SetAllowedRowLength(int(*width))
	tab.SetOutputMirror(os.Stdout)
	tab.Style().Options.SeparateColumns = true
	tab.SetStyle(opts.Style)

	if barWidth() > 0 {
		columns[4].Width = barWidth() + 7
		columns[8].Width = barWidth() + 7
	}
	twidth := tableWidth(opts.Columns, tab.Style().Options.SeparateColumns)

	tab.SetColumnConfigs([]table.ColumnConfig{
		{Number: 1, Hidden: !inColumns(opts.Columns, 1), WidthMax: int(float64(twidth) * 0.4)},
		{Number: 2, Hidden: !inColumns(opts.Columns, 2), Transformer: sizeTransformer, Align: text.AlignRight, AlignHeader: text.AlignRight},
		{Number: 3, Hidden: !inColumns(opts.Columns, 3), Transformer: sizeTransformer, Align: text.AlignRight, AlignHeader: text.AlignRight},
		{Number: 4, Hidden: !inColumns(opts.Columns, 4), Transformer: spaceTransformer, Align: text.AlignRight, AlignHeader: text.AlignRight},
		{Number: 5, Hidden: !inColumns(opts.Columns, 5), Transformer: barTransformer, AlignHeader: text.AlignCenter},
		{Number: 6, Hidden: !inColumns(opts.Columns, 6), Align: text.AlignRight, AlignHeader: text.AlignRight},
		{Number: 7, Hidden: !inColumns(opts.Columns, 7), Align: text.AlignRight, AlignHeader: text.AlignRight},
		{Number: 8, Hidden: !inColumns(opts.Columns, 8), Align: text.AlignRight, AlignHeader: text.AlignRight},
		{Number: 9, Hidden: !inColumns(opts.Columns, 9), Transformer: barTransformer, AlignHeader: text.AlignCenter},
		{Number: 10, Hidden: !inColumns(opts.Columns, 10), WidthMax: int(float64(twidth) * 0.2)},
		{Number: 11, Hidden: !inColumns(opts.Columns, 11), WidthMax: int(float64(twidth) * 0.4)},
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
			termenv.String(v.Mountpoint).Foreground(theme.colorBlue), // mounted on
			v.Total,      // size
			v.Used,       // used
			v.Free,       // avail
			usage,        // use%
			v.Inodes,     // inodes
			v.InodesUsed, // inodes used
			v.InodesFree, // inodes avail
			inodeUsage,   // inodes use%
			termenv.String(v.Fstype).Foreground(theme.colorGray), // type
			termenv.String(v.Device).Foreground(theme.colorGray), // filesystem
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

	// tab.AppendFooter(table.Row{fmt.Sprintf("%d %s", tab.Length(), title)})
	sortMode := table.Asc
	if opts.SortBy >= 12 {
		sortMode = table.AscNumeric
	}

	tab.SortBy([]table.SortBy{{Number: opts.SortBy, Mode: sortMode}})
	tab.Render()
}

// sizeTransformer makes a size human-readable.
func sizeTransformer(val interface{}) string {
	return sizeToString(val.(uint64), *prefixSI)
}

// spaceTransformer makes a size human-readable and applies a color coding.
func spaceTransformer(val interface{}) string {
	free := val.(uint64)

	s := termenv.String(sizeToString(free, *prefixSI))
	redAvail, _ := stringToSize(strings.Split(*availThreshold, ",")[1], *prefixSI)
	yellowAvail, _ := stringToSize(strings.Split(*availThreshold, ",")[0], *prefixSI)
	switch {
	case free < redAvail:
		s = s.Foreground(theme.colorRed)
	case free < yellowAvail:
		s = s.Foreground(theme.colorYellow)
	default:
		s = s.Foreground(theme.colorGreen)
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
	redUsage, _ := strconv.ParseFloat(strings.Split(*usageThreshold, ",")[1], 64)
	yellowUsage, _ := strconv.ParseFloat(strings.Split(*usageThreshold, ",")[0], 64)
	switch {
	case usage >= redUsage:
		s = s.Foreground(theme.colorRed)
	case usage >= yellowUsage:
		s = s.Foreground(theme.colorYellow)
	default:
		s = s.Foreground(theme.colorGreen)
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

func sizeToString(bytes uint64, SI bool) string {
	size := float64(bytes)

	unit := float64(1024)
	prefixes := []string{"Ki", "Mi", "Gi", "Ti", "Pi", "Ei", "Zi", "Yi"}

	if SI {
		unit = float64(1000)
		prefixes = []string{"k", "M", "G", "T", "P", "E", "Z", "Y"}
	}

	if size < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	exp := int(math.Log(size) / math.Log(unit))
	prefix := prefixes[exp-1]
	value := size / math.Pow(unit, float64(exp))

	return fmt.Sprintf("%.1f%sB", value, prefix)
}

// stringToSize transforms an SI size into a number.
func stringToSize(s string, SI bool) (uint64, error) {
	var prefix string
	var num float64
	if _, err := fmt.Sscanf(s, "%f%s", &num, &prefix); err != nil {
		return 0, err
	}

	if prefix == "" {
		return uint64(num), nil
	}

	unit := float64(1024)
	prefixes := []string{"", "k", "M", "G", "T", "P", "E", "Z", "Y"}
	if SI {
		unit = float64(1000)
		prefixes[1] = "K"
	}

	var exp int
	for i, p := range prefixes {
		if p == prefix {
			exp = i
			break
		}
	}

	if exp == 0 {
		return 0, fmt.Errorf("prefix '%s' not allowed, valid prefixes are k, M, G, T, P, E, Z, Y and K with SI format", prefix)
	}

	return uint64(num * math.Pow(unit, float64(exp))), nil
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
