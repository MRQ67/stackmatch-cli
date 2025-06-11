package utils

import (
	"fmt"
	"os"
)

// ExitWithError prints a formatted error message to stderr and exits the program.
func ExitWithError(err error) {
	fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	os.Exit(1)
}
