package transport

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/melih-ucgun/monarch/internal/config"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
	"golang.org/x/term"
)

type SSHTransport struct {
	client *ssh.Client
	config *ssh.ClientConfig
}

func NewSSHTransport(host config.Host) (*SSHTransport, error) {
	// SSH Config hazırlığı (Mevcut kodunuzdaki gibi)
	sshConfig := &ssh.ClientConfig{
		User:            host.User,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // Geliştirme aşamasında host key kontrolü kapalı
		Timeout:         10 * time.Second,
	}

	// Eğer kullanıcı known_hosts kullanmak isterse:
	home, err := os.UserHomeDir()
	if err == nil {
		hostKeyCallback, err := knownhosts.New(filepath.Join(home, ".ssh", "known_hosts"))
		if err == nil {
			sshConfig.HostKeyCallback = hostKeyCallback
		}
	}

	// Kimlik doğrulama yöntemleri
	authMethods := []ssh.AuthMethod{}

	// 1. SSH Agent (Varsa)
	if socket := os.Getenv("SSH_AUTH_SOCK"); socket != "" {
		if conn, err := net.Dial("unix", socket); err == nil {
			conn.Close() // EKLENEN SATIR: Bağlantıyı kapatarak değişkeni kullanmış oluyoruz.
			// agent.NewClient(conn) kullanabilirdik ama harici paket import etmemek için
			// şimdilik basit tutuyoruz veya private key'e öncelik veriyoruz.
		}
	}

	// 2. Private Key (~/.ssh/id_rsa vb.)
	keyPath := filepath.Join(home, ".ssh", "id_rsa") // Varsayılan key
	key, err := os.ReadFile(keyPath)
	if err == nil {
		signer, err := ssh.ParsePrivateKey(key)
		if err == nil {
			authMethods = append(authMethods, ssh.PublicKeys(signer))
		}
	} else if host.Password != "" {
		// 3. Şifre (Config'den geliyorsa)
		authMethods = append(authMethods, ssh.Password(host.Password))
	} else {
		// 4. İnteraktif Şifre (Terminalden sor)
		fmt.Printf("%s@%s için SSH şifresi: ", host.User, host.Address)
		pass, err := term.ReadPassword(int(os.Stdin.Fd()))
		if err == nil {
			fmt.Println() // Yeni satır
			authMethods = append(authMethods, ssh.Password(string(pass)))
		}
	}

	sshConfig.Auth = authMethods

	addr := fmt.Sprintf("%s:%d", host.Address, host.Port)
	client, err := ssh.Dial("tcp", addr, sshConfig)
	if err != nil {
		return nil, fmt.Errorf("SSH bağlantı hatası: %w", err)
	}

	return &SSHTransport{client: client, config: sshConfig}, nil
}

func (t *SSHTransport) Close() error {
	return t.client.Close()
}

// GetRemoteSystemInfo, uzak sunucunun OS ve Arch bilgisini döner (örn: linux, amd64)
func (t *SSHTransport) GetRemoteSystemInfo() (string, string, error) {
	session, err := t.client.NewSession()
	if err != nil {
		return "", "", err
	}
	defer session.Close()

	var b bytes.Buffer
	session.Stdout = &b
	// Tek komutla OS ve Arch bilgisini alıyoruz
	if err := session.Run("uname -s && uname -m"); err != nil {
		return "", "", fmt.Errorf("sistem bilgisi alınamadı: %w", err)
	}

	parts := strings.Split(strings.TrimSpace(b.String()), "\n")
	if len(parts) < 2 {
		return "", "", fmt.Errorf("beklenmeyen sistem çıktısı: %v", parts)
	}

	osName := strings.ToLower(parts[0]) // Linux -> linux
	arch := strings.ToLower(parts[1])   // x86_64 -> amd64, aarch64 -> arm64

	// Go uyumlu isimlendirmelere çevir
	if arch == "x86_64" {
		arch = "amd64"
	} else if arch == "aarch64" {
		arch = "arm64"
	}

	return osName, arch, nil
}

// CopyFile, yerel bir dosyayı uzak sunucuya kopyalar (SCP benzeri, ama basit cat yöntemiyle)
// Büyük dosyalar için SFTP kullanılması daha iyidir ama binary kopyalamak için bu yöntem şimdilik yeterli.
func (t *SSHTransport) CopyFile(localPath, remotePath string) error {
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

	// SCP protokolünü simüle ediyoruz (basitçe hedef klasöre yazıyoruz)
	cmd := fmt.Sprintf("scp -t %s", filepath.Dir(remotePath))
	if err := session.Run(cmd); err != nil {
		return fmt.Errorf("dosya kopyalama hatası: %w", err)
	}
	return nil
}

// DownloadFile, uzak sunucudaki bir dosyayı yerel bir yola güvenli bir şekilde indirir (SFTP).
func (t *SSHTransport) DownloadFile(remotePath, localPath string) error {
	// SFTP istemcisini başlat
	sftpClient, err := sftp.NewClient(t.client)
	if err != nil {
		return fmt.Errorf("SFTP başlatılamadı: %w", err)
	}
	defer sftpClient.Close()

	// Uzak dosyayı aç
	remoteFile, err := sftpClient.Open(remotePath)
	if err != nil {
		return fmt.Errorf("uzak dosya açılamadı (%s): %w", remotePath, err)
	}
	defer remoteFile.Close()

	// Yerel dosyayı oluştur
	localFile, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("yerel dosya oluşturulamadı (%s): %w", localPath, err)
	}
	defer localFile.Close()

	// Veriyi kopyala (Stream)
	_, err = io.Copy(localFile, remoteFile)
	if err != nil {
		return fmt.Errorf("dosya indirilirken hata oluştu: %w", err)
	}

	// Dosyanın diske yazıldığından emin ol
	return localFile.Sync()
}

// CaptureRemoteOutput, uzak sunucuda bir komut çalıştırıp çıktısını döner (basit string veriler için)
func (t *SSHTransport) CaptureRemoteOutput(cmd string) (string, error) {
	session, err := t.client.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()

	output, err := session.CombinedOutput(cmd)
	if err != nil {
		return "", err
	}
	return string(output), nil
}

// RunRemoteSecure, sudo gerektiren komutları güvenli bir şekilde çalıştırır
func (t *SSHTransport) RunRemoteSecure(cmd string, becomePass string) error {
	session, err := t.client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	// TTY (Pseudo-terminal) iste, böylece sudo şifresi sorulabilir
	modes := ssh.TerminalModes{
		ssh.ECHO:          0,     // Şifreyi ekrana yazma
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
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

	// Komutu sudo ile başlat (eğer root değilsek)
	finalCmd := cmd
	if t.config.User != "root" && becomePass != "" {
		finalCmd = fmt.Sprintf("sudo -S -p '' %s", cmd)
	}

	if err := session.Start(finalCmd); err != nil {
		return err
	}

	// Sudo şifresi gerekirse gönder
	if t.config.User != "root" && becomePass != "" {
		// Basit bir bekleme veya çıktı kontrolü yapılabilir, burada direkt gönderiyoruz
		// (Daha güvenli hali: prompt beklemektir ama karmaşıklığı artırır)
		_, _ = stdin.Write([]byte(becomePass + "\n"))
	}

	// Çıktıyı canlı olarak ekrana bas (kullanıcı görsün)
	go io.Copy(os.Stdout, stdout)

	return session.Wait()
}
