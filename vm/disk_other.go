//go:build !windows

package vm

import "os"

const diskFile = "disk.img"

// finalizeDisk promotes the raw image to its final resting place as a plain .img.
func finalizeDisk(rawPath, diskPath string) error {
	return os.Rename(rawPath, diskPath)
}