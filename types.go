package compute_image

// ContainerRef identifies a container image and optional storage root.
type ContainerRef struct {
	Image string
	Dir   string
}

// VMRef identifies a VM disk image along with its registry, target architecture,
// and pull-time options.
type VMRef struct {
	Image    string
	Registry string
	Arch     string
	Dir      string
	// ExtractKernel, when true, extracts "vmlinuz" and "initrd" files into the
	// same directory as the disk image after a successful pull.
	ExtractKernel bool
}

// ContainerPaths holds the resolved filesystem paths for a pulled container image.
// Fields not applicable to the current platform are left empty.
type ContainerPaths struct {
	Dir     string // image directory
	Layers  string // Linux overlayfs lower layers (read-only)
	Upper   string // Linux overlayfs upper layer (writable)
	Work    string // Linux overlayfs work directory
	Base    string // Windows base VHD
	Scratch string // Windows scratch VHD
	Cache   string // shared download cache
}

// VMPaths holds the resolved filesystem paths for a pulled VM disk image.
type VMPaths struct {
	Dir   string // image directory
	Disk  string // disk.img on Linux/macOS, disk.vhd on Windows
	Cache string // shared download cache
}