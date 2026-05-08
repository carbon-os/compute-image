package fedora

import (
	"fmt"

	"github.com/carbon-os/compute-image/vm/boot"
)

func BootConfig(version, arch string) (boot.Config, error) {
	cfg := boot.Config{
		BootDir:    "/",
		KernelGlob: "vmlinuz-*",
		InitrdGlob: "initramfs-*.img",
	}
	switch arch {
	case "amd64":
		cfg.Partition = 3
	case "arm64":
		cfg.Partition = 2
	default:
		return cfg, fmt.Errorf("fedora: unsupported arch %q", arch)
	}
	_ = version
	return cfg, nil
}