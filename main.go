package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/muesli/termenv"
	"golang.org/x/crypto/ssh/terminal"
)

var (
	term = termenv.ColorProfile()

	all = flag.Bool("all", false, "show all devices")

	hideLocal    = flag.Bool("hide-local", false, "hide local devices")
	hideNetwork  = flag.Bool("hide-network", false, "hide network devices")
	hideFuse     = flag.Bool("hide-fuse", false, "hide fuse devices")
	hideSpecial  = flag.Bool("hide-special", false, "hide special devices")
	hideLoopback = flag.Bool("hide-loopback", true, "hide loopback devices")
	hideBinds    = flag.Bool("hide-binds", true, "hide bind mounts")

	sortBy = flag.String("sort", "mountpoint", "sort output by key (mountpoint, size, used, avail, usage, type, filesystem)")
	width  = flag.Uint("width", 0, "max output width")

	jsonOutput = flag.Bool("json", false, "output all devices in JSON format")
)

func renderTables(m []Mount, sortCol int) {
	var local, network, fuse, special []Mount

	// sort/filter devices
	for _, v := range m {
		if isNetworkFs(v.Stat) {
			network = append(network, v)
			continue
		}
		if isFuseFs(v.Stat) {
			fuse = append(fuse, v)
			continue
		}
		if isSpecialFs(v.Stat) {
			special = append(special, v)
			continue
		}

		local = append(local, v)
	}

	// print tables
	if !*hideLocal || *all {
		printTable("local", local, sortCol)
	}
	if !*hideNetwork || *all {
		printTable("network", network, sortCol)
	}
	if !*hideFuse || *all {
		printTable("FUSE", fuse, sortCol)
	}
	if !*hideSpecial || *all {
		printTable("special", special, sortCol)
	}
}

func renderJSON(m []Mount) error {
	output, err := json.MarshalIndent(m, "", " ")
	if err != nil {
		return fmt.Errorf("error formatting the json output: %s", err)
	}

	fmt.Println(string(output))
	return nil
}

func main() {
	flag.Parse()

	sortCol, err := stringToColumn(*sortBy)
	if err != nil {
		fmt.Fprintln(os.Stderr, "unknown sort key, valid values: mountpoint, size, used, avail, usage, type, filesystem")
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

	renderTables(m, sortCol)
}
