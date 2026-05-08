package opensuse

import (
	"fmt"

	"github.com/carbon-os/compute-image/vm"
)

// BootConfig returns the BootConfig for the given openSUSE version and arch.
//
// openSUSE Leap disk layout (verified May 2026, notes.txt, 15.6 and 16.0):
//
//	amd64 → partition 3, /boot, kernel: vmlinuz-*-default, initrd: initrd-*-default
//	arm64 → partition 2, /boot, kernel: Image-*-default,   initrd: initrd-*-default
//
// arm64 uses the AArch64 EFI image name "Image-*" rather than "vmlinuz-*".
// openSUSE Leap 16.0 stores the versioned vmlinuz/Image files as symlinks;
// the extraction logic follows those symlinks automatically via vol.Stat.
// Tumbleweed is assumed to share the same layout as Leap 16.x.
func BootConfig(version, arch string) (vm.BootConfig, error) {
	if !ValidVersions[version] {
		return vm.BootConfig{}, fmt.Errorf("opensuse: no boot config for version %q", version)
	}

	cfg := vm.BootConfig{
		BootDir:    "/boot",
		InitrdGlob: "initrd-*-default",
	}
	switch arch {
	case "amd64":
		cfg.Partition = 3
		cfg.KernelGlob = "vmlinuz-*-default"
	case "arm64":
		cfg.Partition = 2
		cfg.KernelGlob = "Image-*-default"
	default:
		return cfg, fmt.Errorf("opensuse: unsupported arch %q", arch)
	}
	return cfg, nil
}