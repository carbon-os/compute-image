package opensuse

const (
	DefaultReg = "download.opensuse.org"

	// Leap 15.x: openSUSE-Leap-15.6-Minimal-VM.x86_64-Cloud.qcow2
	leap15Path = "https://%s/distribution/leap/%s/appliances/openSUSE-Leap-%s-Minimal-VM.%s-Cloud.qcow2"

	// Leap 16.x: dropped the "openSUSE-" prefix in the filename.
	leap16Path = "https://%s/distribution/leap/%s/appliances/Leap-%s-Minimal-VM.%s-Cloud.qcow2"

	// Tumbleweed: rolling, no version segment in the path.
	tumbleweedPath = "https://%s/tumbleweed/appliances/openSUSE-Tumbleweed-Minimal-VM.%s-Cloud.qcow2"
)

// ValidVersions is the set of supported openSUSE versions.
//
// As of May 2026:
//   - "tumbleweed"  — rolling release, always current
//   - "15.6"        — Leap 15.x current stable (EOL Dec 2025; still downloadable)
//   - "16.0"        — Leap 16.x current stable (released Apr 2025)
//
// Leap 15.5 and earlier are end-of-life and no longer receive updates.
// Source: https://get.opensuse.org/leap/ and download.opensuse.org
var ValidVersions = map[string]bool{
	"tumbleweed": true,
	"15.6":       true,
	"16.0":       true,
}

// ValidArches are the canonical arch names accepted by this package.
// openSUSE Leap and Tumbleweed Minimal-VM images are built for x86_64 and aarch64.
var ValidArches = map[string]bool{
	"amd64": true,
	"arm64": true,
}