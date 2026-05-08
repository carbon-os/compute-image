package fedora

import (
	"fmt"

	"github.com/carbon-os/compute-image/vm"
)

// BootConfig returns the BootConfig for the given Fedora version and arch.
//
// Fedora Cloud Base Generic disk layout (verified May 2026, notes.txt, F41–F42):
//
//	amd64 (x86_64)  → partition 3, /
//	arm64 (aarch64) → partition 2, /
//
// The boot partition root "/" directly holds the kernel and initrd alongside
// grub2/, efi/, and loader/ directories. Rescue kernel entries whose names
// contain "rescue" are excluded automatically by the extraction logic.
// The version argument is unused because the layout is identical across F41–F42.
func BootConfig(version, arch string) (vm.BootConfig, error) {
	cfg := vm.BootConfig{
		BootDir:    "/",
		KernelGlob: "vmlinuz-*",
		InitrdGlob: "initramfs-*.img",
	}
	switch arch {
	case "amd64":
		cfg.Partition = 3
	case "arm64":
		cfg.Partition = 2
	default:
		return cfg, fmt.Errorf("fedora: unsupported arch %q", arch)
	}
	_ = version
	return cfg, nil
}