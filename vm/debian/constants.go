package debian

const (
	DefaultReg   = "cloud.debian.org"
	downloadPath = "https://%s/images/cloud/%s/latest/debian-%s-genericcloud-%s.qcow2"
)

// ValidArches are the canonical arch names accepted by this package.
// Debian genericcloud images are published for amd64 and arm64.
// Source: https://cloud.debian.org/images/cloud/trixie/latest/ — May 2026.
var ValidArches = map[string]bool{
	"amd64": true,
	"arm64": true,
}