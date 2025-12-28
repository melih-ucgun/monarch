package transport

import (
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/melih-ucgun/monarch/internal/config"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

// SSHTransport, uzak sunucu ile iletişim kuran ana yapıdır.
type SSHTransport struct {
	Host   config.Host
	Client *ssh.Client
}

// NewSSHTransport, gelişmiş kimlik doğrulama yöntemlerini (Agent, Key, Password) deneyerek bağlantı kurar.
func NewSSHTransport(h config.Host) (*SSHTransport, error) {
	var authMethods []ssh.AuthMethod

	// 1. SSH Agent Desteği
	// Eğer sistemde aktif bir ssh-agent varsa (SSH_AUTH_SOCK tanımlıysa) kullanır.
	if socket := os.Getenv("SSH_AUTH_SOCK"); socket != "" {
		conn, err := net.Dial("unix", socket)
		if err == nil {
			authMethods = append(authMethods, ssh.PublicKeysCallback(agent.NewClient(conn).Signers))
		}
	}

	// 2. Private Key Desteği
	if h.KeyPath != "" {
		expandedPath := expandHome(h.KeyPath)
		key, err := os.ReadFile(expandedPath)
		if err != nil {
			return nil, fmt.Errorf("SSH anahtarı okunamadı (%s): %w", expandedPath, err)
		}

		var signer ssh.Signer
		if h.Passphrase != "" {
			signer, err = ssh.ParsePrivateKeyWithPassphrase(key, []byte(h.Passphrase))
		} else {
			signer, err = ssh.ParsePrivateKey(key)
		}

		if err != nil {
			return nil, fmt.Errorf("SSH anahtarı ayrıştırılamadı: %w", err)
		}
		authMethods = append(authMethods, ssh.PublicKeys(signer))
	}

	// 3. Şifre Desteği (Fallback)
	if h.Password != "" {
		authMethods = append(authMethods, ssh.Password(h.Password))
	}

	// Kimlik doğrulama yöntemi bulunamadıysa hata ver
	if len(authMethods) == 0 {
		return nil, fmt.Errorf("hiçbir SSH kimlik doğrulama yöntemi (Agent, Key veya Password) yapılandırılmadı")
	}

	clientConfig := &ssh.ClientConfig{
		User:            h.User,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // MVP için. Production'da known_hosts kontrolü yapılmalı.
	}

	// Adreste port belirtilmemişse varsayılan 22 portunu ekle
	addr := h.Address
	if !strings.Contains(addr, ":") {
		addr = addr + ":22"
	}

	client, err := ssh.Dial("tcp", addr, clientConfig)
	if err != nil {
		return nil, fmt.Errorf("SSH bağlantısı kurulamadı (%s): %w", addr, err)
	}

	return &SSHTransport{Host: h, Client: client}, nil
}

// RunRemote, uzak sunucuda bir komut çalıştırır ve çıktıyı yerel konsola (stdout/stderr) yönlendirir.
func (s *SSHTransport) RunRemote(command string) error {
	session, err := s.Client.NewSession()
	if err != nil {
		return fmt.Errorf("SSH session oluşturulamadı: %w", err)
	}
	defer session.Close()

	session.Stdout = os.Stdout
	session.Stderr = os.Stderr

	return session.Run(command)
}

// CopyFile, yerel bir dosyayı SCP protokolü kullanarak uzak sunucuya aktarır.
func (s *SSHTransport) CopyFile(localPath, remotePath string) error {
	file, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("yerel dosya açılamadı: %w", err)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return fmt.Errorf("dosya bilgileri alınamadı: %w", err)
	}

	session, err := s.Client.NewSession()
	if err != nil {
		return fmt.Errorf("SCP için session oluşturulamadı: %w", err)
	}
	defer session.Close()

	// Arka planda SCP protokolünü başlat
	go func() {
		w, err := session.StdinPipe()
		if err != nil {
			return
		}
		defer w.Close()

		// Protokol formatı: C[İzinler] [Boyut] [DosyaAdı]\n
		fmt.Fprintf(w, "C%04o %d %s\n", 0755, stat.Size(), filepath.Base(localPath))
		io.Copy(w, file)
		fmt.Fprint(w, "\x00") // Transfer sonu işareti
	}()

	// Uzak sunucuda scp'yi 'sink' modunda çalıştır
	remoteDir := filepath.Dir(remotePath)
	if err := session.Run(fmt.Sprintf("/usr/bin/scp -t %s", remoteDir)); err != nil {
		return fmt.Errorf("dosya kopyalama başarısız (scp -t %s): %w", remoteDir, err)
	}

	return nil
}

// expandHome, '~' ile başlayan yolları kullanıcının home dizini ile değiştirir.
func expandHome(path string) string {
	if len(path) > 0 && path[0] == '~' {
		home, err := os.UserHomeDir()
		if err == nil {
			return filepath.Join(home, path[1:])
		}
	}
	return path
}
