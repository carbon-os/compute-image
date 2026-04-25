# image-cli

Command-line tool for pulling and managing images for Carbon Compute. Handles
both Windows container images (HCS/ociwclayer) and Linux VM disk images
(qcow2 → raw → ext4).

## Usage

```
image-cli pull container <image> [--dir <path>]
image-cli pull vm        <image> --registry <host> [--dir <path>]
image-cli info container <image> [--dir <path>]
image-cli info vm        <image> --registry <host> [--dir <path>]
image-cli ls             [--dir <path>]
image-cli rm             <image> [--dir <path>]
```

## Examples

```powershell
# Pull a Windows container image (paths resolved automatically)
image-cli pull container mcr.microsoft.com/windows/nanoserver:ltsc2022

# Pull to a specific directory
image-cli pull container mcr.microsoft.com/windows/nanoserver:ltsc2022 --dir C:\images

# Pull a Ubuntu VM image
image-cli pull vm ubuntu:22.04 --registry cloud-images.ubuntu.com

# Pull a VM image to a specific directory
image-cli pull vm ubuntu:22.04 --registry cloud-images.ubuntu.com --dir C:\images

# Inspect paths and status without pulling anything
image-cli info container mcr.microsoft.com/windows/nanoserver:ltsc2022
image-cli info vm ubuntu:22.04 --registry cloud-images.ubuntu.com

# List all cached layers and images
image-cli ls

# List cache in a specific directory
image-cli ls --dir C:\images

# Remove a cached image
image-cli rm ubuntu:22.04
```

## pull container

Downloads and imports a Windows container image. All paths are derived from
`--dir` and the image reference — you never need to specify base or scratch
locations manually.

```
image-cli pull container <image> [--dir <path>]
```

Output:

```
[*] Pulling container image: mcr.microsoft.com/windows/nanoserver:ltsc2022
    Dir: C:\Users\you\AppData\Local\carbon\mcr.microsoft.com\windows\nanoserver\ltsc2022
...
[+] Container image ready.
    Dir:     C:\Users\you\AppData\Local\carbon\mcr.microsoft.com\windows\nanoserver\ltsc2022
    Base:    C:\Users\you\AppData\Local\carbon\mcr.microsoft.com\windows\nanoserver\ltsc2022\base
    Scratch: C:\Users\you\AppData\Local\carbon\mcr.microsoft.com\windows\nanoserver\ltsc2022\scratch
    Cache:   C:\Users\you\AppData\Local\carbon\cache
```

Must be run as Administrator — required for HCS layer import privileges.

## pull vm

Downloads and prepares a Linux VM disk image.

```
image-cli pull vm <image> --registry <host> [--dir <path>]
```

Output:

```
[*] Pulling VM image: ubuntu:22.04 from cloud-images.ubuntu.com
    Dir: C:\Users\you\AppData\Local\carbon\cloud-images.ubuntu.com\ubuntu\22.04
...
[+] VM image ready.
    Dir:   C:\Users\you\AppData\Local\carbon\cloud-images.ubuntu.com\ubuntu\22.04
    Disk:  C:\Users\you\AppData\Local\carbon\cloud-images.ubuntu.com\ubuntu\22.04\disk.raw
    Cache: C:\Users\you\AppData\Local\carbon\cache
```

## info

Shows all resolved paths and current status for an image without downloading
or importing anything. Useful for checking what's on disk or getting paths
to pass to other tools.

```
image-cli info container <image> [--dir <path>]
image-cli info vm        <image> --registry <host> [--dir <path>]
```

Container output:

```
Image:   mcr.microsoft.com/windows/nanoserver:ltsc2022
Dir:     C:\Users\you\AppData\Local\carbon\mcr.microsoft.com\windows\nanoserver\ltsc2022
Base:    C:\Users\you\AppData\Local\carbon\mcr.microsoft.com\windows\nanoserver\ltsc2022\base
Scratch: C:\Users\you\AppData\Local\carbon\mcr.microsoft.com\windows\nanoserver\ltsc2022\scratch
Cache:   C:\Users\you\AppData\Local\carbon\cache
Status:  ready
```

VM output:

```
Image:    ubuntu:22.04
Registry: cloud-images.ubuntu.com
Dir:      C:\Users\you\AppData\Local\carbon\cloud-images.ubuntu.com\ubuntu\22.04
Disk:     C:\Users\you\AppData\Local\carbon\cloud-images.ubuntu.com\ubuntu\22.04\disk.raw
Cache:    C:\Users\you\AppData\Local\carbon\cache
Status:   ready
```

Status values:

| Status | Meaning |
|---|---|
| `ready` | Image is fully prepared and ready to use |
| `cached` | Raw layers are downloaded but not yet imported or converted |
| `not pulled` | Nothing on disk yet |

## Passing paths to compute-container

The base and scratch paths from `pull container` map directly to
`compute-container`'s `--base` and `--scratch` flags. Use `info` to get
the exact paths for your machine:

```powershell
image-cli info container mcr.microsoft.com/windows/nanoserver:ltsc2022
# copy Base and Scratch from the output, then:

container-cli run `
    --base    C:\Users\you\AppData\Local\carbon\mcr.microsoft.com\windows\nanoserver\ltsc2022\base `
    --scratch C:\Users\you\AppData\Local\carbon\mcr.microsoft.com\windows\nanoserver\ltsc2022\scratch `
    -- cmd.exe
```

## Windows container images

All Windows base images are pulled from `mcr.microsoft.com`. Must be run as
Administrator (required for HCS layer import privileges).

| Image | Tag | Size | Use when |
|---|---|---|---|
| `mcr.microsoft.com/windows/nanoserver` | `ltsc2022`, `ltsc2019` | ~100 MB | .NET Core, minimal footprint |
| `mcr.microsoft.com/windows/servercore` | `ltsc2025`, `ltsc2022`, `ltsc2019`, `ltsc2016` | ~1.5 GB | .NET Framework, lift-and-shift |
| `mcr.microsoft.com/windows/server` | `ltsc2025`, `ltsc2022` | ~3.1 GB | Full Windows API, GPU, IIS |
| `mcr.microsoft.com/windows` | `ltsc2022`, `ltsc2019` | ~3.4 GB | Full Windows API + GDI/graphics libs |

### Nano Server

Nano Server is the smallest image but has a significantly reduced API surface.
**PowerShell, WMI, and the Windows servicing stack are not included.** It is
suitable for .NET Core applications and modern open source frameworks only. If
your workload needs PowerShell, use Server Core instead.

### Choosing a tag

`ltsc` tags (Long-Term Servicing Channel) are the stable production choice.
The host OS version must be compatible with the container image version — you
cannot run an `ltsc2025` container on a Windows Server 2019 host.

| Host OS | Recommended tag |
|---|---|
| Windows Server 2025 | `ltsc2025` |
| Windows Server 2022 | `ltsc2022` |
| Windows Server 2019 | `ltsc2019` |
| Windows Server 2016 | `ltsc2016` |

## Cache

Downloaded layers are cached by digest and reused across pulls. The default
cache location is `%LOCALAPPDATA%\carbon\cache`. Override with `--dir`.

Re-running `pull container` for the same image is a no-op if all layers are
already cached — only the HCS import and scratch creation run again.