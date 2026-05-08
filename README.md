# compute-image

A Go library and CLI tool for pulling, converting, and managing VM and container images across Linux and Windows.

## Overview

`compute-image` provides a unified interface for working with two image types:

- **Container images** — pulled from OCI registries, unpacked as overlayfs layers (Linux) or HCS layers (Windows)
- **VM images** — downloaded as QCOW2, converted to raw `.img` (Linux) or fixed VHD (Windows), with optional kernel/initrd extraction

## Installation

```bash
go get github.com/carbon-os/compute-image
```

## Quick Start

```go
package main

import (
    "fmt"
    "log"

    compute_image "github.com/carbon-os/compute-image"
)

func main() {
    img, err := compute_image.Pull(compute_image.VMRef{
        Image: "ubuntu:24.04",
        Arch:  "amd64",
    })
    if err != nil {
        log.Fatal(err)
    }

    paths := compute_image.VMPathsFromImage(img)
    fmt.Println("Disk:", paths.Disk)
}
```

## Package Reference

### Pull

Downloads and prepares a VM or container image. Returns an opaque image value on success.

```go
img, err := compute_image.Pull(ref)
```

**VM image:**

```go
img, err := compute_image.Pull(compute_image.VMRef{
    Image:         "ubuntu:24.04",   // name:version or name:codename
    Arch:          "amd64",          // required: "amd64" or "arm64"
    Registry:      "",               // optional: uses distro default if empty
    Dir:           "",               // optional: uses platform default if empty
    ExtractKernel: true,             // extract vmlinuz + initrd alongside the disk
})
```

**Container image:**

```go
img, err := compute_image.Pull(compute_image.ContainerRef{
    Image: "ubuntu:24.04",  // OCI image ref
    Dir:   "",              // optional: uses platform default if empty
})
```

---

### VMPathsFromImage / ContainerPathsFromImage

Extract resolved on-disk paths from a pulled image value.

```go
paths := compute_image.VMPathsFromImage(img)
fmt.Println(paths.Dir)   // image directory
fmt.Println(paths.Disk)  // disk.img (Linux) or disk.vhd (Windows)
fmt.Println(paths.Cache) // cache directory
```

```go
paths := compute_image.ContainerPathsFromImage(img)
fmt.Println(paths.Dir)     // image directory

// Linux only
fmt.Println(paths.Layers)  // extracted OCI layer subdirectories
fmt.Println(paths.Upper)   // overlayfs upper (writable) dir
fmt.Println(paths.Work)    // overlayfs work dir

// Windows only
fmt.Println(paths.Base)    // HCS read-only layer subdirectories
fmt.Println(paths.Scratch) // writable HCS scratch layer

fmt.Println(paths.Cache)   // raw downloaded layer tarballs
```

---

### ResolveVMPaths / ResolveContainerPaths

Resolve on-disk paths for a ref without pulling anything. Useful for checking
whether an image is already present before deciding to pull.

```go
paths, err := compute_image.ResolveVMPaths(compute_image.VMRef{
    Image: "ubuntu:24.04",
    Arch:  "amd64",
})

paths, err := compute_image.ResolveContainerPaths(compute_image.ContainerRef{
    Image: "ubuntu:24.04",
})
```

---

### Remove

Deletes all on-disk data for an image. Cache files are left intact.

```go
err := compute_image.Remove(compute_image.VMRef{
    Image: "ubuntu:24.04",
    Arch:  "amd64",
})

err := compute_image.Remove(compute_image.ContainerRef{
    Image: "ubuntu:24.04",
})
```

---

### RemoveAll

Removes an entire root directory, including all images and cache. On Windows this
correctly tears down HCS layer locks and strips restrictive file attributes before
deletion.

```go
err := compute_image.RemoveAll("/path/to/root")
```

---

### DefaultRootDir

Returns the platform default root directory (`$HOME/.local/share/carbon` on Linux,
`%LOCALAPPDATA%\carbon` on Windows).

```go
dir := compute_image.DefaultRootDir()
```

---

### HumanBytes

Formats a byte count as a human-readable string (e.g. `"1.4 GB"`).

```go
s := compute_image.HumanBytes(1_400_000_000)
```

## Ref Types

**`VMRef`**

| Field           | Type     | Description                                                   |
|-----------------|----------|---------------------------------------------------------------|
| `Image`         | `string` | Image ref, e.g. `"ubuntu:24.04"`                              |
| `Arch`          | `string` | Target architecture: `"amd64"` or `"arm64"` (required)        |
| `Registry`      | `string` | Registry hostname — uses distro default if empty              |
| `Dir`           | `string` | Root directory for image data — uses platform default if empty |
| `ExtractKernel` | `bool`   | Extract `vmlinuz` and `initrd` alongside the disk image       |

**`ContainerRef`**

| Field   | Type     | Description                                                        |
|---------|----------|--------------------------------------------------------------------|
| `Image` | `string` | OCI image ref, e.g. `"ubuntu:24.04"`                               |
| `Dir`   | `string` | Root directory for image data — uses platform default if empty     |

## Supported VM Images

| Distro        | Versions                   | Architectures |
|---------------|----------------------------|---------------|
| Alpine        | 3.20, 3.21, 3.22, 3.23     | amd64, arm64  |
| Debian        | 12 (bookworm), 13 (trixie) | amd64, arm64  |
| Ubuntu        | 22.04, 24.04, 25.10, 26.04 | amd64, arm64  |
| Fedora        | 41, 42                     | amd64, arm64  |
| openSUSE      | tumbleweed, 15.6, 16.0     | amd64, arm64  |
| Rocky Linux   | 8, 9, 10                   | amd64, arm64  |
| AlmaLinux     | 8, 9, 10                   | amd64, arm64  |
| CentOS Stream | 9, 10                      | amd64, arm64  |
| Arch Linux    | latest                     | amd64         |

Image refs accept both numeric versions and codenames where applicable
(e.g. `ubuntu:24.04` or `ubuntu:noble`, `debian:12` or `debian:bookworm`).
Fedora requires a full build string (e.g. `fedora:42-1.1`).

## Storage Layout

### Default root directories

| Platform | Path                        |
|----------|-----------------------------|
| Linux    | `$HOME/.local/share/carbon` |
| Windows  | `%LOCALAPPDATA%\carbon`     |

### VM images

```
<root>/<registry>/<name>/<version>/<arch>/
    disk.img          # Linux
    disk.vhd          # Windows (fixed VHD)
    vmlinuz           # if ExtractKernel was set
    initrd            # if ExtractKernel was set
<root>/cache/
    <name>-<version>-<arch>.qcow2
```

### Container images — Linux

```
<root>/<registry>/<repo>/<tag>/
    layers/00, 01, …  # extracted OCI layers (overlayfs lower dirs)
    upper/            # overlayfs upper dir (writable, not mounted here)
    work/             # overlayfs work dir
<root>/cache/
    <digest>.tar.gz
```

### Container images — Windows

```
<root>/<registry>/<repo>/<tag>/
    base/00, 01, …    # imported HCS read-only layers
    scratch/          # writable HCS scratch layer
        layerchain.json
<root>/cache/
    <digest>.tar.gz
```

## Platform Notes

- **Linux VM images** are stored as raw `.img` files, ready for use with QEMU/KVM or any compatible hypervisor.
- **Windows VM images** are stored as fixed VHD files, compatible with Hyper-V.
- **Linux containers** use overlayfs; the caller is responsible for mounting the layer stack with `paths.Layers` as the lower dirs, `paths.Upper` as the upper dir, and `paths.Work` as the work dir.
- **Windows containers** use the HCS (Host Compute Service) layer model via `hcsshim`. Pulling requires Administrator privileges. Removal goes through `hcsshim.DestroyLayer` to correctly release layer locks before deleting files.

## CLI

A standalone CLI is available for manual image management.

### Install

```bash
go install github.com/carbon-os/compute-image/cmd/image-cli@latest
```

### Commands

```
image-cli pull container <image>   [--dir <path>]
image-cli pull vm        <image>   --arch <arch> [--registry <host>] [--dir <path>] [--extract-kernel]
image-cli info container <image>   [--dir <path>]
image-cli info vm        <image>   --arch <arch> [--registry <host>] [--dir <path>]
image-cli ls                       [--dir <path>]
image-cli rm             <image>   [--dir <path>]
image-cli rm-all                   [--dir <path>]
```

### Examples

```bash
image-cli pull vm ubuntu:22.04  --arch amd64 --extract-kernel
image-cli pull vm ubuntu:noble  --arch arm64 --extract-kernel
image-cli pull vm debian:13     --arch amd64 --extract-kernel
image-cli pull vm alpine:3.22   --arch arm64 --extract-kernel
image-cli pull vm fedora:42     --arch amd64 --extract-kernel
image-cli pull vm rocky:9       --arch amd64 --extract-kernel
image-cli pull vm arch:latest   --arch amd64 --extract-kernel

image-cli pull container ubuntu:24.04
image-cli pull container mcr.microsoft.com/windows/nanoserver:ltsc2022

image-cli info vm ubuntu:22.04 --arch amd64
image-cli info container ubuntu:24.04

image-cli ls
image-cli rm ubuntu:22.04
image-cli rm-all
```

## License

MIT