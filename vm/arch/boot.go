package arch

import (
	"fmt"

	"github.com/carbon-os/compute-image/vm/boot"
)

func BootConfig(version, arch string) (boot.Config, error) {
	if arch != "amd64" {
		return boot.Config{}, fmt.Errorf(
			"arch: unsupported arch %q — only amd64 is officially published", arch)
	}
	_ = version
	return boot.Config{
		Partition:  3,
		BootDir:    "/boot",
		KernelGlob: "vmlinuz-linux",
		InitrdGlob: "initramfs-linux.img",
	}, nil
}