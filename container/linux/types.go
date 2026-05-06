package linux

// Ref describes a Linux container image to pull.
type Ref struct {
	Image string // e.g. "ubuntu:24.04"
	Dir   string // optional; root dir for all image data — defaults to $HOME/.local/share/carbon
}

// Image is returned by Pull.
type Image struct {
	Image  string
	Paths  Paths
	RootFS string // path to the top extracted layer; use as lower dir for overlayfs
}

// Paths holds all resolved on-disk locations for a Linux container image.
type Paths struct {
	Dir    string // e.g. /home/user/.local/share/carbon/index.docker.io/library/ubuntu/24.04
	Layers string // numbered extracted layer subdirectories (lower dirs for overlayfs)
	Upper  string // overlayfs upper (writable) dir — created but not mounted here
	Work   string // overlayfs work dir — created but not mounted here
	Cache  string // raw downloaded layer tarballs
}