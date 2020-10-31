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

const (
	localDevice   = "local"
	networkDevice = "network"
	fuseDevice    = "fuse"
	specialDevice = "special"
	loopsDevice   = "loops"
	bindsMount    = "binds"
)

var (
	Version   = ""
	CommitSHA = ""

	term  = termenv.EnvColorProfile()
	theme Theme

	allowedValues = strings.Join([]string{localDevice, networkDevice, fuseDevice, specialDevice, loopsDevice, bindsMount}, ", ")

	all         = flag.Bool("all", false, "include pseudo, duplicate, inaccessible file systems")
	hideDevices = flag.String("hide", "", "hide specific devices, separated with commas:\n"+allowedValues)
	hideFs      = flag.String("hide-fs", "", "hide specific filesystems, separated with commas")

	onlyDevices = flag.String("only", "", "show only specific devices, separated with commas:\n"+allowedValues)
	onlyFs      = flag.String("only-fs", "", "only specific filesystems, separated with commas")

	output   = flag.String("output", "", "output fields: "+strings.Join(columnIDs(), ", "))
	sortBy   = flag.String("sort", "mountpoint", "sort output by: "+strings.Join(columnIDs(), ", "))
	width    = flag.Uint("width", 0, "max output width")
	themeOpt = flag.String("theme", defaultThemeName(), "color themes: dark, light")
	styleOpt = flag.String("style", defaultStyleName(), "style: unicode, ascii")

	inodes     = flag.Bool("inodes", false, "list inode information instead of block usage")
	jsonOutput = flag.Bool("json", false, "output all devices in JSON format")
	warns      = flag.Bool("warnings", false, "output all warnings to STDERR")
	version    = flag.Bool("version", false, "display version")
)

// renderTables renders all tables.
func renderTables(m []Mount, columns []int, sortCol int, style table.Style) {
	var local, network, fuse, special []Mount
	hideDevicesMap := parseCommaSeparatedValues(*hideDevices)
	onlyDevicesMap := parseCommaSeparatedValues(*onlyDevices)
	hasOnlyDevices := len(onlyDevicesMap) != 0
	hideFsMap := parseCommaSeparatedValues(*hideFs)
	onlyFsMap := parseCommaSeparatedValues(*onlyFs)

	_, hideLocal := hideDevicesMap[localDevice]
	_, hideNetwork := hideDevicesMap[networkDevice]
	_, hideFuse := hideDevicesMap[fuseDevice]
	_, hideSpecial := hideDevicesMap[specialDevice]
	_, hideLoops := hideDevicesMap[loopsDevice]
	_, hideBinds := hideDevicesMap[bindsMount]

	_, onlyLocal := onlyDevicesMap[localDevice]
	_, onlyNetwork := onlyDevicesMap[networkDevice]
	_, onlyFuse := onlyDevicesMap[fuseDevice]
	_, onlySpecial := onlyDevicesMap[specialDevice]
	_, onlyLoops := onlyDevicesMap[loopsDevice]
	_, onlyBinds := onlyDevicesMap[bindsMount]

	// sort/filter devices
	for _, v := range m {
		if len(onlyFsMap) != 0 {
			// skip not onlyFs
			if _, ok := onlyFsMap[v.Fstype]; !ok {
				continue
			}
		} else {
			// skip hideFs
			if _, ok := hideFsMap[v.Fstype]; ok {
				continue
			}
		}
		// skip autofs
		if v.Fstype == "autofs" {
			continue
		}

		// skip bind-mounts
		if strings.Contains(v.Opts, "bind") {
			if (hasOnlyDevices && !onlyBinds) || (hideBinds && !*all) {
				continue
			}
		}

		// skip loop devices
		if strings.HasPrefix(v.Device, "/dev/loop") {
			if (hasOnlyDevices && !onlyLoops) || (hideLoops && !*all) {
				continue
			}
		}

		// skip special devices
		if v.Blocks == 0 && (!*all || !hasOnlyDevices) {
			continue
		}
		// skip zero size devices
		if v.BlockSize == 0 && (!*all || !hasOnlyDevices) {
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
	if onlyLocal {
		printTable("local", local, sortCol, columns, style)
	}
	if onlyNetwork {
		printTable("network", network, sortCol, columns, style)
	}
	if onlyFuse {
		printTable("FUSE", fuse, sortCol, columns, style)
	}
	if onlySpecial {
		printTable("special", special, sortCol, columns, style)
	}

	if !hasOnlyDevices {
		if !hideLocal || *all {
			printTable("local", local, sortCol, columns, style)
		}
		if !hideNetwork || *all {
			printTable("network", network, sortCol, columns, style)
		}
		if !hideFuse || *all {
			printTable("FUSE", fuse, sortCol, columns, style)
		}
		if !hideSpecial || *all {
			printTable("special", special, sortCol, columns, style)
		}
	}
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
		return table.Style{}, fmt.Errorf("unknown style option: %s", styleOpt)
	}
}

// parseCommaSeparatedValues parses comma separated string into a map.
func parseCommaSeparatedValues(values string) map[string]struct{} {
	items := make(map[string]struct{})
	for _, value := range strings.Split(values, ",") {
		value = strings.TrimSpace(value)
		if len(value) == 0 {
			continue
		}
		value = strings.ToLower(value)

		items[value] = struct{}{}
	}
	return items
}

func main() {
	flag.Parse()

	if *version {
		if len(CommitSHA) > 7 {
			CommitSHA = CommitSHA[:7]
		}
		if Version == "" {
			Version = "(built from source)"
		}

		fmt.Printf("duf %s", Version)
		if len(CommitSHA) > 0 {
			fmt.Printf(" (%s)", CommitSHA)
		}

		fmt.Println()
		os.Exit(0)
	}

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
	if *warns {
		for _, warning := range warnings {
			fmt.Fprintln(os.Stderr, warning)
		}
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
	renderTables(m, columns, sortCol, style)
}
