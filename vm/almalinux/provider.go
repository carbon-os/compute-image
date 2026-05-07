package almalinux

import (
	"fmt"
	"strconv"
)

func DefaultRegistry() string { return DefaultReg }

// BuildURL constructs the AlmaLinux GenericCloud qcow2 latest-symlink URL.
// version must be a major version string: "8", "9", or "10".
//
// AlmaLinux 10 switched the x86_64 path segment to x86_64_v2 (v2 microarch
// baseline); aarch64 is unchanged across all majors.
func BuildURL(reg, version, arch string) string {
	archPath := toAlmaArch(version, arch)
	return fmt.Sprintf(downloadPath, reg, version, archPath, version, archPath)
}

// Validate checks that version is a supported AlmaLinux major release and
// arch is one of the supported canonical names.
func Validate(version, arch string) error {
	if !ValidArches[arch] {
		return fmt.Errorf("almalinux: unsupported arch %q — valid arches: amd64, arm64", arch)
	}
	if !ValidMajors[version] {
		return fmt.Errorf(
			"almalinux: unsupported version %q — valid versions: 8, 9, 10",
			version,
		)
	}
	return nil
}

// toAlmaArch maps canonical arch names to the path segment used in the
// AlmaLinux repo. AlmaLinux 10+ uses x86_64_v2 for amd64; earlier majors
// use plain x86_64. aarch64 is the same across all majors.
func toAlmaArch(version, arch string) string {
	switch arch {
	case "arm64":
		return "aarch64"
	case "amd64":
		major, err := strconv.Atoi(version)
		if err == nil && major >= 10 {
			return "x86_64_v2"
		}
		return "x86_64"
	default:
		return arch
	}
}