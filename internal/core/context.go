package core

import (
	"context"
	"io"
	"os"
)

// SystemContext, uygulamanın çalışma anındaki bağlamını (context) tutar.
// Standart Go "context" paketini sarmalar ve Monarch'a özel alanlar ekler.
type SystemContext struct {
	context.Context

	// İşletim Sistemi Bilgileri
	OS       string // runtime.GOOS (linux, darwin)
	Distro   string // ubuntu, arch, fedora
	Version  string // 22.04, 38, rolling
	Hostname string // Makine adı

	// Donanım Bilgileri
	Hardware SystemHardware

	// Çevresel Değişkenler
	Env SystemEnv

	// Dosya Sistemi
	FS SystemFS

	// Kullanıcı Bilgileri
	User    string // Mevcut kullanıcı
	HomeDir string // Kullanıcının ev dizini
	UID     string // User ID
	GID     string // Group ID

	// Çalışma Modu
	DryRun bool // Eğer true ise, hiçbir değişiklik yapılmaz, sadece simüle edilir.

	// Logger veya Output (İleride loglama için)
	Stdout io.Writer
	Stderr io.Writer
}

type SystemHardware struct {
	CPUModel  string // "AMD Ryzen 7 5800X"
	CPUCore   int    // Çekirdek sayısı
	RAMTotal  string // "16GB"
	GPUVendor string // "NVIDIA", "AMD", "Intel"
	GPUModel  string // "RTX 3070"
}

type SystemEnv struct {
	Shell    string // "/bin/zsh"
	Lang     string // "en_US.UTF-8"
	Term     string // "xterm-256color"
	Timezone string // "Europe/Istanbul"
}

type SystemFS struct {
	RootFSType string // "ext4", "btrfs", "zfs"
}

// NewSystemContext, temel bir context oluşturur.
func NewSystemContext(dryRun bool) *SystemContext {
	return &SystemContext{
		Context: context.Background(),
		OS:      "unknown",
		Distro:  "unknown",
		User:    os.Getenv("USER"),
		HomeDir: os.Getenv("HOME"),
		DryRun:  dryRun,
		Stdout:  os.Stdout,
		Stderr:  os.Stderr,
		// Diğer alt structlar zero-value olarak başlar, detector tarafından doldurulur.
	}
}
