package cmd

import (
	"fmt"
	"os"
)

// fatal prints err to stderr and exits 1.
func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}

// fatalf prints a formatted message to stderr and exits 1.
func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}

// info prints an informational message to stderr.
// Informational and status messages go to stderr so they don't
// pollute stdout, which is reserved for primary machine-readable output.
func info(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
}
