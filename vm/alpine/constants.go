package alpine

const (
	DefaultReg   = "dl-cdn.alpinelinux.org"
	downloadPath = "https://%s/alpine/v%s/releases/cloud/nocloud_alpine-%s-%s-bios-tiny-r0.qcow2"
)

// ValidBranches is the set of active Alpine release branches (major.minor).
// Patch versions within these branches are accepted (e.g. "3.23.4").
// Source: https://alpinelinux.org/releases/ — branches with cloud images as of May 2026.
var ValidBranches = map[string]bool{
	"3.20": true, // EOL 2026-05-01 — still receives security patches
	"3.21": true, // EOL 2026-11-01
	"3.22": true, // EOL 2027-05-01
	"3.23": true, // EOL 2027-11-01 — current stable
}

// ValidArches is the set of canonical arch names accepted by this package.
// Alpine cloud images are built for x86_64 and aarch64 only.
var ValidArches = map[string]bool{
	"amd64": true,
	"arm64": true,
}