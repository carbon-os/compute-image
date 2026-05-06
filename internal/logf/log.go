package logf

import "fmt"

// Logf prints a progress line to stdout. Replace with a structured logger if needed.
func Logf(format string, args ...any) {
	fmt.Printf(format+"\n", args...)
}