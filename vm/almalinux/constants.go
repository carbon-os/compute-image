package almalinux

const (
	DefaultReg = "repo.almalinux.org"

	// https://{reg}/almalinux/{major}/cloud/{arch}/images/AlmaLinux-{major}-GenericCloud-latest.{arch}.qcow2
	downloadPath = "https://%s/almalinux/%s/cloud/%s/images/AlmaLinux-%s-GenericCloud-latest.%s.qcow2"
)

// ValidMajors is the set of supported AlmaLinux major release numbers.
//
// As of May 2026:
//   - 8  — EOL May 2029 (RHEL 8 lifecycle)
//   - 9  — EOL May 2032 (RHEL 9 lifecycle)
//   - 10 — current stable (released May 2025, EOL ~2035)
//
// Source: https://wiki.almalinux.org/cloud/Generic-cloud
var ValidMajors = map[string]bool{
	"8":  true,
	"9":  true,
	"10": true,
}

// ValidArches are the canonical arch names accepted by this package.
// AlmaLinux GenericCloud images are published for amd64 and arm64.
var ValidArches = map[string]bool{
	"amd64": true,
	"arm64": true,
}