package arch

import (
	"fmt"

	"github.com/carbon-os/compute-image/vm"
)

// BootConfig returns the BootConfig for Arch Linux.
//
// Arch Linux cloud image disk layout (verified May 2026, notes.txt):
//
//	latest amd64 → partition 3, /boot
//
// Kernel and initrd use fixed filenames with no version suffix,
// consistent with Arch's rolling-release single-kernel model.
// Only amd64 is officially published by arch-boxes.
func BootConfig(version, arch string) (vm.BootConfig, error) {
	if arch != "amd64" {
		return vm.BootConfig{}, fmt.Errorf(
			"arch: unsupported arch %q — only amd64 is officially published", arch)
	}
	_ = version
	return vm.BootConfig{
		Partition:  3,
		BootDir:    "/boot",
		KernelGlob: "vmlinuz-linux",
		InitrdGlob: "initramfs-linux.img",
	}, nil
}