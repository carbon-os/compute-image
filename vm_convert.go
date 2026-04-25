package compute_image

import (
	"fmt"
	"io"
	"os"
	"time"

	qcow2reader "github.com/lima-vm/go-qcow2reader"
)

// convertQcow2ToRaw converts a qcow2 image to a flat raw disk image.
func convertQcow2ToRaw(srcPath, dstPath string) error {
	if !isQcow2(srcPath) {
		return fmt.Errorf("%s is not a valid QCOW2 image", srcPath)
	}

	f, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer f.Close()

	img, err := qcow2reader.Open(f)
	if err != nil {
		return fmt.Errorf("parse qcow2: %w", err)
	}

	size := img.Size()
	logf("    Virtual size: %s", humanBytes(size))

	out, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer func() {
		out.Close()
		if err != nil {
			os.Remove(dstPath)
		}
	}()

	src := io.NewSectionReader(img, 0, size)
	return copyWithProgress(out, src, size)
}

func isQcow2(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()
	magic := make([]byte, 4)
	if _, err := f.Read(magic); err != nil {
		return false
	}
	return magic[0] == 'Q' && magic[1] == 'F' && magic[2] == 'I' && magic[3] == 0xfb
}

func copyWithProgress(dst io.Writer, src io.Reader, total int64) error {
	const chunkSize = 4 * 1024 * 1024
	buf := make([]byte, chunkSize)
	var written int64
	start := time.Now()

	for {
		n, err := src.Read(buf)
		if n > 0 {
			nw, werr := dst.Write(buf[:n])
			written += int64(nw)
			if werr != nil {
				return werr
			}
			elapsed := time.Since(start).Seconds()
			speed := float64(written) / elapsed / 1024 / 1024
			fmt.Printf("\r    %.1f%%  %s / %s  (%.1f MB/s)   ",
				float64(written)/float64(total)*100,
				humanBytes(written), humanBytes(total), speed)
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func humanBytes(b int64) string {
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