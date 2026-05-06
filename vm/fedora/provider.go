package fedora

import (
	"fmt"
	"strings"
)

func DefaultRegistry() string { return DefaultReg }

// BuildURL constructs the Fedora Cloud Base Generic qcow2 download URL.
// version must be the full build string, e.g. "42-1.1".
func BuildURL(reg, version, arch string) string {
	fa := toFedoraArch(arch)
	return fmt.Sprintf(downloadPath, reg, majorVersion(version), fa, version, fa)
}

// Validate checks that version's major segment is a supported Fedora release
// and that arch is valid. Version must be a full build string like "42-1.1".
func Validate(version, arch string) error {
	if !ValidArches[arch] {
		return fmt.Errorf("fedora: unsupported arch %q — valid arches: amd64, arm64", arch)
	}
	major := majorVersion(version)
	if !ValidMajors[major] {
		return fmt.Errorf(
			"fedora: unsupported version %q (major %q) — "+
				"valid majors: 41, 42  (full build string required, e.g. \"42-1.1\")",
			version, major,
		)
	}
	return nil
}

// majorVersion extracts the major release number from a build string.
// "42-1.1" → "42". Falls back to the full string if no dash is present.
func majorVersion(v string) string {
	if i := strings.IndexByte(v, '-'); i > 0 {
		return v[:i]
	}
	return v
}

func toFedoraArch(arch string) string {
	switch arch {
	case "amd64":
		return "x86_64"
	case "arm64":
		return "aarch64"
	default:
		return arch
	}
}