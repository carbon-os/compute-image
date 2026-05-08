package centos

import (
	"fmt"

	"github.com/carbon-os/compute-image/vm/boot"
)

func BootConfig(version, arch string) (boot.Config, error) {
	if !ValidMajors[version] {
		return boot.Config{}, fmt.Errorf("centos: no boot config for version %q", version)
	}
	if !ValidArches[arch] {
		return boot.Config{}, fmt.Errorf("centos: unsupported arch %q", arch)
	}

	cfg := boot.Config{
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