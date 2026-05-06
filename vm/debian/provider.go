package debian

import "fmt"

func DefaultRegistry() string { return DefaultReg }

func BuildURL(reg, version, arch string) string {
	codename, major := resolveRef(version)
	return fmt.Sprintf(downloadPath, reg, codename, major, arch)
}

// Validate checks that version resolves to a supported Debian release and
// that arch is one of the supported canonical names.
//
// Supported releases as of May 2026:
//   - 12 / bookworm — oldstable LTS (EOL ~2028)
//   - 13 / trixie   — current stable (released Aug 2025)
//
// Debian 11 (bullseye) reached end of life in August 2025.
// Debian 14 (forky) does not yet have published cloud images.
func Validate(version, arch string) error {
	if !ValidArches[arch] {
		return fmt.Errorf("debian: unsupported arch %q — valid arches: amd64, arm64", arch)
	}
	if _, ok := releases[version]; !ok {
		return fmt.Errorf(
			"debian: unsupported version %q — valid versions: "+
				"12, bookworm, 13, trixie",
			version,
		)
	}
	return nil
}

type release struct {
	codename string
	major    string
}

// releases is the authoritative map of supported Debian cloud releases.
// Keys are both numeric majors and codenames for user convenience.
var releases = map[string]release{
	"12":       {"bookworm", "12"},
	"bookworm": {"bookworm", "12"},
	"13":       {"trixie", "13"},
	"trixie":   {"trixie", "13"},
}

// resolveRef accepts either a codename ("bookworm") or major version number
// ("12") and returns the (codename, major) pair needed to build the URL.
func resolveRef(version string) (codename, major string) {
	if r, ok := releases[version]; ok {
		return r.codename, r.major
	}
	return version, version
}