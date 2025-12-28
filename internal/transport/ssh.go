package transport

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/melih-ucgun/monarch/internal/config"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

// SSHTransport, uzak sunucu ile tüm iletişimi yöneten yapıdır.
type SSHTransport struct {
	client *ssh.Client
	host   config.Host
}

// NewSSHTransport, verilen host konfigürasyonuna göre güvenli bir SSH bağlantısı açar.
func NewSSHTransport(h config.Host) (*SSHTransport, error) {
	var authMethods []ssh.AuthMethod

	// 1. Şifre tabanlı yetkilendirme ekle
	if h.Password != "" {
		authMethods = append(authMethods, ssh.Password(h.Password))
	}

	// 2. Known Hosts dosyasını bul (Genellikle ~/.ssh/known_hosts)
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("ana dizin bulunamadı: %w", err)
	}
	knownHostsPath := filepath.Join(homeDir, ".ssh", "known_hosts")

	// 3. Host Key Callback oluştur (Bağlanılan sunucuyu doğrular)
	hostKeyCallback, err := knownhosts.New(knownHostsPath)
	if err != nil {
		// Eğer dosya yoksa, kullanıcıyı uyar ama güvenli olmayan bir fallback sunma
		return nil, fmt.Errorf("known_hosts dosyası yüklenemedi (%s): %w. Lütfen sunucuya önce manuel ssh ile bağlanıp anahtarı kaydedin", knownHostsPath, err)
	}

	clientConfig := &ssh.ClientConfig{
		User:            h.User,
		Auth:            authMethods,
		HostKeyCallback: hostKeyCallback, // Artık güvenli doğrulama kullanıyoruz
		Timeout:         15 * time.Second,
	}

	// Port kontrolü
	port := h.Port
	if port == 0 {
		port = 22
	}

	addr := fmt.Sprintf("%s:%d", h.Address, port)
	client, err := ssh.Dial("tcp", addr, clientConfig)
	if err != nil {
		// Eğer hata bir "host key mismatch" ise bunu kullanıcıya net belirtmeliyiz
		return nil, fmt.Errorf("SSH bağlantısı kurulamadı. Sunucu kimliği doğrulanamadı veya bağlantı reddedildi: %w", err)
	}

	return &SSHTransport{client: client, host: h}, nil
}

// GetRemoteSystemInfo, uzak sunucunun OS ve ARCH bilgisini döner.
func (t *SSHTransport) GetRemoteSystemInfo() (string, string, error) {
	osOut, err := t.CaptureRemoteOutput("uname -s")
	if err != nil {
		return "", "", err
	}
	remoteOS := strings.ToLower(strings.TrimSpace(osOut))

	archOut, err := t.CaptureRemoteOutput("uname -m")
	if err != nil {
		return "", "", err
	}
	rawArch := strings.TrimSpace(archOut)

	remoteArch := "amd64"
	switch rawArch {
	case "x86_64":
		remoteArch = "amd64"
	case "aarch64", "arm64":
		remoteArch = "arm64"
	case "i386", "i686":
		remoteArch = "386"
	}

	return remoteOS, remoteArch, nil
}

// CopyFile, yerel bir dosyayı uzak sunucudaki hedefe kopyalar.
func (t *SSHTransport) CopyFile(srcPath, destPath string) error {
	f, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer f.Close()

	stat, _ := f.Stat()
	session, err := t.client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	go func() {
		w, _ := session.StdinPipe()
		defer w.Close()
		fmt.Fprintf(w, "C%04o %d %s\n", stat.Mode().Perm(), stat.Size(), "file")
		io.Copy(w, f)
		fmt.Fprint(w, "\x00")
	}()

	if err := session.Run(fmt.Sprintf("scp -t %s", destPath)); err != nil {
		return fmt.Errorf("dosya kopyalama hatası: %w", err)
	}

	return nil
}

// CaptureRemoteOutput, uzak sunucuda bir komut çalıştırır ve stdout çıktısını döner.
func (t *SSHTransport) CaptureRemoteOutput(cmd string) (string, error) {
	session, err := t.client.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()

	var stdout bytes.Buffer
	session.Stdout = &stdout

	if err := session.Run(cmd); err != nil {
		return "", err
	}

	return stdout.String(), nil
}

// RunRemoteSecure, uzak sunucuda sudo destekli komut çalıştırır.
func (t *SSHTransport) RunRemoteSecure(cmd string, sudoPassword string) error {
	session, err := t.client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	session.Stdout = os.Stdout
	session.Stderr = os.Stderr

	if sudoPassword != "" {
		in, err := session.StdinPipe()
		if err != nil {
			return err
		}

		fullCmd := fmt.Sprintf("sudo -S -p '' %s", cmd)
		if err := session.Start(fullCmd); err != nil {
			return err
		}

		fmt.Fprintln(in, sudoPassword)
		return session.Wait()
	}

	return session.Run(cmd)
}

// Close, SSH bağlantısını güvenli bir şekilde kapatır.
func (t *SSHTransport) Close() error {
	if t.client != nil {
		return t.client.Close()
	}
	return nil
}
