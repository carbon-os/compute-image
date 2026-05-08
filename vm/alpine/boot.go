package alpine

import (
	"fmt"

	"github.com/carbon-os/compute-image/vm/boot"
)

func BootConfig(version, arch string) (boot.Config, error) {
	cfg := boot.Config{
		BootDir:    "/boot",
		KernelGlob: "vmlinuz-virt",
		InitrdGlob: "initramfs-virt",
	}
	switch arch {
	case "amd64":
		cfg.Partition = 1
	case "arm64":
		cfg.Partition = 2
	default:
		return cfg, fmt.Errorf("alpine: unsupported arch %q", arch)
	}
	_ = version
	return cfg, nil
}