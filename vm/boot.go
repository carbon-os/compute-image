package vm

// BootConfig describes where to find the kernel and initrd within a disk image.
// Every field is hardcoded per distro/version/arch in the per-distro boot.go files;
// no runtime probing is performed.
type BootConfig struct {
	// Partition is the 1-based partition number to mount.
	Partition int
	// BootDir is the path within that partition that contains the boot files.
	// "/" for a dedicated boot partition; "/boot" when the root partition is mounted.
	BootDir string
	// KernelGlob is a path.Match pattern for the kernel filename.
	// e.g. "vmlinuz-*-generic", "vmlinuz-virt", "Image-*-default"
	KernelGlob string
	// InitrdGlob is a path.Match pattern for the initrd filename.
	// e.g. "initrd.img-*-generic", "initramfs-virt", "initrd-*-default"
	InitrdGlob string
}