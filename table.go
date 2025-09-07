package main

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/mattn/go-runewidth"
	"github.com/muesli/termenv"
)

// TableOptions contains all options for the table.
type TableOptions struct {
	Columns   []int
	SortBy    int
	Style     table.Style
	StyleName string
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

// initializeTable sets up the table writer with initial configurations.
func initializeTable(tab table.Writer, opts TableOptions) {
	tab.SetAllowedRowLength(int(*width))
	tab.SetOutputMirror(os.Stdout)
	tab.Style().Options.SeparateColumns = true
	tab.SetStyle(opts.Style)
}

// appendHeaders adds the header row to the table.
func appendHeaders(tab table.Writer) {
	headers := table.Row{}
	for _, v := range columns {
		headers = append(headers, v.Name)
	}
	tab.AppendHeader(headers)
}

// appendRows adds data rows to the table for each mount.
func appendRows(tab table.Writer, m []Mount) {
	for _, v := range m {
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
}

// computeMaxContentWidths calculates the maximum content width for each visible column.
func computeMaxContentWidths(m []Mount, opts TableOptions) map[int]int {
	visibleCols := append([]int{}, opts.Columns...)
	maxColContent := map[int]int{}
	// Seed with headers
	for _, ci := range visibleCols {
		maxColContent[ci] = runewidth.StringWidth(columns[ci-1].Name)
	}
	for _, v := range m {
		if inColumns(opts.Columns, 1) {
			if w := runewidth.StringWidth(v.Mountpoint); w > maxColContent[1] {
				maxColContent[1] = w
			}
		}
		if inColumns(opts.Columns, 2) {
			if w := runewidth.StringWidth(sizeToString(v.Total)); w > maxColContent[2] {
				maxColContent[2] = w
			}
		}
		if inColumns(opts.Columns, 3) {
			if w := runewidth.StringWidth(sizeToString(v.Used)); w > maxColContent[3] {
				maxColContent[3] = w
			}
		}
		if inColumns(opts.Columns, 4) {
			if w := runewidth.StringWidth(sizeToString(v.Free)); w > maxColContent[4] {
				maxColContent[4] = w
			}
		}
		if inColumns(opts.Columns, 5) {
			var usage float64
			if v.Total > 0 {
				usage = float64(v.Used) / float64(v.Total)
				if usage > 1.0 {
					usage = 1.0
				}
			}
			percentStr := fmt.Sprintf("%.1f%%", usage*100)
			if w := runewidth.StringWidth(percentStr); w > maxColContent[5] {
				maxColContent[5] = w
			}
		}
		if inColumns(opts.Columns, 6) {
			if w := runewidth.StringWidth(strconv.FormatUint(v.Inodes, 10)); w > maxColContent[6] {
				maxColContent[6] = w
			}
		}
		if inColumns(opts.Columns, 7) {
			if w := runewidth.StringWidth(strconv.FormatUint(v.InodesUsed, 10)); w > maxColContent[7] {
				maxColContent[7] = w
			}
		}
		if inColumns(opts.Columns, 8) {
			if w := runewidth.StringWidth(strconv.FormatUint(v.InodesFree, 10)); w > maxColContent[8] {
				maxColContent[8] = w
			}
		}
		if inColumns(opts.Columns, 9) {
			var usage float64
			if v.Inodes > 0 {
				usage = float64(v.InodesUsed) / float64(v.Inodes)
				if usage > 1.0 {
					usage = 1.0
				}
			}
			percentStr := fmt.Sprintf("%.1f%%", usage*100)
			if w := runewidth.StringWidth(percentStr); w > maxColContent[9] {
				maxColContent[9] = w
			}
		}
		if inColumns(opts.Columns, 10) {
			if w := runewidth.StringWidth(v.Fstype); w > maxColContent[10] {
				maxColContent[10] = w
			}
		}
		if inColumns(opts.Columns, 11) {
			if w := runewidth.StringWidth(v.Device); w > maxColContent[11] {
				maxColContent[11] = w
			}
		}
	}
	return maxColContent
}

// computeAssignedWidths computes the assigned widths for dynamic columns (1, 10, 11).
func computeAssignedWidths(maxColContent map[int]int, opts TableOptions) (map[int]int, int) {
	visibleCols := append([]int{}, opts.Columns...)
	nVis := len(visibleCols)

	// Non-content overhead
	sepWidth := 1
	paddingPerCol := 2
	overhead := (nVis+1)*sepWidth + nVis*paddingPerCol
	totalAllowed := int(*width)

	// Determine targets and their need
	targets := []int{}
	weights := map[int]float64{1: 0.4, 10: 0.2, 11: 0.4}
	weightSum := 0.0
	for _, t := range []int{1, 10, 11} {
		if inColumns(opts.Columns, t) {
			targets = append(targets, t)
			weightSum += weights[t]
		}
	}

	// Sum fixed widths of non-target visible columns
	fixedContentWidth := 0
	for _, ci := range visibleCols {
		if ci == 1 || ci == 10 || ci == 11 {
			continue
		}
		fixedContentWidth += maxColContent[ci]
	}

	availableContent := totalAllowed - overhead - fixedContentWidth
	if availableContent < 0 {
		availableContent = 0
	}

	// Cap target allocations by their max content need
	assigned := map[int]int{}
	used := 0
	if availableContent > 0 && len(targets) > 0 {
		for _, t := range targets {
			share := int(float64(availableContent) * (weights[t] / weightSum))
			if share > maxColContent[t] {
				share = maxColContent[t]
			}
			assigned[t] = share
			used += share
		}
		// remainder distribution
		remainder := availableContent - used
		for remainder > 0 {
			bestCol := 0
			bestNeed := 0
			for _, t := range targets {
				need := maxColContent[t] - assigned[t]
				if need > bestNeed {
					bestNeed = need
					bestCol = t
				}
			}
			if bestNeed <= 0 {
				break
			}
			take := remainder
			if take > bestNeed {
				take = bestNeed
			}
			assigned[bestCol] += take
			remainder -= take
		}
	}

	// Calculate final slack
	predictedTotal := overhead + fixedContentWidth
	for _, t := range targets {
		predictedTotal += assigned[t]
	}
	slack := totalAllowed - predictedTotal
	return assigned, slack
}

// setColumnConfigs configures the columns for the table.
func setColumnConfigs(tab table.Writer, maxColContent map[int]int, assigned map[int]int, opts TableOptions, barTransformerFunc func(interface{}) string) {
	cfgs := []table.ColumnConfig{
		{Number: 1, Hidden: !inColumns(opts.Columns, 1), WidthMax: assigned[1]},
		{Number: 2, Hidden: !inColumns(opts.Columns, 2), Transformer: sizeTransformer, Align: text.AlignRight, AlignHeader: text.AlignRight, WidthMax: maxColContent[2]},
		{Number: 3, Hidden: !inColumns(opts.Columns, 3), Transformer: sizeTransformer, Align: text.AlignRight, AlignHeader: text.AlignRight, WidthMax: maxColContent[3]},
		{Number: 4, Hidden: !inColumns(opts.Columns, 4), Transformer: spaceTransformer, Align: text.AlignRight, AlignHeader: text.AlignRight, WidthMax: maxColContent[4]},
		{Number: 5, Hidden: !inColumns(opts.Columns, 5), Transformer: barTransformerFunc, AlignHeader: text.AlignCenter, WidthMax: maxColContent[5]},
		{Number: 6, Hidden: !inColumns(opts.Columns, 6), Align: text.AlignRight, AlignHeader: text.AlignRight, WidthMax: maxColContent[6]},
		{Number: 7, Hidden: !inColumns(opts.Columns, 7), Align: text.AlignRight, AlignHeader: text.AlignRight, WidthMax: maxColContent[7]},
		{Number: 8, Hidden: !inColumns(opts.Columns, 8), Align: text.AlignRight, AlignHeader: text.AlignRight, WidthMax: maxColContent[8]},
		{Number: 9, Hidden: !inColumns(opts.Columns, 9), Transformer: barTransformerFunc, AlignHeader: text.AlignCenter, WidthMax: maxColContent[9]},
		{Number: 10, Hidden: !inColumns(opts.Columns, 10), WidthMax: assigned[10]},
		{Number: 11, Hidden: !inColumns(opts.Columns, 11), WidthMax: assigned[11]},
		{Number: 12, Hidden: true}, // sortBy helper for size
		{Number: 13, Hidden: true}, // sortBy helper for used
		{Number: 14, Hidden: true}, // sortBy helper for avail
		{Number: 15, Hidden: true}, // sortBy helper for usage
		{Number: 16, Hidden: true}, // sortBy helper for inodes size
		{Number: 17, Hidden: true}, // sortBy helper for inodes used
		{Number: 18, Hidden: true}, // sortBy helper for inodes avail
		{Number: 19, Hidden: true}, // sortBy helper for inodes usage
	}
	tab.SetColumnConfigs(cfgs)
}

// printTable prints an individual table of mounts.
func printTable(title string, m []Mount, opts TableOptions) {
	tab := table.NewWriter()
	initializeTable(tab, opts)
	appendHeaders(tab)
	appendRows(tab, m)

	if tab.Length() == 0 {
		return
	}

	maxColContent := computeMaxContentWidths(m, opts)
	assigned, slack := computeAssignedWidths(maxColContent, opts)

	origPercentWidth5 := maxColContent[5]
	origPercentWidth9 := maxColContent[9]
	percentWidth := origPercentWidth5
	if origPercentWidth9 > percentWidth {
		percentWidth = origPercentWidth9
	}

	barWidth := 0
	numBars := 0
	if inColumns(opts.Columns, 5) {
		numBars++
	}
	if inColumns(opts.Columns, 9) {
		numBars++
	}
	if numBars > 0 && slack >= 6 {
		// Each bar consumes: barWidth + 1 (for space)
		// So for numBars, total consumption is: numBars * (barWidth + 1)
		maxBarWidth := min((slack/numBars)-1, 20)

		if maxBarWidth > 0 {
			barWidth = maxBarWidth
			if inColumns(opts.Columns, 5) {
				maxColContent[5] = barWidth + 1 + percentWidth
			}
			if inColumns(opts.Columns, 9) {
				maxColContent[9] = barWidth + 1 + percentWidth
			}
		}
	}

	// Define barTransformerFunc
	barTransformerFunc := func(val interface{}) string {
		usage := val.(float64)
		s := termenv.String()
		if usage > 0 {
			if barWidth > 0 {
				bw := barWidth
				var filledChar, halfChar, emptyChar string
				if opts.StyleName == "unicode" {
					filledChar = "█"
					halfChar = "▌"
					emptyChar = " "
				} else {
					bw -= 2
					filledChar = "#"
					halfChar = "#"
					emptyChar = "."
				}

				filled := int(usage * float64(bw))
				partial := usage*float64(bw) - float64(filled)
				empty := bw - filled

				var filledStr, emptyStr string
				filledStr = strings.Repeat(filledChar, filled)

				// If we have a sufficiently large partial, render a half block.
				if partial >= 0.5 {
					filledStr += halfChar
					empty--
				}

				if empty < 0 {
					empty = 0
				}
				emptyStr = strings.Repeat(emptyChar, empty)

				var format string
				if opts.StyleName == "unicode" {
					format = "%s%s %*s"
				} else {
					format = "[%s%s] %*s"
				}

				// Apply colors
				redUsage, _ := strconv.ParseFloat(strings.Split(*usageThreshold, ",")[1], 64)
				yellowUsage, _ := strconv.ParseFloat(strings.Split(*usageThreshold, ",")[0], 64)

				var fgColor termenv.Color
				switch {
				case usage >= redUsage:
					fgColor = theme.colorRed
				case usage >= yellowUsage:
					fgColor = theme.colorYellow
				default:
					fgColor = theme.colorGreen
				}

				filledPart := termenv.String(filledStr).Foreground(fgColor)
				emptyPart := termenv.String(emptyStr)
				if opts.StyleName == "unicode" {
					// Add background to filled part to prevent black spaces in half blocks
					// Use a background color that complements the foreground
					var bgColor termenv.Color
					switch {
					case usage >= redUsage:
						bgColor = theme.colorBgRed
					case usage >= yellowUsage:
						bgColor = theme.colorBgYellow
					default:
						bgColor = theme.colorBgGreen
					}
					filledPart = filledPart.Background(bgColor).Foreground(fgColor)
					// Use a neutral background for empty areas
					emptyPart = emptyPart.Background(bgColor)
				}

				percentStr := fmt.Sprintf("%.1f%%", usage*100)
				s = termenv.String(fmt.Sprintf(format, filledPart, emptyPart, percentWidth, percentStr))
			} else {
				percentStr := fmt.Sprintf("%.1f%%", usage*100)
				s = termenv.String(fmt.Sprintf("%*s", percentWidth, percentStr))
			}
		}

		return s.String()
	}

	setColumnConfigs(tab, maxColContent, assigned, opts, barTransformerFunc)

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
	return sizeToString(val.(uint64))
}

// spaceTransformer makes a size human-readable and applies a color coding.
func spaceTransformer(val interface{}) string {
	free := val.(uint64)

	s := termenv.String(sizeToString(free))
	redAvail, _ := stringToSize(strings.Split(*availThreshold, ",")[1])
	yellowAvail, _ := stringToSize(strings.Split(*availThreshold, ",")[0])
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

// stringToSize transforms an SI size into a number.
func stringToSize(s string) (size uint64, err error) {
	regex := regexp.MustCompile(`^(\d+)([KMGTPE]?)$`)
	matches := regex.FindStringSubmatch(s)
	if len(matches) == 0 {
		return 0, fmt.Errorf("'%s' is not valid, must have integer with optional SI prefix", s)
	}

	num, err := strconv.ParseUint(matches[1], 10, 64)
	if err != nil {
		return 0, err
	}
	if matches[2] != "" {
		prefix := matches[2]
		switch prefix {
		case "K":
			size = num << 10
		case "M":
			size = num << 20
		case "G":
			size = num << 30
		case "T":
			size = num << 40
		case "P":
			size = num << 50
		case "E":
			size = num << 60
		default:
			err = fmt.Errorf("prefix '%s' not allowed, valid prefixes are K, M, G, T, P, E", prefix)
			return
		}
	} else {
		size = num
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
