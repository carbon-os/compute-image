package fedora

const (
	DefaultReg   = "download.fedoraproject.org"
	downloadPath = "https://%s/pub/fedora/linux/releases/%s/Cloud/%s/images/Fedora-Cloud-Base-Generic-%s.%s.qcow2"
)

// ValidMajors is the set of supported Fedora major release numbers.
// Fedora supports the current release plus the previous one (≈13 months).
//
// As of May 2026:
//   - 41  released Oct 2024, EOL Nov 2025 — images still downloadable
//   - 42  released Apr 2025 — current supported release
//
// Source: https://fedoraproject.org/wiki/Releases
var ValidMajors = map[string]bool{
	"41": true,
	"42": true,
}

// ValidArches are the canonical arch names accepted by this package.
// Fedora Cloud Generic images are built for x86_64 and aarch64.
var ValidArches = map[string]bool{
	"amd64": true,
	"arm64": true,
}