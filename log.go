package compute_image

import "fmt"

// logf prints a progress line. Replace with a structured logger if needed.
func logf(format string, args ...any) {
	fmt.Printf(format+"\n", args...)
}