package arch

const (
	DefaultReg = "geo.mirror.pkgbuild.com"

	// Arch is a rolling release — there is no version segment in the path.
	// Only x86_64 is officially produced by the arch-boxes project.
	downloadPath = "https://%s/images/latest/Arch-Linux-%s-cloudimg.qcow2"
)

// ValidVersions is the set of accepted version strings for Arch Linux.
// Arch is a rolling release; the only meaningful version is "latest".
var ValidVersions = map[string]bool{
	"latest": true,
}

// ValidArches are the canonical arch names accepted by this package.
// Only amd64 (x86_64) is officially published by arch-boxes.
// Source: https://geo.mirror.pkgbuild.com/images/latest/
var ValidArches = map[string]bool{
	"amd64": true,
}