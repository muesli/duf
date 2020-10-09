package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/muesli/termenv"
	"golang.org/x/crypto/ssh/terminal"
)

var (
	term  = termenv.ColorProfile()
	theme Theme

	all         = flag.Bool("all", false, "include pseudo, duplicate, inaccessible file systems")
	hideLocal   = flag.Bool("hide-local", false, "hide local devices")
	hideNetwork = flag.Bool("hide-network", false, "hide network devices")
	hideFuse    = flag.Bool("hide-fuse", false, "hide fuse devices")
	hideSpecial = flag.Bool("hide-special", false, "hide special devices")
	hideLoops   = flag.Bool("hide-loops", true, "hide loop devices")
	hideBinds   = flag.Bool("hide-binds", true, "hide bind mounts")
	hideFs      = flag.String("hide-fs", "", "hide specific filesystems, separated with commas")

	output   = flag.String("output", "", "output fields: "+strings.Join(columnIDs(), ", "))
	sortBy   = flag.String("sort", "mountpoint", "sort output by: "+strings.Join(columnIDs(), ", "))
	width    = flag.Uint("width", 0, "max output width")
	themeOpt = flag.String("theme", defaultThemeName(), "color themes: dark, light")
	styleOpt = flag.String("style", defaultStyleName(), "style: unicode, ascii")

	inodes     = flag.Bool("inodes", false, "list inode information instead of block usage")
	jsonOutput = flag.Bool("json", false, "output all devices in JSON format")
)

// renderTables renders all tables.
func renderTables(m []Mount, columns []int, sortCol int, style table.Style) error {
	var local, network, fuse, special []Mount
	hideFsMap := parseHideFs(*hideFs)

	// sort/filter devices
	for _, v := range m {
		// skip hideFs
		if _, ok := hideFsMap[v.Fstype]; ok {
			continue
		}
		// skip autofs
		if v.Fstype == "autofs" {
			continue
		}
		// skip bind-mounts
		if *hideBinds && !*all && strings.Contains(v.Opts, "bind") {
			continue
		}
		// skip loop devices
		if *hideLoops && !*all && strings.HasPrefix(v.Device, "/dev/loop") {
			continue
		}
		// skip special devices
		if v.Blocks == 0 && !*all {
			continue
		}
		// skip zero size devices
		if v.BlockSize == 0 && !*all {
			continue
		}

		if isNetworkFs(v) {
			network = append(network, v)
			continue
		}
		if isFuseFs(v) {
			fuse = append(fuse, v)
			continue
		}
		if isSpecialFs(v) {
			special = append(special, v)
			continue
		}

		local = append(local, v)
	}

	// print tables
	if !*hideLocal || *all {
		printTable("local", local, sortCol, columns, style)
	}
	if !*hideNetwork || *all {
		printTable("network", network, sortCol, columns, style)
	}
	if !*hideFuse || *all {
		printTable("FUSE", fuse, sortCol, columns, style)
	}
	if !*hideSpecial || *all {
		printTable("special", special, sortCol, columns, style)
	}
	return nil
}

// renderJSON encodes the JSON output and prints it.
func renderJSON(m []Mount) error {
	output, err := json.MarshalIndent(m, "", " ")
	if err != nil {
		return fmt.Errorf("error formatting the json output: %s", err)
	}

	fmt.Println(string(output))
	return nil
}

// parseColumns parses the supplied output flag into a slice of column indices.
func parseColumns(cols string) ([]int, error) {
	var i []int

	s := strings.Split(cols, ",")
	for _, v := range s {
		v = strings.TrimSpace(v)
		if len(v) == 0 {
			continue
		}

		col, err := stringToColumn(v)
		if err != nil {
			return nil, err
		}

		i = append(i, col)
	}

	return i, nil
}

// parseStyle converts user-provided style option into a table.Style.
func parseStyle(styleOpt string) (table.Style, error) {
	switch styleOpt {
	case "unicode":
		return table.StyleRounded, nil
	case "ascii":
		return table.StyleDefault, nil
	default:
		return table.Style{}, fmt.Errorf("Unknown style option: %s", styleOpt)
	}
}

// parseHideFs parses the supplied hide-fs flag into a map of fs types which should be skipped.
func parseHideFs(hideFs string) map[string]struct{} {
	hideMap := make(map[string]struct{})
	for _, fs := range strings.Split(hideFs, ",") {
		fs = strings.TrimSpace(fs)
		if len(fs) == 0 {
			continue
		}
		hideMap[fs] = struct{}{}
	}
	return hideMap
}

func main() {
	flag.Parse()

	// validate flags
	var err error
	theme, err = loadTheme(*themeOpt)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	style, err := parseStyle(*styleOpt)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	columns, err := parseColumns(*output)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if len(columns) == 0 {
		// no columns supplied, use defaults
		if *inodes {
			columns = []int{1, 6, 7, 8, 9, 10, 11}
		} else {
			columns = []int{1, 2, 3, 4, 5, 10, 11}
		}
	}

	sortCol, err := stringToSortIndex(*sortBy)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// detect terminal width
	isTerminal := terminal.IsTerminal(int(os.Stdout.Fd()))
	if isTerminal && *width == 0 {
		w, _, err := terminal.GetSize(int(os.Stdout.Fd()))
		if err == nil {
			*width = uint(w)
		}
	}
	if *width == 0 {
		*width = 80
	}

	// read mount table
	m, warnings, err := mounts()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// print out warnings
	for _, warning := range warnings {
		fmt.Fprintln(os.Stderr, warning)
	}

	// print JSON
	if *jsonOutput {
		err := renderJSON(m)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}

		return
	}

	// print tables
	err = renderTables(m, columns, sortCol, style)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}
