package compute_image

// ContainerRef describes a container image to pull.
type ContainerRef struct {
	Image string
	Dir   string // optional; defaults to platform data dir
}

// VMRef describes a VM disk image to pull.
type VMRef struct {
	Image    string
	Registry string // optional for well-known images
	Arch     string // e.g. "amd64", "arm64", "arm", "ppc64el", "riscv64"
	Dir      string // optional; defaults to platform data dir
}

// ContainerPaths holds resolved on-disk locations for a container image.
// Base and Scratch are populated on Windows; Layers, Upper, and Work on Linux.
type ContainerPaths struct {
	Dir     string
	Base    string // Windows: read-only HCS layer directories
	Scratch string // Windows: writable scratch layer
	Layers  string // Linux: extracted layer directories (overlayfs lower)
	Upper   string // Linux: overlayfs upper (writable) dir
	Work    string // Linux: overlayfs work dir
	Cache   string
}

// VMPaths holds resolved on-disk locations for a VM image.
type VMPaths struct {
	Dir   string
	Disk  string
	Cache string
}