# image-cli

Command-line tool for pulling and managing images for Carbon Compute. Handles
both container images and Linux VM disk images (qcow2 → raw → VHD) across
Windows and Linux.

## Usage

```
image-cli pull container <image>        [--dir <path>]
image-cli pull vm        <image>        --arch <arch> [--registry <host>] [--dir <path>]
image-cli info container <image>        [--dir <path>]
image-cli info vm        <image>        --arch <arch> [--registry <host>] [--dir <path>]
image-cli ls                            [--dir <path>]
image-cli rm             <image>        [--dir <path>]
image-cli rm-all                        [--dir <path>]
```

## Examples

```sh
# Pull a container image
image-cli pull container mcr.microsoft.com/windows/nanoserver:ltsc2022
image-cli pull container ubuntu:24.04

# Pull to a specific directory
image-cli pull container ubuntu:24.04 --dir /tmp/images

# Pull a VM image by version number
image-cli pull vm ubuntu:22.04 --arch amd64
image-cli pull vm ubuntu:22.04 --arch arm64

# Pull a VM image by codename
image-cli pull vm ubuntu:noble --arch amd64
image-cli pull vm debian:bookworm --arch amd64
image-cli pull vm debian:bookworm --arch arm64

# Debian also accepts major version numbers
image-cli pull vm debian:12 --arch amd64

# Alpine, Fedora, openSUSE, Rocky, and Arch all have default registries too
image-cli pull vm alpine:3.23 --arch amd64
image-cli pull vm fedora:42-1.1 --arch amd64
image-cli pull vm opensuse:16.0 --arch amd64
image-cli pull vm rocky:9 --arch amd64
image-cli pull vm arch:latest --arch amd64

# Pull from a custom registry (overrides the default for any distro)
image-cli pull vm ubuntu:22.04 --arch amd64 --registry my.registry.example.com

# Inspect paths and status without pulling anything
image-cli info container ubuntu:24.04
image-cli info vm ubuntu:22.04 --arch amd64
image-cli info vm debian:bookworm --arch arm64

# List all cached layers and images
image-cli ls
image-cli ls --dir /tmp/images

# Remove a container image
image-cli rm ubuntu:24.04

# Remove all image data under the default carbon directory
image-cli rm-all
image-cli rm-all --dir /tmp/images
```

## pull container

Downloads and imports a container image. All paths are resolved automatically
from `--dir` and the image reference.

```
image-cli pull container <image> [--dir <path>]
```

On Windows, images are imported as HCS layers and must be run as Administrator.
On Linux, layers are extracted for use with overlayfs.

Output varies by platform:

```
# Windows
[+] Container image ready.
    Dir:     C:\Users\you\AppData\Local\carbon\mcr.microsoft.com\windows\nanoserver\ltsc2022
    Base:    C:\Users\you\AppData\Local\carbon\mcr.microsoft.com\windows\nanoserver\ltsc2022\base
    Scratch: C:\Users\you\AppData\Local\carbon\mcr.microsoft.com\windows\nanoserver\ltsc2022\scratch
    Cache:   C:\Users\you\AppData\Local\carbon\cache

# Linux
[+] Container image ready.
    Dir:    /home/you/.local/share/carbon/index.docker.io/library/ubuntu/24.04
    Layers: /home/you/.local/share/carbon/index.docker.io/library/ubuntu/24.04/layers
    Upper:  /home/you/.local/share/carbon/index.docker.io/library/ubuntu/24.04/upper
    Work:   /home/you/.local/share/carbon/index.docker.io/library/ubuntu/24.04/work
    Cache:  /home/you/.local/share/carbon/cache
```

## pull vm

Downloads and prepares a Linux VM disk image. The source qcow2 is converted
to a fixed VHD; the intermediate raw file is removed automatically.

```
image-cli pull vm <image> --arch <arch> [--registry <host>] [--dir <path>]
```

`--arch` is required. `--registry` is optional for all well-known images — see
the registry table below.

The arch is part of the on-disk path, so multiple architectures of the same
image can coexist without conflict.

```
[+] VM image ready.
    Dir:   /home/you/.local/share/carbon/cloud-images.ubuntu.com/ubuntu/22.04/amd64
    Disk:  /home/you/.local/share/carbon/cloud-images.ubuntu.com/ubuntu/22.04/amd64/disk.vhd
    Cache: /home/you/.local/share/carbon/cache
```

Supported architectures: `amd64`, `arm64`

Note: not every distro publishes images for every architecture. Arch Linux only
publishes `amd64`. See the well-known registries table below for per-distro
arch availability.

## info

Shows all resolved paths and current status without downloading or importing
anything. Useful for checking what is on disk or getting paths to pass to
other tools.

```
image-cli info container <image> [--dir <path>]
image-cli info vm        <image> --arch <arch> [--registry <host>] [--dir <path>]
```

Status values:

| Status | Meaning |
|---|---|
| `ready` | Image is fully prepared and ready to use |
| `cached` | Layers are downloaded but not yet imported or converted |
| `not pulled` | Nothing on disk yet |

## rm

Removes the on-disk image data for a container image. Matching layer tarballs
in the cache are also removed. VM cache files are left intact — use `rm-all`
to wipe everything.

```
image-cli rm <image> [--dir <path>]
```

On Windows this also destroys the HCS layer locks before deletion.

## rm-all

Removes all image data under the carbon root directory. On Windows, HCS layer
locks are released and restrictive file attributes are stripped before deletion
to avoid permission errors.

```
image-cli rm-all [--dir <path>]
```

Defaults to `%LOCALAPPDATA%\carbon` on Windows, `~/.local/share/carbon` on Linux.

## Cache

Downloaded layers are cached by digest and reused across pulls, so re-running
`pull` for an already-downloaded image only re-runs the import or conversion
step. VM images are cached per arch — `ubuntu:22.04 amd64` and
`ubuntu:22.04 arm64` are stored as separate cache entries. The default cache
location is inside the carbon root directory. Override with `--dir`.

## Well-known VM registries

`--registry` can be omitted for all of the following images:

| Image | Default registry | Supported arches | Accepted versions |
|---|---|---|---|
| `alpine` | `dl-cdn.alpinelinux.org` | `amd64`, `arm64` | `3.20`, `3.21`, `3.22`, `3.23` (patch versions accepted, e.g. `3.23.4`) |
| `debian` | `cloud.debian.org` | `amd64`, `arm64` | `12`, `bookworm`, `13`, `trixie` |
| `ubuntu` | `cloud-images.ubuntu.com` | `amd64`, `arm64` | see table below |
| `fedora` | `download.fedoraproject.org` | `amd64`, `arm64` | full build string required, e.g. `42-1.1` (majors: `41`, `42`) |
| `opensuse` | `download.opensuse.org` | `amd64`, `arm64` | `tumbleweed`, `15.6`, `16.0` |
| `rocky` | `dl.rockylinux.org` | `amd64`, `arm64` | `8`, `9`, `10` (patch versions accepted, e.g. `9.7`) |
| `arch` | `geo.mirror.pkgbuild.com` | `amd64` | `latest` only |

### Ubuntu image refs

Ubuntu accepts either a version number or a codename:

| Ref | Resolves to |
|---|---|
| `ubuntu:22.04` | Latest release build of 22.04 (Jammy Jellyfish) |
| `ubuntu:24.04` | Latest release build of 24.04 (Noble Numbat) |
| `ubuntu:25.10` | Latest release build of 25.10 (Questing Quokka) |
| `ubuntu:26.04` | Latest release build of 26.04 (Resolute Ratel) |
| `ubuntu:jammy` | Latest daily build of Jammy |
| `ubuntu:noble` | Latest daily build of Noble |
| `ubuntu:questing` | Latest daily build of Questing |
| `ubuntu:resolute` | Latest daily build of Resolute |

Version numbers use the stable releases path; codenames use the daily builds path.

### Debian image refs

Debian accepts either a codename or a major version number. Only actively
supported releases with published cloud images are accepted:

| Ref | Codename | Major version |
|---|---|---|
| `debian:bookworm` or `debian:12` | bookworm | 12 |
| `debian:trixie` or `debian:13` | trixie | 13 |

Codenames are preferred since Debian's directory structure is codename-based.
Passing a major version number resolves to the same image.

### Alpine image refs

Alpine versions must be a patch release within a supported branch:

| Branch | Example ref | EOL |
|---|---|---|
| `3.20` | `alpine:3.20.6` | 2026-05-01 |
| `3.21` | `alpine:3.21.3` | 2026-11-01 |
| `3.22` | `alpine:3.22.0` | 2027-05-01 |
| `3.23` | `alpine:3.23.4` | 2027-11-01 |

Bare branch strings (e.g. `alpine:3.23`) are also accepted.

### Fedora image refs

Fedora requires the full build string — the major version alone is not accepted:

| Ref | Major | Notes |
|---|---|---|
| `fedora:42-1.1` | 42 | Current supported release |
| `fedora:41-1.4` | 41 | Previous release (EOL Nov 2025; images still downloadable) |

### Arch Linux image refs

Arch is a rolling release. The only accepted version is `latest`:

```sh
image-cli pull vm arch:latest --arch amd64
```

Only `amd64` is officially published by the arch-boxes project.

## Windows container images

Must be run as Administrator (required for HCS layer import privileges).

| Image | Tag | Size | Use when |
|---|---|---|---|
| `mcr.microsoft.com/windows/nanoserver` | `ltsc2022`, `ltsc2019` | ~100 MB | .NET Core, minimal footprint |
| `mcr.microsoft.com/windows/servercore` | `ltsc2025`, `ltsc2022`, `ltsc2019`, `ltsc2016` | ~1.5 GB | .NET Framework, lift-and-shift |
| `mcr.microsoft.com/windows/server` | `ltsc2025`, `ltsc2022` | ~3.1 GB | Full Windows API, GPU, IIS |
| `mcr.microsoft.com/windows` | `ltsc2022`, `ltsc2019` | ~3.4 GB | Full Windows API + GDI/graphics libs |

`ltsc` tags are the stable production choice. The host OS version must be
compatible with the container image version — you cannot run an `ltsc2025`
container on a Windows Server 2019 host.

| Host OS | Recommended tag |
|---|---|
| Windows Server 2025 | `ltsc2025` |
| Windows Server 2022 | `ltsc2022` |
| Windows Server 2019 | `ltsc2019` |
| Windows Server 2016 | `ltsc2016` |

Nano Server is the smallest image but PowerShell, WMI, and the Windows
servicing stack are not included. Use Server Core if your workload needs
PowerShell.