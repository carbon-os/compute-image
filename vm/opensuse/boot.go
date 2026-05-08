package opensuse

import (
	"fmt"

	"github.com/carbon-os/compute-image/vm/boot"
)

func BootConfig(version, arch string) (boot.Config, error) {
	if !ValidVersions[version] {
		return boot.Config{}, fmt.Errorf("opensuse: no boot config for version %q", version)
	}

	cfg := boot.Config{
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