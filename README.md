# compute-image

Go library and CLI for pulling and managing Linux container images, Windows container images, and VM disk images. Designed for use in compute runtimes that need reproducible, cached, on-disk image layouts.

---

## Package usage

### Installation

```go
import compute_image "github.com/carbon-os/compute-image"
```

---

### Container images

#### Pull a Linux container image

```go
img, err := compute_image.Pull(compute_image.ContainerRef{
    Image: "ubuntu:24.04",
})
if err != nil {
    log.Fatal(err)
}

paths := compute_image.ContainerPathsFromImage(img)
// paths.Layers — extracted layer directories (overlayfs lower dirs)
// paths.Upper  — overlayfs upper (writable) dir
// paths.Work   — overlayfs work dir
```

Layers are unpacked into numbered subdirectories under `paths.Layers`. The caller is responsible for mounting overlayfs if a writable view is needed.

#### Pull a Windows container image

```go
img, err := compute_image.Pull(compute_image.ContainerRef{
    Image: "mcr.microsoft.com/windows/nanoserver:ltsc2022",
})
if err != nil {
    log.Fatal(err)
}

paths := compute_image.ContainerPathsFromImage(img)
// paths.Base    — read-only HCS layer directories
// paths.Scratch — writable scratch layer
```

Windows pulls use `hcsshim` and require the process to run as Administrator.

#### Resolve paths without pulling

```go
paths, err := compute_image.ResolveContainerPaths(compute_image.ContainerRef{
    Image: "ubuntu:24.04",
})
```

Useful for checking whether an image is already present before pulling.

#### Remove a container image

```go
err := compute_image.Remove(compute_image.ContainerRef{
    Image: "ubuntu:24.04",
})
```

Cache tarballs are left intact. On Windows, HCS layers are properly destroyed before the directory is removed.

---

### VM images

Supported distros: `ubuntu`, `debian`, `alpine`, `fedora`, `opensuse`, `rocky`, `arch`.  
Supported architectures: `amd64`, `arm64`, `arm`, `ppc64el`, `riscv64`.

Images are downloaded as qcow2, converted to raw, then written as a fixed VHD — no external tooling required.

#### Pull a VM image

```go
img, err := compute_image.Pull(compute_image.VMRef{
    Image: "ubuntu:24.04",
    Arch:  "amd64",
})
if err != nil {
    log.Fatal(err)
}

paths := compute_image.VMPathsFromImage(img)
// paths.Disk — ready-to-use fixed VHD
```

#### Pull with a custom registry

```go
img, err := compute_image.Pull(compute_image.VMRef{
    Image:    "alpine:3.19",
    Arch:     "amd64",
    Registry: "my.registry.example.com",
})
```

#### Resolve VM paths without pulling

```go
paths, err := compute_image.ResolveVMPaths(compute_image.VMRef{
    Image: "debian:bookworm",
    Arch:  "arm64",
})
```

#### Remove a VM image

```go
err := compute_image.Remove(compute_image.VMRef{
    Image: "ubuntu:24.04",
    Arch:  "amd64",
})
```

---

### Common options

Both `ContainerRef` and `VMRef` accept an optional `Dir` field to override the default storage root:

| Platform | Default |
|----------|---------|
| Linux / macOS | `$HOME/.local/share/carbon` |
| Windows | `%LOCALAPPDATA%\carbon` |

```go
compute_image.Pull(compute_image.ContainerRef{
    Image: "ubuntu:24.04",
    Dir:   "/mnt/images",
})
```

#### Remove everything

```go
err := compute_image.RemoveAll(compute_image.DefaultRootDir())
```

On Windows this properly destroys HCS layer locks and strips restrictive file attributes before removal.

---

### On-disk layout

**Container (Linux)**
```
$root/
  cache/                         # raw downloaded layer tarballs (shared)
  index.docker.io/library/ubuntu/24.04/
    layers/00, 01, …             # extracted layers (overlayfs lower dirs)
    upper/                       # overlayfs upper dir
    work/                        # overlayfs work dir
```

**Container (Windows)**
```
%LOCALAPPDATA%\carbon\
  cache\
  mcr.microsoft.com\windows\nanoserver\ltsc2022\
    base\00, 01, …               # read-only HCS layers
    scratch\                     # writable scratch layer + layerchain.json
```

**VM**
```
$root/
  cache/                         # downloaded qcow2 files (shared)
  <registry>/<name>/<version>/<arch>/
    disk.vhd                     # fixed VHD, ready to attach
```

---

## CLI

```
image-cli pull container <image>     [--dir <path>]
image-cli pull vm        <image>     --arch <arch> [--registry <host>] [--dir <path>]
image-cli info container <image>     [--dir <path>]
image-cli info vm        <image>     --arch <arch> [--registry <host>] [--dir <path>]
image-cli ls                         [--dir <path>]
image-cli rm             <image>     [--dir <path>]
image-cli rm-all                     [--dir <path>]
```

### Examples

```sh
# Pull a Linux container image
image-cli pull container ubuntu:24.04

# Pull a Windows container image (must run as Administrator)
image-cli pull container mcr.microsoft.com/windows/nanoserver:ltsc2022

# Pull a VM image
image-cli pull vm ubuntu:22.04   --arch amd64
image-cli pull vm ubuntu:noble   --arch arm64
image-cli pull vm debian:bookworm --arch amd64
image-cli pull vm alpine:3.19    --arch amd64 --registry my.registry.example.com

# Inspect resolved paths and pull status without downloading
image-cli info container ubuntu:24.04
image-cli info vm debian:12 --arch arm64

# List cached layer tarballs
image-cli ls

# Remove a specific image (cache preserved)
image-cli rm ubuntu:24.04

# Remove all image data including cache
image-cli rm-all
```