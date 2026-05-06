


# 1. Init module
go mod init img2tar
go get github.com/diskfs/go-diskfs

# 2. Build
go build -o img-2-tar .

# 3. Run — auto-detect partition
./img-2-tar /Users/galaxy/.local/share/carbon/cloud.debian.org/debian/12/arm64/disk.img rootfs.tar

# 4. Explicit partition + verbose
./img-2-tar ubuntu.img rootfs.tar -p 2 -v