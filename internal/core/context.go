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
	OS      string // runtime.GOOS (linux, darwin)
	Distro  string // ubuntu, arch, fedora
	Version string // 22.04, 38, rolling

	// Kullanıcı Bilgileri
	User    string // Mevcut kullanıcı
	HomeDir string // Kullanıcının ev dizini

	// Çalışma Modu
	DryRun bool // Eğer true ise, hiçbir değişiklik yapılmaz, sadece simüle edilir.

	// Logger veya Output (İleride loglama için)
	Stdout io.Writer
	Stderr io.Writer
}

// NewSystemContext, temel bir context oluşturur.
// Not: Distro tespiti logic'i buraya değil, 'internal/system' paketine gelecek.
// Burası sadece veri yapısını tutar.
func NewSystemContext(dryRun bool) *SystemContext {
	return &SystemContext{
		Context: context.Background(),
		OS:      "unknown", // Doldurulacak
		Distro:  "unknown", // Doldurulacak
		User:    os.Getenv("USER"),
		HomeDir: os.Getenv("HOME"),
		DryRun:  dryRun,
		Stdout:  os.Stdout,
		Stderr:  os.Stderr,
	}
}
