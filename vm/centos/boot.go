package centos

import (
	"fmt"

	"github.com/carbon-os/compute-image/vm"
)

// BootConfig returns the BootConfig for the given CentOS Stream version and arch.
//
// CentOS Stream disk layout (verified May 2026, notes.txt):
//
//	9  amd64 → partition 1, /boot   (root filesystem, /boot subdir)
//	9  arm64 → partition 2, /boot
//	10 amd64 → partition 1, /boot   (inferred from Stream 9 amd64 pattern)
//	10 arm64 → partition 2, /boot
func BootConfig(version, arch string) (vm.BootConfig, error) {
	if !ValidMajors[version] {
		return vm.BootConfig{}, fmt.Errorf("centos: no boot config for version %q", version)
	}
	if !ValidArches[arch] {
		return vm.BootConfig{}, fmt.Errorf("centos: unsupported arch %q", arch)
	}

	cfg := vm.BootConfig{
		BootDir:    "/boot",
		KernelGlob: "vmlinuz-*",
		InitrdGlob: "initramfs-*.img",
	}
	if arch == "amd64" {
		cfg.Partition = 1
	} else {
		cfg.Partition = 2
	}
	return cfg, nil
}