package ubuntu

import "fmt"

func DefaultRegistry() string { return DefaultReg }

func BuildURL(reg, version, arch string) string {
	if looksLikeVersion(version) {
		return fmt.Sprintf(releasesPath, reg, version, version, arch)
	}
	return fmt.Sprintf(dailyPath, reg, version, version, arch)
}

// Validate checks that version is a supported Ubuntu release and arch is valid.
func Validate(version, arch string) error {
	if !ValidArches[arch] {
		return fmt.Errorf("ubuntu: unsupported arch %q — valid arches: amd64, arm64", arch)
	}
	if _, ok := ValidVersions[version]; !ok {
		return fmt.Errorf(
			"ubuntu: unsupported version %q — valid versions: "+
				"22.04 (jammy), 24.04 (noble), 25.10 (questing), 26.04 (resolute)",
			version,
		)
	}
	return nil
}

// looksLikeVersion returns true for numeric versions like "24.04",
// false for codenames like "noble".
func looksLikeVersion(s string) bool {
	return len(s) > 0 && s[0] >= '0' && s[0] <= '9'
}