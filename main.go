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

	groups        = []string{localDevice, networkDevice, fuseDevice, specialDevice, loopsDevice, bindsMount}
	allowedValues = strings.Join(groups, ", ")

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

	// read mount table
	m, warnings, err := mounts()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// print JSON
	if *jsonOutput {
		err := renderJSON(m)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}

		return
	}

	// validate theme
	theme, err = loadTheme(*themeOpt)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// validate style
	style, err := parseStyle(*styleOpt)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// validate output columns
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

	// validate sort column
	sortCol, err := stringToSortIndex(*sortBy)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// validate filters
	filters := FilterOptions{
		HiddenDevices:     parseCommaSeparatedValues(*hideDevices),
		OnlyDevices:       parseCommaSeparatedValues(*onlyDevices),
		HiddenFilesystems: parseCommaSeparatedValues(*hideFs),
		OnlyFilesystems:   parseCommaSeparatedValues(*onlyFs),
	}

	// validate arguments
	if len(flag.Args()) > 0 {
		var mounts []Mount

		for _, v := range flag.Args() {
			fm, err := findMounts(m, v)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			mounts = append(mounts, fm...)
		}

		m = mounts
	}

	// print out warnings
	if *warns {
		for _, warning := range warnings {
			fmt.Fprintln(os.Stderr, warning)
		}
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

	// print tables
	renderTables(m, filters, TableOptions{
		Columns: columns,
		SortBy:  sortCol,
		Style:   style,
	})
}
