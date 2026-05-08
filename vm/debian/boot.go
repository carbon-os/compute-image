package debian

import (
	"fmt"

	"github.com/carbon-os/compute-image/vm/boot"
)

func BootConfig(version, arch string) (boot.Config, error) {
	switch version {
	case "bookworm":
		version = "12"
	case "trixie":
		version = "13"
	}

	switch version {
	case "12", "13":
		return boot.Config{
			Partition:  1,
			BootDir:    "/boot",
			KernelGlob: "vmlinuz-*",
			InitrdGlob: "initrd.img-*",
		}, nil
	default:
		return boot.Config{}, fmt.Errorf("debian: no boot config for version %q", version)
	}
}