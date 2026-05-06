//go:build windows

package vm

const diskFile = "disk.vhd"

// finalizeDisk converts the intermediate raw image to a fixed VHD footer format.
func finalizeDisk(rawPath, diskPath string) error {
	return convertRawToVHD(rawPath, diskPath)
}