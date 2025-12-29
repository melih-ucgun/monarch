package transport

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/melih-ucgun/monarch/internal/config"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

type SSHTransport struct {
	client *ssh.Client
	host   config.Host
}

func NewSSHTransport(ctx context.Context, host config.Host) (*SSHTransport, error) {
	// Private Key Okuma
	key, err := os.ReadFile(host.SSHKeyPath)
	if err != nil {
		return nil, fmt.Errorf("ssh anahtarı okunamadı: %w", err)
	}

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, fmt.Errorf("ssh anahtarı parse edilemedi: %w", err)
	}

	sshConfig := &ssh.ClientConfig{
		User: host.User,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // DİKKAT: Prod ortamında known_hosts kullanılmalı
		Timeout:         10 * time.Second,
	}

	// HostName port içeriyor mu kontrol et, içermiyorsa ekle
	address := host.HostName
	if host.Port != 0 {
		address = fmt.Sprintf("%s:%d", host.HostName, host.Port)
	} else if _, _, err := net.SplitHostPort(address); err != nil {
		address = fmt.Sprintf("%s:22", address) // Varsayılan port
	}

	client, err := ssh.Dial("tcp", address, sshConfig)
	if err != nil {
		return nil, fmt.Errorf("ssh bağlantısı başarısız: %w", err)
	}

	return &SSHTransport{
		client: client,
		host:   host,
	}, nil
}

func (t *SSHTransport) Close() error {
	return t.client.Close()
}

func (t *SSHTransport) CopyFile(ctx context.Context, src, dst string) error {
	// SFTP Client Oluştur
	sftpClient, err := sftp.NewClient(t.client)
	if err != nil {
		return fmt.Errorf("sftp başlatılamadı: %w", err)
	}
	defer sftpClient.Close()

	// Kaynak dosyayı aç
	localFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("yerel dosya açılamadı: %w", err)
	}
	defer localFile.Close()

	// Hedef dizini oluştur (gerekirse)
	dstDir := filepath.Dir(dst)
	t.RunRemote(ctx, fmt.Sprintf("mkdir -p %s", dstDir))

	// Hedef dosyayı oluştur
	remoteFile, err := sftpClient.Create(dst)
	if err != nil {
		return fmt.Errorf("uzak dosya oluşturulamadı: %w", err)
	}
	defer remoteFile.Close()

	// Kopyala
	if _, err := io.Copy(remoteFile, localFile); err != nil {
		return fmt.Errorf("dosya kopyalanamadı: %w", err)
	}

	// Dosya izinlerini ayarla (Executable olması gerekebilir)
	remoteFile.Chmod(0755)

	return nil
}

func (t *SSHTransport) RunRemote(ctx context.Context, command string) error {
	session, err := t.client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	// Çıktıları yönlendir (Debug için)
	// session.Stdout = os.Stdout
	// session.Stderr = os.Stderr

	return session.Run(command)
}

// RunRemoteSecure: Sudo şifresini güvenli bir şekilde stdin'den verir.
// input: Eğer komut stdin bekliyorsa (örn: config dosyası içeriği) buraya verilir.
func (t *SSHTransport) RunRemoteSecure(ctx context.Context, command, sudoPassword, input string) error {
	session, err := t.client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	stdin, err := session.StdinPipe()
	if err != nil {
		return err
	}

	// stdout ve stderr'i yakala
	session.Stdout = os.Stdout
	session.Stderr = os.Stderr

	// Komutu başlat
	// Sudo şifresini sorması için -S parametresi kullanılır: echo 'password' | sudo -S command
	// Ancak biz daha güvenli olan stdin yöntemini kullanacağız.

	// Eğer sudo şifresi varsa komutu sudo ile sarmala
	finalCmd := command
	if sudoPassword != "" {
		finalCmd = fmt.Sprintf("sudo -S -p '' %s", command)
	}

	if err := session.Start(finalCmd); err != nil {
		return fmt.Errorf("komut başlatılamadı: %w", err)
	}

	// Goroutine ile stdin beslemesi yap
	go func() {
		defer stdin.Close()

		// Eğer sudo şifresi gerekiyorsa önce onu yaz
		if sudoPassword != "" {
			fmt.Fprintln(stdin, sudoPassword)
		}

		// Ardından asıl inputu (varsa) yaz
		if input != "" {
			fmt.Fprint(stdin, input)
		}
	}()

	// Komutun bitmesini bekle
	if err := session.Wait(); err != nil {
		// Exit code kontrolü yapılabilir
		return fmt.Errorf("komut hatası: %w", err)
	}

	return nil
}

func (t *SSHTransport) GetRemoteSystemInfo(ctx context.Context) (osName, arch string, err error) {
	session, err := t.client.NewSession()
	if err != nil {
		return "", "", err
	}
	defer session.Close()

	// GOOS ve GOARCH formatında çıktı almak için uname kullanıyoruz
	// Linux -> linux, x86_64 -> amd64, aarch64 -> arm64
	cmd := `uname -s | tr '[:upper:]' '[:lower:]'; uname -m`
	output, err := session.Output(cmd)
	if err != nil {
		return "", "", fmt.Errorf("sistem bilgisi alınamadı: %w", err)
	}

	var rawOS, rawArch string
	fmt.Sscanf(string(output), "%s\n%s", &rawOS, &rawArch)

	// Arch mapping
	arch = rawArch
	if rawArch == "x86_64" {
		arch = "amd64"
	} else if rawArch == "aarch64" {
		arch = "arm64"
	}

	return rawOS, arch, nil
}

func (t *SSHTransport) DownloadFile(ctx context.Context, remotePath, localPath string) error {
	sftpClient, err := sftp.NewClient(t.client)
	if err != nil {
		return err
	}
	defer sftpClient.Close()

	remoteFile, err := sftpClient.Open(remotePath)
	if err != nil {
		return err
	}
	defer remoteFile.Close()

	localFile, err := os.Create(localPath)
	if err != nil {
		return err
	}
	defer localFile.Close()

	_, err = remoteFile.WriteTo(localFile)
	return err
}
