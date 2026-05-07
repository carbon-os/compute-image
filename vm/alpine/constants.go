package alpine

const (
	DefaultReg   = "dl-cdn.alpinelinux.org"
	downloadPath = "https://%s/alpine/v%s/releases/cloud/nocloud_alpine-%s-%s-%s-tiny-r0.qcow2"
)

// ValidBranches is the set of active Alpine release branches (major.minor).
// Source: https://alpinelinux.org/releases/ — as of May 2026.
var ValidBranches = map[string]bool{
	"3.20": true, // EOL 2026-05-01
	"3.21": true, // EOL 2026-11-01
	"3.22": true, // EOL 2027-05-01
	"3.23": true, // EOL 2027-11-01 — current stable
}

// LatestPatch maps each supported branch to its current patch release.
// Update these when new patch releases are published.
// Source: https://alpinelinux.org/releases/
var LatestPatch = map[string]string{
	"3.20": "3.20.10",
	"3.21": "3.21.7",
	"3.22": "3.22.4",
	"3.23": "3.23.4",
}

var ValidArches = map[string]bool{
	"amd64": true,
	"arm64": true,
}