package compute_image

// Ref is implemented by ContainerRef and VMRef.
type Ref interface {
	imageType() string
}

// ContainerRef describes a container image to pull.
type ContainerRef struct {
	Image string // e.g. "mcr.microsoft.com/windows/nanoserver:ltsc2022"
	Dir   string // optional; root dir for all image data — defaults to %LOCALAPPDATA%\carbon
}

// VMRef describes a VM image to pull.
type VMRef struct {
	Image    string // e.g. "ubuntu:22.04"
	Registry string // e.g. "cloud-images.ubuntu.com"
	Dir      string // optional; root dir for all image data — defaults to %LOCALAPPDATA%\carbon
}

func (ContainerRef) imageType() string { return "container" }
func (VMRef) imageType() string        { return "vm" }

// ContainerPaths holds all resolved on-disk locations for a container image.
type ContainerPaths struct {
	Dir     string // e.g. C:\carbon\mcr.microsoft.com\windows\nanoserver\ltsc2022
	Base    string // read-only HCS base layer
	Scratch string // writable scratch layer
	Cache   string // raw downloaded layer tarballs
}

// VMPaths holds all resolved on-disk locations for a VM image.
type VMPaths struct {
	Dir   string // e.g. C:\carbon\cloud-images.ubuntu.com\ubuntu\22.04
	Disk  string // prepared disk image
	Cache string // raw downloaded qcow2
}

// ContainerImage is returned by Pull for a ContainerRef.
type ContainerImage struct {
	Image     string
	Paths     ContainerPaths
	BaseLayer string // = Paths.Base; ready to pass to compute_container.ImageMount
	Scratch   string // = Paths.Scratch; ready to pass to compute_container.ImageMount
}

// VMImage is returned by Pull for a VMRef.
type VMImage struct {
	Image   string
	Paths   VMPaths
	OutPath string // = Paths.Disk; path to the prepared disk image
}