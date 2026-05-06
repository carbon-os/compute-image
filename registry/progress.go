package registry

import (
	"fmt"
	"strings"
	"time"
	"io"
)

// ProgressReader wraps an io.Reader and prints download progress to stdout.
type ProgressReader struct {
	R        io.Reader
	Total    int64
	current  int64
	lastTick time.Time
}

func (pr *ProgressReader) Read(p []byte) (int, error) {
	n, err := pr.R.Read(p)
	pr.current += int64(n)
	if time.Since(pr.lastTick) > 100*time.Millisecond || err == io.EOF {
		pr.lastTick = time.Now()
		pr.printBar()
	}
	return n, err
}

func (pr *ProgressReader) printBar() {
	const width = 40
	pct := float64(pr.current) / float64(pr.Total) * 100
	if pr.Total <= 0 {
		pct = 0
	}
	filled := int(pct / 100 * width)
	if filled > width {
		filled = width
	}
	bar := strings.Repeat("=", filled) + strings.Repeat(" ", width-filled)
	fmt.Printf("\r    [%s] %.1f%% (%d MB)", bar, pct, pr.current/1024/1024)
}

// HumanBytes formats a byte count as a human-readable string.
func HumanBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}