package transport

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/melih-ucgun/monarch/internal/config"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

// RetryConfig
const (
	MaxRetries = 3
	BaseDelay  = 2 * time.Second
)

type SSHTransport struct {
	client *ssh.Client
	config *ssh.ClientConfig
}

// retry, context iptal edilirse beklemeden çıkar, yoksa backoff uygular.
func retry(ctx context.Context, operationName string, operation func() error) error {
	var err error
	delay := BaseDelay

	for i := 0; i < MaxRetries; i++ {
		// Context iptal edilmiş mi kontrol et (döngü başında)
		if ctx.Err() != nil {
			return ctx.Err()
		}

		err = operation()
		if err == nil {
			return nil
		}

		if i == MaxRetries-1 {
			break
		}

		slog.Warn("İşlem başarısız, tekrar deneniyor...",
			"islem", operationName,
			"deneme", i+1,
			"hata", err,
		)

		// Bekleme süresi boyunca Context iptalini dinle
		select {
		case <-ctx.Done():
			return ctx.Err() // Kullanıcı iptal ettiyse bekleme yapma
		case <-time.After(delay):
			delay *= 2
		}
	}

	return fmt.Errorf("%s başarısız oldu: %w", operationName, err)
}

// NewSSHTransport artık Context kabul ediyor ve bağlantı sırasında bunu gözetiyor.
func NewSSHTransport(ctx context.Context, host config.Host) (*SSHTransport, error) {
	sshConfig := &ssh.ClientConfig{
		User:            host.User,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second, // TCP seviyesinde timeout
	}

	home, err := os.UserHomeDir()
	if err == nil {
		if hk, err := knownhosts.New(filepath.Join(home, ".ssh", "known_hosts")); err == nil {
			sshConfig.HostKeyCallback = hk
		}
	}

	authMethods := []ssh.AuthMethod{}

	// 1. SSH Agent
	if socket := os.Getenv("SSH_AUTH_SOCK"); socket != "" {
		if conn, err := net.Dial("unix", socket); err == nil {
			conn.Close()
		}
	}

	// 2. Private Key
	keyPath := filepath.Join(home, ".ssh", "id_rsa")
	if key, err := os.ReadFile(keyPath); err == nil {
		if signer, err := ssh.ParsePrivateKey(key); err == nil {
			authMethods = append(authMethods, ssh.PublicKeys(signer))
		}
	} else if host.Password != "" {
		authMethods = append(authMethods, ssh.Password(host.Password))
	} else {
		// Şifre sorma kısmını basitleştirdik, buraya interaktif logic eklenebilir
		// Context ile uyumlu olması için burada uzun süre bloklamamak iyi olur
	}

	sshConfig.Auth = authMethods
	addr := fmt.Sprintf("%s:%d", host.Address, host.Port)

	var client *ssh.Client

	// Bağlantıyı retry ve context ile sarıyoruz
	connectErr := retry(ctx, "SSH Bağlantısı", func() error {
		// ssh.Dial yerine net.Dialer kullanıp context'i TCP seviyesine indiriyoruz
		d := net.Dialer{Timeout: sshConfig.Timeout}
		conn, err := d.DialContext(ctx, "tcp", addr)
		if err != nil {
			return err
		}

		c, chans, reqs, err := ssh.NewClientConn(conn, addr, sshConfig)
		if err != nil {
			conn.Close()
			return err
		}
		client = ssh.NewClient(c, chans, reqs)
		return nil
	})

	if connectErr != nil {
		return nil, connectErr
	}

	return &SSHTransport{client: client, config: sshConfig}, nil
}

func (t *SSHTransport) Close() error {
	if t.client != nil {
		return t.client.Close()
	}
	return nil
}

func (t *SSHTransport) GetRemoteSystemInfo(ctx context.Context) (string, string, error) {
	// Basit komutlar için RunRemoteSecure logic'ini veya direkt session kullanabiliriz.
	// Hızlı olduğu için direkt çağırıyoruz ama session başlatma context kontrolü yapılabilir.
	if ctx.Err() != nil {
		return "", "", ctx.Err()
	}

	session, err := t.client.NewSession()
	if err != nil {
		return "", "", err
	}
	defer session.Close()

	var b bytes.Buffer
	session.Stdout = &b
	if err := session.Run("uname -s && uname -m"); err != nil {
		return "", "", err
	}

	parts := strings.Split(strings.TrimSpace(b.String()), "\n")
	if len(parts) < 2 {
		return "linux", "amd64", nil // Fallback
	}

	osName := strings.ToLower(parts[0])
	arch := strings.ToLower(parts[1])
	if arch == "x86_64" {
		arch = "amd64"
	} else if arch == "aarch64" {
		arch = "arm64"
	}
	return osName, arch, nil
}

func (t *SSHTransport) CopyFile(ctx context.Context, localPath, remotePath string) error {
	return retry(ctx, fmt.Sprintf("Dosya Kopyalama (%s)", filepath.Base(localPath)), func() error {
		// Context iptal kontrolü
		if ctx.Err() != nil {
			return ctx.Err()
		}

		f, err := os.Open(localPath)
		if err != nil {
			return err
		}
		defer f.Close()

		stat, err := f.Stat()
		if err != nil {
			return err
		}

		session, err := t.client.NewSession()
		if err != nil {
			return err
		}
		defer session.Close()

		go func() {
			w, _ := session.StdinPipe()
			defer w.Close()
			fmt.Fprintln(w, "C0"+fmt.Sprintf("%o", stat.Mode().Perm()), stat.Size(), filepath.Base(remotePath))
			io.Copy(w, f)
			fmt.Fprint(w, "\x00")
		}()

		// SCP komutunu context ile bekleyelim (basit implementasyon)
		done := make(chan error, 1)
		go func() {
			done <- session.Run(fmt.Sprintf("scp -t %s", filepath.Dir(remotePath)))
		}()

		select {
		case err := <-done:
			return err
		case <-ctx.Done():
			session.Close() // Bağlantıyı kopar
			return ctx.Err()
		}
	})
}

func (t *SSHTransport) DownloadFile(ctx context.Context, remotePath, localPath string) error {
	return retry(ctx, fmt.Sprintf("Dosya İndirme (%s)", remotePath), func() error {
		if ctx.Err() != nil {
			return ctx.Err()
		}

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

		// io.Copy iptal edilemez, bu yüzden küçük buffer'larla okuyup context kontrolü yapan bir döngü yazmak en iyisidir
		// Ancak pratiklik adına bir 'goroutine wrapper' kullanabiliriz.
		type copyResult struct {
			n   int64
			err error
		}
		done := make(chan copyResult, 1)

		go func() {
			n, err := io.Copy(localFile, remoteFile)
			done <- copyResult{n, err}
		}()

		select {
		case res := <-done:
			if res.err == nil {
				return localFile.Sync()
			}
			return res.err
		case <-ctx.Done():
			// SFTP client'ı kapatmak okuma işlemini iptal eder
			sftpClient.Close()
			return ctx.Err()
		}
	})
}

func (t *SSHTransport) RunRemoteSecure(ctx context.Context, cmd string, becomePass string) error {
	session, err := t.client.NewSession()
	if err != nil {
		return err
	}
	// Defer session.Close() burada riskli olabilir çünkü sinyal gönderince zaten kapanabilir,
	// ama güvenli tarafta kalmak için ekliyoruz, çift kapama sorun yaratmaz.
	defer session.Close()

	modes := ssh.TerminalModes{
		ssh.ECHO:          0,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}

	if err := session.RequestPty("xterm", 80, 40, modes); err != nil {
		return fmt.Errorf("PTY isteği başarısız: %w", err)
	}

	stdin, err := session.StdinPipe()
	if err != nil {
		return err
	}
	stdout, err := session.StdoutPipe()
	if err != nil {
		return err
	}

	finalCmd := cmd
	if t.config.User != "root" && becomePass != "" {
		finalCmd = fmt.Sprintf("sudo -S -p '' %s", cmd)
	}

	if err := session.Start(finalCmd); err != nil {
		return err
	}

	if t.config.User != "root" && becomePass != "" {
		_, _ = stdin.Write([]byte(becomePass + "\n"))
	}

	go io.Copy(os.Stdout, stdout)

	// Komutun bitmesini bekleyen kanal
	done := make(chan error, 1)
	go func() {
		done <- session.Wait()
	}()

	select {
	case err := <-done:
		// Komut normal şekilde bitti (başarılı veya başarısız)
		return err
	case <-ctx.Done():
		// Kullanıcı iptal etti veya timeout
		slog.Warn("İşlem iptal edildi, uzak süreç sonlandırılıyor...")
		// Uzak sürece SIGINT/SIGKILL gönder
		_ = session.Signal(ssh.SIGKILL)
		session.Close()
		return ctx.Err()
	}
}
