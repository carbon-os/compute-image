package rocky

import (
	"fmt"

	"github.com/carbon-os/compute-image/vm/boot"
)

func BootConfig(version, arch string) (boot.Config, error) {
	if arch != "amd64" && arch != "arm64" {
		return boot.Config{}, fmt.Errorf("rocky: unsupported arch %q", arch)
	}

	cfg := boot.Config{
		BootDir:    "/",
		KernelGlob: "vmlinuz-*",
		InitrdGlob: "initramfs-*.img",
	}

	major := majorVersion(version)
	switch major {
	case "8":
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