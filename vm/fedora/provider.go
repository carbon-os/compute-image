package fedora

import (
	"fmt"
	"strings"
)

func DefaultRegistry() string { return DefaultReg }

// BuildURL constructs the Fedora Cloud Base Generic qcow2 download URL.
// version may be a bare major ("42") or a full build string ("42-1.1") —
// both are resolved to the canonical build string via BuildStrings.
// Archived releases ignore reg and always use ArchiveReg with the archive path.
func BuildURL(reg, version, arch string) string {
	fa := toFedoraArch(arch)
	major := majorVersion(version)
	build := resolveBuild(version)
	if ArchivedMajors[major] {
		return fmt.Sprintf(archivePath, ArchiveReg, major, fa, build, fa)
	}
	return fmt.Sprintf(livePath, reg, major, fa, build, fa)
}

// Validate checks that version resolves to a known Fedora major and that arch
// is supported. Bare majors ("42") and full build strings ("42-1.1") are both
// accepted.
func Validate(version, arch string) error {
	if !ValidArches[arch] {
		return fmt.Errorf("fedora: unsupported arch %q — valid arches: amd64, arm64", arch)
	}
	major := majorVersion(version)
	if _, ok := BuildStrings[major]; !ok {
		return fmt.Errorf(
			"fedora: unsupported version %q (major %q) — valid majors: 41, 42",
			version, major,
		)
	}
	return nil
}

// resolveBuild returns the full build string for version.
// "42" → "42-1.1",  "42-1.1" → "42-1.1" (pass-through if already qualified).
func resolveBuild(v string) string {
	major := majorVersion(v)
	if build, ok := BuildStrings[major]; ok {
		return build
	}
	return v
}

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