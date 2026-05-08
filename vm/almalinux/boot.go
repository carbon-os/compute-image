package almalinux

import (
	"fmt"

	"github.com/carbon-os/compute-image/vm"
)

// BootConfig returns the BootConfig for the given AlmaLinux version and arch.
//
// AlmaLinux disk layout (verified May 2026, notes.txt, versions 8–10):
//
//	amd64 → partition 3, /
//	arm64 → partition 2, /
//
// The partition assignment is consistent across all three supported majors.
// The boot partition root "/" directly holds the kernel and initrd.
// Rescue entries whose names contain "rescue" are excluded automatically.
func BootConfig(version, arch string) (vm.BootConfig, error) {
	if !ValidMajors[version] {
		return vm.BootConfig{}, fmt.Errorf("almalinux: no boot config for version %q", version)
	}
	if !ValidArches[arch] {
		return vm.BootConfig{}, fmt.Errorf("almalinux: unsupported arch %q", arch)
	}

	cfg := vm.BootConfig{
		BootDir:    "/",
		KernelGlob: "vmlinuz-*",
		InitrdGlob: "initramfs-*.img",
	}
	if arch == "amd64" {
		cfg.Partition = 3
	} else {
		cfg.Partition = 2
	}
	return cfg, nil
}