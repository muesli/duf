package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/muesli/termenv"
	"golang.org/x/crypto/ssh/terminal"
)

var (
	term = termenv.ColorProfile()

	all         = flag.Bool("all", false, "include pseudo, duplicate, inaccessible file systems")
	hideLocal   = flag.Bool("hide-local", false, "hide local devices")
	hideNetwork = flag.Bool("hide-network", false, "hide network devices")
	hideFuse    = flag.Bool("hide-fuse", false, "hide fuse devices")
	hideSpecial = flag.Bool("hide-special", false, "hide special devices")
	hideLoops   = flag.Bool("hide-loops", true, "hide loop devices")
	hideBinds   = flag.Bool("hide-binds", true, "hide bind mounts")

	output = flag.String("output", "", "output fields: "+strings.Join(columnIDs(), ", "))
	sortBy = flag.String("sort", "mountpoint", "sort output by: "+strings.Join(columnIDs(), ", "))
	width  = flag.Uint("width", 0, "max output width")

	inodes     = flag.Bool("inodes", false, "list inode information instead of block usage")
	jsonOutput = flag.Bool("json", false, "output all devices in JSON format")
)

func renderTables(m []Mount, sortCol int) error {
	var local, network, fuse, special []Mount

	// sort/filter devices
	for _, v := range m {
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

	columns, err := parseColumns(*output)
	if err != nil {
		return err
	}
	if len(columns) == 0 {
		// no columns supplied, use defaults
		if *inodes {
			columns = []int{1, 6, 7, 8, 9, 10, 11}
		} else {
			columns = []int{1, 2, 3, 4, 5, 10, 11}
		}
	}

	// print tables
	if !*hideLocal || *all {
		printTable("local", local, sortCol, columns)
	}
	if !*hideNetwork || *all {
		printTable("network", network, sortCol, columns)
	}
	if !*hideFuse || *all {
		printTable("FUSE", fuse, sortCol, columns)
	}
	if !*hideSpecial || *all {
		printTable("special", special, sortCol, columns)
	}
	return nil
}

func renderJSON(m []Mount) error {
	output, err := json.MarshalIndent(m, "", " ")
	if err != nil {
		return fmt.Errorf("error formatting the json output: %s", err)
	}

	fmt.Println(string(output))
	return nil
}

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

func main() {
	flag.Parse()

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

	err = renderTables(m, sortCol)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}
