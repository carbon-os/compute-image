package fedora

const (
	DefaultReg  = "download.fedoraproject.org"
	ArchiveReg  = "archives.fedoraproject.org"
	livePath    = "https://%s/pub/fedora/linux/releases/%s/Cloud/%s/images/Fedora-Cloud-Base-Generic-%s.%s.qcow2"
	archivePath = "https://%s/pub/archive/fedora/linux/releases/%s/Cloud/%s/images/Fedora-Cloud-Base-Generic-%s.%s.qcow2"
)

// BuildStrings maps each supported Fedora major to its canonical release build
// string. This is what appears in the filename on the download servers.
//
// Sources verified May 2026:
//   - 41-1.4  archives.fedoraproject.org (EOL Nov 2025, moved to archive)
//   - 42-1.1  download.fedoraproject.org (current release, Apr 2025)
var BuildStrings = map[string]string{
	"41": "41-1.4",
	"42": "42-1.1",
}

// ArchivedMajors are releases that have moved off the main mirror network to
// archives.fedoraproject.org, which also uses a different URL path prefix.
var ArchivedMajors = map[string]bool{
	"41": true,
}

// ValidArches are the canonical arch names accepted by this package.
var ValidArches = map[string]bool{
	"amd64": true,
	"arm64": true,
}