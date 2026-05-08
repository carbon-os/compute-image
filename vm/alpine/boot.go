package alpine

import (
	"fmt"

	"github.com/carbon-os/compute-image/vm"
)

// BootConfig returns the BootConfig for the given Alpine version and arch.
//
// Alpine disk layout (verified May 2026, notes.txt, versions 3.20–3.23):
//
//	amd64 (bios image)  → partition 1, /boot
//	arm64 (uefi image)  → partition 2, /boot
//
// The kernel and initrd filenames are fixed across all supported branches.
// The version argument is unused because the layout is identical for 3.20–3.23.
func BootConfig(version, arch string) (vm.BootConfig, error) {
	cfg := vm.BootConfig{
		BootDir:    "/boot",
		KernelGlob: "vmlinuz-virt",
		InitrdGlob: "initramfs-virt",
	}
	switch arch {
	case "amd64":
		cfg.Partition = 1
	case "arm64":
		cfg.Partition = 2
	default:
		return cfg, fmt.Errorf("alpine: unsupported arch %q", arch)
	}
	_ = version
	return cfg, nil
}