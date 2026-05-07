# Supported VM Images

## Alpine

| Version | Arch  |
|---------|-------|
| 3.20    | amd64 |
| 3.20    | arm64 |
| 3.21    | amd64 |
| 3.21    | arm64 |
| 3.22    | amd64 |
| 3.22    | arm64 |
| 3.23    | amd64 |
| 3.23    | arm64 |

> Branch refs are accepted (e.g. `alpine:3.23`) and resolve to the latest patch automatically. Full patch versions also work (e.g. `alpine:3.23.4`). 3.20 reached EOL 2026-05-01 but images remain downloadable.

---

## Debian

| Version | Codename | Arch  |
|---------|----------|-------|
| 12      | bookworm | amd64 |
| 12      | bookworm | arm64 |
| 13      | trixie   | amd64 |
| 13      | trixie   | arm64 |

> Both numeric and codename forms are accepted (e.g. `debian:12` or `debian:bookworm`).

---

## Ubuntu

| Version | Codename  | Arch  |
|---------|-----------|-------|
| 22.04   | jammy     | amd64 |
| 22.04   | jammy     | arm64 |
| 24.04   | noble     | amd64 |
| 24.04   | noble     | arm64 |
| 25.10   | questing  | amd64 |
| 25.10   | questing  | arm64 |
| 26.04   | resolute  | amd64 |
| 26.04   | resolute  | arm64 |

> Both numeric and codename forms are accepted (e.g. `ubuntu:24.04` or `ubuntu:noble`).

---

## Fedora

| Version | Arch  |
|---------|-------|
| 41      | amd64 |
| 41      | arm64 |
| 42      | amd64 |
| 42      | arm64 |

> A full build string is required (e.g. `fedora:42-1.1`), not just the major number.

---

## openSUSE

| Version     | Arch  |
|-------------|-------|
| tumbleweed  | amd64 |
| tumbleweed  | arm64 |
| 15.6        | amd64 |
| 15.6        | arm64 |
| 16.0        | amd64 |
| 16.0        | arm64 |

---

## Rocky Linux

| Version | Arch  |
|---------|-------|
| 8       | amd64 |
| 8       | arm64 |
| 9       | amd64 |
| 9       | arm64 |
| 10      | amd64 |
| 10      | arm64 |

> Patch versions are accepted (e.g. `rocky:9.7`); only the major is used to resolve the download.

---

## AlmaLinux

| Version | Arch  |
|---------|-------|
| 8       | amd64 |
| 8       | arm64 |
| 9       | amd64 |
| 9       | arm64 |
| 10      | amd64 |
| 10      | arm64 |

> AlmaLinux 10 amd64 uses the `x86_64_v2` microarch baseline. Older hypervisors may not support this.

---

## CentOS Stream

| Version | Arch  |
|---------|-------|
| 9       | amd64 |
| 9       | arm64 |
| 10      | amd64 |
| 10      | arm64 |

---

## Arch Linux

| Version | Arch  |
|---------|-------|
| latest  | amd64 |

> Arch Linux is a rolling release. `arm64` is not officially published via arch-boxes.