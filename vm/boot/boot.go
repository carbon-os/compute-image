// vm/boot/boot.go
package boot

type Config struct {
    Partition  int
    BootDir    string
    KernelGlob string
    InitrdGlob string
}