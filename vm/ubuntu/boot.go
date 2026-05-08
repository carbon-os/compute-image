package ubuntu

import (
	"fmt"

	"github.com/carbon-os/compute-image/vm/boot"
)

func BootConfig(version, arch string) (boot.Config, error) {
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

	cfg := boot.Config{
		KernelGlob: "vmlinuz-*-generic",
		InitrdGlob: "initrd.img-*-generic",
	}

	switch version {
	case "22.04":
		cfg.BootDir = "/boot"
		switch arch {
		case "amd64":
			cfg.Partition = 1  // fix: was 16, boot lives on partition 1
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