package centos

const (
	DefaultReg = "cloud.centos.org"

	// https://{reg}/centos/{major}-stream/{arch}/images/CentOS-Stream-GenericCloud-{major}-latest.{arch}.qcow2
	downloadPath = "https://%s/centos/%s-stream/%s/images/CentOS-Stream-GenericCloud-%s-latest.%s.qcow2"
)

// ValidMajors is the set of supported CentOS Stream major release numbers.
//
// As of May 2026:
//   - 9  — EOL May 2027 (tracks RHEL 9)
//   - 10 — current stable (tracks RHEL 10, released ~2024)
//
// CentOS Stream 8 reached end of life in December 2024.
// Source: https://cloud.centos.org and https://www.centos.org/download/
var ValidMajors = map[string]bool{
	"9":  true,
	"10": true,
}

// ValidArches are the canonical arch names accepted by this package.
// CentOS Stream GenericCloud images are published for amd64 and arm64.
var ValidArches = map[string]bool{
	"amd64": true,
	"arm64": true,
}