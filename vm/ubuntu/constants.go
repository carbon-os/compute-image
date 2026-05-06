package ubuntu

const (
	DefaultReg   = "cloud-images.ubuntu.com"
	releasesPath = "https://%s/releases/%s/release/ubuntu-%s-server-cloudimg-%s.img"
	dailyPath    = "https://%s/%s/current/%s-server-cloudimg-%s.img"
)

// ValidVersions maps every accepted version string (numeric or codename) to
// its canonical codename. Both forms are accepted by Validate.
//
// Active releases as of May 2026 — source: https://cloud-images.ubuntu.com/
//   - 22.04 / jammy     — LTS, standard support until 2027
//   - 24.04 / noble     — LTS, current (released Apr 2024)
//   - 25.10 / questing  — interim (releasing Oct 2025)
//   - 26.04 / resolute  — next LTS (releasing Apr 2026)
//
// 20.04 (focal) and 25.04 (plucky) have reached end of standard support.
var ValidVersions = map[string]string{
	"22.04":    "jammy",
	"jammy":    "jammy",
	"24.04":    "noble",
	"noble":    "noble",
	"25.10":    "questing",
	"questing": "questing",
	"26.04":    "resolute",
	"resolute": "resolute",
}

// ValidArches are the canonical arch names accepted by this package.
// Ubuntu cloud images are published for amd64 and arm64.
var ValidArches = map[string]bool{
	"amd64": true,
	"arm64": true,
}