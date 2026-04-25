package compute_image

// Ref is implemented by ContainerRef and VMRef.
type Ref interface {
	imageType() string
}

// ContainerRef describes a container image to pull.
type ContainerRef struct {
	Image   string // e.g. "mcr.microsoft.com/windows/nanoserver:ltsc2022"
	BaseDir string // destination for the read-only HCS base layer
	Scratch string // destination for the writable scratch layer
	Cache   string // optional; defaults to %LOCALAPPDATA%\carbon\cache
}

// VMRef describes a VM image to pull.
type VMRef struct {
	Image    string // e.g. "ubuntu:22.04"
	Registry string // e.g. "cloud-images.ubuntu.com"
	Out      string // destination path for the prepared disk image
	Cache    string // optional; defaults to %LOCALAPPDATA%\carbon\cache
}

func (ContainerRef) imageType() string { return "container" }
func (VMRef) imageType() string        { return "vm" }

// ContainerImage is returned by Pull for a ContainerRef.
type ContainerImage struct {
	BaseLayer string // ready to pass to compute_container.ImageMount.BaseLayer
	Scratch   string // ready to pass to compute_container.ImageMount.Scratch
}

// VMImage is returned by Pull for a VMRef.
type VMImage struct {
	OutPath string // path to the prepared disk image
}