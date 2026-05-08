package rocky

import (
	"fmt"

	"github.com/carbon-os/compute-image/vm"
)

// BootConfig returns the BootConfig for the given Rocky Linux version and arch.
//
// Rocky Linux disk layout (verified May 2026, notes.txt):
//
//	8  amd64 + arm64 → partition 2, /
//	9  amd64         → partition 3, /
//	9  arm64         → partition 2, /
//	10 amd64         → partition 3, /
//	10 arm64         → partition 2, /
//
// The boot partition root "/" directly holds the kernel and initrd alongside
// grub2/, efi/, and loader/ directories. Rescue entries whose names contain
// "rescue" are excluded automatically by the extraction logic.
func BootConfig(version, arch string) (vm.BootConfig, error) {
	if arch != "amd64" && arch != "arm64" {
		return vm.BootConfig{}, fmt.Errorf("rocky: unsupported arch %q", arch)
	}

	cfg := vm.BootConfig{
		BootDir:    "/",
		KernelGlob: "vmlinuz-*",
		InitrdGlob: "initramfs-*.img",
	}

	major := majorVersion(version)
	switch major {
	case "8":
		// Both arches share partition 2 on Rocky 8.
		cfg.Partition = 2
	case "9", "10":
		if arch == "amd64" {
			cfg.Partition = 3
		} else {
			cfg.Partition = 2
		}
	default:
		return cfg, fmt.Errorf("rocky: no boot config for version %q", version)
	}
	return cfg, nil
}