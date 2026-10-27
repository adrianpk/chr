// Command goquality enforces production signature policy over the complete source tree.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	root := flag.String("root", ".", "repository root")
	inventory := flag.Bool("inventory", false, "print declarations, imports and syntactic call sites as JSON")

	flag.Parse()

	scan, err := scanSources(*root)
	if err != nil {
		return err
	}

	if *inventory {
		return json.NewEncoder(os.Stdout).Encode(scan)
	}

	return validate(*root, scan)
}

func validate(root string, scan sourceScan) error {
	var rules policy
	if err := readJSON(filepath.Join(root, "scripts/goquality/debt.json"), &rules); err != nil {
		return fmt.Errorf("read debt policy: %w", err)
	}

	var exceptions []allowance
	if err := readJSON(filepath.Join(root, "scripts/goquality/exceptions.json"), &exceptions); err != nil {
		return fmt.Errorf("read external exceptions: %w", err)
	}

	if failures := checkPolicy(scan, rules, exceptions); len(failures) != 0 {
		return fmt.Errorf("production signature policy failed:\n%s", strings.Join(failures, "\n"))
	}

	if failures := checkLeafDependencies(scan.Imports); len(failures) != 0 {
		return fmt.Errorf("production dependency policy failed:\n%s", strings.Join(failures, "\n"))
	}

	_, err := fmt.Fprintf(os.Stdout, "production signatures: %d declarations, ceiling %d, %d exact debt entries, %d external exceptions\n", len(scan.Declarations), rules.MaxParameters, len(rules.Debt), len(exceptions))

	return err
}
