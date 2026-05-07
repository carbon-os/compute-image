package alpine

const (
	DefaultReg = "dl-cdn.alpinelinux.org"

	// https://{reg}/alpine/v{major.minor}/releases/cloud/nocloud_alpine-{version}-{arch}-{firmware}-tiny-r0.qcow2
	// version  = full patch string, e.g. "3.20.10"
	// arch     = x86_64 | aarch64
	// firmware = bios (x86_64 only) | uefi (aarch64 only)
	downloadPath = "https://%s/alpine/v%s/releases/cloud/nocloud_alpine-%s-%s-%s-tiny-r0.qcow2"
)

// ValidBranches is the set of active Alpine release branches (major.minor).
// Patch versions within these branches are accepted (e.g. "3.21.7").
// Users MUST supply a full patch version — bare branch refs (e.g. "3.21")
// have no corresponding file on the mirror and will 404.
//
// Source: https://alpinelinux.org/releases/ — as of May 2026.
var ValidBranches = map[string]bool{
	// "3.20" reached EOL 2026-05-01; images remain downloadable but
	// no further security patches will be issued.
	"3.20": true,
	"3.21": true, // EOL 2026-11-01
	"3.22": true, // EOL 2027-05-01
	"3.23": true, // EOL 2027-11-01
	"3.24": true, // released 2026-04-21, EOL 2028-05-01
}

// ValidArches are the canonical arch names accepted by this package.
// Alpine cloud images are built for x86_64 and aarch64 only.
var ValidArches = map[string]bool{
	"amd64": true,
	"arm64": true,
}