package centos

import "fmt"

func DefaultRegistry() string { return DefaultReg }

// BuildURL constructs the CentOS Stream GenericCloud qcow2 latest-symlink URL.
// version must be a major version string: "9" or "10".
func BuildURL(reg, version, arch string) string {
	archPath := toCentOSArch(arch)
	return fmt.Sprintf(downloadPath, reg, version, archPath, version, archPath)
}

// Validate checks that version is a supported CentOS Stream major release and
// arch is one of the supported canonical names.
func Validate(version, arch string) error {
	if !ValidArches[arch] {
		return fmt.Errorf("centos: unsupported arch %q — valid arches: amd64, arm64", arch)
	}
	if !ValidMajors[version] {
		return fmt.Errorf(
			"centos: unsupported version %q — valid versions: 9, 10",
			version,
		)
	}
	return nil
}

// toCentOSArch maps canonical arch names to the path segment used in the
// CentOS Stream cloud image repo.
func toCentOSArch(arch string) string {
	switch arch {
	case "amd64":
		return "x86_64"
	case "arm64":
		return "aarch64"
	default:
		return arch
	}
}