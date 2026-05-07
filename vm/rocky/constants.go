package rocky

const (
	DefaultReg = "dl.rockylinux.org"

	// Rocky 8 and 9 use the bare GenericCloud name.
	downloadPath = "https://%s/pub/rocky/%s/images/%s/Rocky-%s-GenericCloud.latest.%s.qcow2"

	// Rocky 10+ split into -Base and -LVM variants; we default to -Base.
	downloadPathV10 = "https://%s/pub/rocky/%s/images/%s/Rocky-%s-GenericCloud-Base.latest.%s.qcow2"
)

// ValidMajors is the set of supported Rocky Linux major versions.
// Rocky publishes a "latest" symlink per major, so only the major is needed.
//
// As of May 2026:
//   - 8   — active (EOL May 2029)
//   - 9   — active (EOL May 2032)
//   - 10  — active (released May 2025)
//
// Source: https://rockylinux.org/download and dl.rockylinux.org/pub/rocky/
var ValidMajors = map[string]bool{
	"8":  true,
	"9":  true,
	"10": true,
}

// ValidArches are the canonical arch names accepted by this package.
// Rocky GenericCloud images are built for x86_64 and aarch64.
var ValidArches = map[string]bool{
	"amd64": true,
	"arm64": true,
}