package ubuntu

import (
	"fmt"

	"github.com/carbon-os/compute-image/vm"
)

// BootConfig returns the BootConfig for the given Ubuntu version and arch.
//
// Partition numbers and boot directories verified from cloud image inspection
// in May 2026 (notes.txt):
//
//	22.04 amd64 → partition 16, /boot   (root filesystem, /boot subdir)
//	22.04 arm64 → partition 1,  /boot   (root filesystem, /boot subdir)
//	24.04 *     → partition 16, /       (dedicated boot partition)
//	25.10 *     → partition 13, /       (dedicated boot partition)
//	26.04 *     → partition 13, /       (dedicated boot partition)
func BootConfig(version, arch string) (vm.BootConfig, error) {
	// Normalize codenames to numeric versions.
	switch version {
	case "jammy":
		version = "22.04"
	case "noble":
		version = "24.04"
	case "questing":
		version = "25.10"
	case "resolute":
		version = "26.04"
	}

	cfg := vm.BootConfig{
		KernelGlob: "vmlinuz-*-generic",
		InitrdGlob: "initrd.img-*-generic",
	}

	switch version {
	case "22.04":
		cfg.BootDir = "/boot"
		switch arch {
		case "amd64":
			cfg.Partition = 16
		case "arm64":
			cfg.Partition = 1
		default:
			return cfg, fmt.Errorf("ubuntu: unsupported arch %q", arch)
		}
	case "24.04":
		cfg.BootDir = "/"
		cfg.Partition = 16
	case "25.10", "26.04":
		cfg.BootDir = "/"
		cfg.Partition = 13
	default:
		return cfg, fmt.Errorf("ubuntu: no boot config for version %q", version)
	}
	return cfg, nil
}