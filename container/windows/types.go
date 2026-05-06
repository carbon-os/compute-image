//go:build windows

package windows

// Ref describes a Windows container image to pull.
type Ref struct {
	Image string // e.g. "mcr.microsoft.com/windows/nanoserver:ltsc2022"
	Dir   string // optional; root dir for all image data — defaults to %LOCALAPPDATA%\carbon
}

// Image is returned by Pull.
type Image struct {
	Image     string
	Paths     Paths
	BaseLayer string // topmost read-only HCS layer; pass to compute_container.ImageMount
	Scratch   string // writable scratch layer; pass to compute_container.ImageMount
}

// Paths holds all resolved on-disk locations for a Windows container image.
type Paths struct {
	Dir     string // e.g. C:\carbon\mcr.microsoft.com\windows\nanoserver\ltsc2022
	Base    string // root dir for numbered read-only HCS layer subdirectories
	Scratch string // writable scratch layer
	Cache   string // raw downloaded layer tarballs
}