package debian

import (
	"fmt"

	"github.com/carbon-os/compute-image/vm"
)

// BootConfig returns the BootConfig for the given Debian version and arch.
//
// Debian genericcloud images (verified May 2026, notes.txt):
//
//	12 (bookworm) amd64 + arm64 → partition 1, /boot
//	13 (trixie)   amd64 + arm64 → partition 1, /boot
//
// Both arches share a single-partition layout where partition 1 is the root
// filesystem and /boot holds the kernel and initrd.
func BootConfig(version, arch string) (vm.BootConfig, error) {
	// Normalize codenames.
	switch version {
	case "bookworm":
		version = "12"
	case "trixie":
		version = "13"
	}

	switch version {
	case "12", "13":
		return vm.BootConfig{
			Partition:  1,
			BootDir:    "/boot",
			KernelGlob: "vmlinuz-*",
			InitrdGlob: "initrd.img-*",
		}, nil
	default:
		return vm.BootConfig{}, fmt.Errorf("debian: no boot config for version %q", version)
	}
}