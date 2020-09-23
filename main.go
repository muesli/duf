package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/muesli/termenv"
)

var (
	term = termenv.ColorProfile()

	all = flag.Bool("all", false, "show all devices")

	hideLocal   = flag.Bool("hide-local", false, "hides local devices")
	hideNetwork = flag.Bool("hide-network", false, "hides network devices")
	hideFuse    = flag.Bool("hide-fuse", false, "hides fuse devices")
	hideSpecial = flag.Bool("hide-special", false, "hides special devices")
	hideBinds   = flag.Bool("hide-binds", true, "hides bind mounts")

	jsonOutput = flag.Bool("json", false, "output all devices in JSON format")
)

func renderTables(m []Mount) {
	var local, network, fuse, special []Mount
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

	if !*hideLocal || *all {
		printTable("local", local)
	}
	if !*hideNetwork || *all {
		printTable("network", network)
	}
	if !*hideFuse || *all {
		printTable("FUSE", fuse)
	}
	if !*hideSpecial || *all {
		printTable("special", special)
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

	renderTables(m)
}
