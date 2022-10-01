package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	wildcard "github.com/IGLOU-EU/go-wildcard"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/muesli/termenv"
	"github.com/peterbourgon/ff/v3"
	"golang.org/x/term"
)

var (
	// Version contains the application version number. It's set via ldflags
	// when building.
	Version = ""

	// CommitSHA contains the SHA of the commit that this application was built
	// against. It's set via ldflags when building.
	CommitSHA = ""

	env   = termenv.EnvColorProfile()
	theme Theme

	groups        = []string{localDevice, networkDevice, fuseDevice, specialDevice, loopsDevice, bindsMount}
	allowedValues = strings.Join(groups, ", ")

	fs          = flag.NewFlagSet("duf", flag.ContinueOnError)
	all         = fs.Bool("all", false, "include pseudo, duplicate, inaccessible file systems")
	hideDevices = fs.String("hide", "", "hide specific devices, separated with commas:\n"+allowedValues)
	hideFs      = fs.String("hide-fs", "", "hide specific filesystems, separated with commas")
	hideMp      = fs.String("hide-mp", "", "hide specific mount points, separated with commas (supports wildcards)")
	onlyDevices = fs.String("only", "", "show only specific devices, separated with commas:\n"+allowedValues)
	onlyFs      = fs.String("only-fs", "", "only specific filesystems, separated with commas")
	onlyMp      = fs.String("only-mp", "", "only specific mount points, separated with commas (supports wildcards)")

	output   = fs.String("output", "", "output fields: "+strings.Join(columnIDs(), ", "))
	sortBy   = fs.String("sort", "mountpoint", "sort output by: "+strings.Join(columnIDs(), ", "))
	width    = fs.Uint("width", 0, "max output width")
	themeOpt = fs.String("theme", defaultThemeName(), "color themes: dark, light, ansi")
	styleOpt = fs.String("style", defaultStyleName(), "style: unicode, ascii")

	availThreshold = fs.String("avail-threshold", "10G,1G", "specifies the coloring threshold (yellow, red) of the avail column, must be integer with optional SI prefixes")
	usageThreshold = fs.String("usage-threshold", "0.5,0.9", "specifies the coloring threshold (yellow, red) of the usage bars as a floating point number from 0 to 1")

	inodes     = fs.Bool("inodes", false, "list inode information instead of block usage")
	jsonOutput = fs.Bool("json", false, "output all devices in JSON format")
	warns      = fs.Bool("warnings", false, "output all warnings to STDERR")
	version    = fs.Bool("version", false, "display version")
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
	m := make(map[string]struct{})
	for _, v := range strings.Split(values, ",") {
		v = strings.TrimSpace(v)
		if len(v) == 0 {
			continue
		}

		v = strings.ToLower(v)
		m[v] = struct{}{}
	}
	return m
}

// validateGroups validates the parsed group maps.
func validateGroups(m map[string]struct{}) error {
	for k := range m {
		found := false
		for _, g := range groups {
			if g == k {
				found = true
				break
			}
		}

		if !found {
			return fmt.Errorf("unknown device group: %s", k)
		}
	}

	return nil
}

// findInKey parse a slice of pattern to match the given key.
func findInKey(str string, km map[string]struct{}) bool {
	for p := range km {
		if wildcard.Match(p, str) {
			return true
		}
	}

	return false
}

func main() {
	if err := ff.Parse(fs, os.Args[1:], ff.WithEnvVarPrefix("DUF")); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

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
	if env == termenv.ANSI {
		// enforce ANSI theme for limited color support
		theme, err = loadTheme("ansi")
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
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
		HiddenMountPoints: parseCommaSeparatedValues(*hideMp),
		OnlyMountPoints:   parseCommaSeparatedValues(*onlyMp),
	}
	err = validateGroups(filters.HiddenDevices)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	err = validateGroups(filters.OnlyDevices)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
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

	// validate availability thresholds
	availbilityThresholds := strings.Split(*availThreshold, ",")
	if len(availbilityThresholds) != 2 {
		fmt.Fprintln(os.Stderr, fmt.Errorf("error parsing avail-threshold: invalid option '%s'", *availThreshold))
		os.Exit(1)
	}
	for _, thresold := range availbilityThresholds {
		_, err = stringToSize(thresold)
		if err != nil {
			fmt.Fprintln(os.Stderr, "error parsing avail-threshold:", err)
			os.Exit(1)
		}
	}

	// validate usage thresholds
	usageThresholds := strings.Split(*usageThreshold, ",")
	if len(usageThresholds) != 2 {
		fmt.Fprintln(os.Stderr, fmt.Errorf("error parsing usage-threshold: invalid option '%s'", *usageThreshold))
		os.Exit(1)
	}
	for _, thresold := range usageThresholds {
		_, err = strconv.ParseFloat(thresold, 64)
		if err != nil {
			fmt.Fprintln(os.Stderr, "error parsing usage-threshold:", err)
			os.Exit(1)
		}
	}

	// print out warnings
	if *warns {
		for _, warning := range warnings {
			fmt.Fprintln(os.Stderr, warning)
		}
	}

	// detect terminal width
	isTerminal := term.IsTerminal(int(os.Stdout.Fd()))
	if isTerminal && *width == 0 {
		w, _, err := term.GetSize(int(os.Stdout.Fd()))
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
