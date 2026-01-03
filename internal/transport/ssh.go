package transport

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/melih-ucgun/veto/internal/core"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

type HostConfig struct {
	Name           string
	Address        string
	User           string
	Port           int
	SSHKeyPath     string
	BecomeMethod   string
	BecomePassword string
}

type SSHTransport struct {
	client     *ssh.Client
	sftpClient *sftp.Client
	sftpFS     *SFTPFS
	config     HostConfig
}

func NewSSHTransport(ctx context.Context, host HostConfig) (*SSHTransport, error) {
	var authMethods []ssh.AuthMethod

	if host.SSHKeyPath != "" {
		key, err := os.ReadFile(host.SSHKeyPath)
		if err != nil {
			return nil, fmt.Errorf("ssh anahtarı okunamadı: %w", err)
		}
		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			return nil, fmt.Errorf("ssh anahtarı parse edilemedi: %w", err)
		}
		authMethods = append(authMethods, ssh.PublicKeys(signer))
	} else {
		// Şifre ile kimlik doğrulama
		authMethods = append(authMethods, ssh.Password(host.BecomePassword))
	}

	sshConfig := &ssh.ClientConfig{
		User:            host.User,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // Not: Prodüksiyonda known_hosts doğrulaması önerilir
		Timeout:         10 * time.Second,
	}

	addr := fmt.Sprintf("%s:%d", host.Address, host.Port)
	client, err := ssh.Dial("tcp", addr, sshConfig)
	if err != nil {
		// DÜZELTME: host.HostName yerine host.Name kullanıldı
		return nil, fmt.Errorf("ssh bağlantı hatası (%s): %w", host.Name, err)
	}

	sftpClient, err := sftp.NewClient(client)
	if err != nil {
		client.Close()
		return nil, fmt.Errorf("sftp başlatılamadı: %w", err)
	}

	return &SSHTransport{
		client:     client,
		sftpClient: sftpClient,
		sftpFS:     NewSFTPFS(sftpClient),
		config:     host,
	}, nil
}

func (t *SSHTransport) Close() error {
	if t.sftpClient != nil {
		t.sftpClient.Close()
	}
	if t.client != nil {
		return t.client.Close()
	}
	return nil
}

// Execute runs a command and returns its combined output.
func (t *SSHTransport) Execute(ctx context.Context, cmd string) (string, error) {
	session, err := t.client.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()

	out, err := session.CombinedOutput(cmd)
	return string(out), err
}

// RunRemoteSecure: Komutu çalıştırır, sudo gerekirse şifreyi pipe ile verir.
// Bu fonksiyon, parolanın process listesinde veya history'de görünmesini engeller.
func (t *SSHTransport) RunRemoteSecure(ctx context.Context, cmdStr string, sudoPassword string, stdinData string) error {
	session, err := t.client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	// Eğer sudo kullanılıyorsa, komutu şifre isteyecek şekilde (promptsuz) ayarla
	finalCmd := cmdStr
	if t.config.BecomeMethod == "sudo" && sudoPassword != "" {
		// -S: Stdin'den şifre oku
		// -p '': Prompt gösterme
		finalCmd = fmt.Sprintf("sudo -S -p '' %s", cmdStr)
	}

	stdin, err := session.StdinPipe()
	if err != nil {
		return err
	}

	// Çıktıları yönlendir
	session.Stdout = os.Stdout
	session.Stderr = os.Stderr

	if err := session.Start(finalCmd); err != nil {
		return err
	}

	// Stdin yönetimi (Şifre ve veri gönderimi)
	go func() {
		defer stdin.Close()
		// 1. Önce sudo şifresini gönder (Güvenli aktarım)
		if t.config.BecomeMethod == "sudo" && sudoPassword != "" {
			fmt.Fprintln(stdin, sudoPassword)
		}
		// 2. Sonra asıl datayı gönder (varsa)
		if stdinData != "" {
			io.WriteString(stdin, stdinData)
		}
	}()

	return session.Wait()
}

func (t *SSHTransport) CopyFile(ctx context.Context, localPath, remotePath string) error {
	localFile, err := os.Open(localPath)
	if err != nil {
		return err
	}
	defer localFile.Close()

	stat, err := localFile.Stat()
	if err != nil {
		return err
	}

	remoteFile, err := t.sftpClient.Create(remotePath)
	if err != nil {
		return err
	}
	defer remoteFile.Close()

	if _, err := io.Copy(remoteFile, localFile); err != nil {
		return err
	}

	return remoteFile.Chmod(stat.Mode().Perm())
}

func (t *SSHTransport) DownloadFile(ctx context.Context, remotePath, localPath string) error {
	remoteFile, err := t.sftpClient.Open(remotePath)
	if err != nil {
		return err
	}
	defer remoteFile.Close()

	localFile, err := os.Create(localPath)
	if err != nil {
		return err
	}
	defer localFile.Close()

	_, err = io.Copy(localFile, remoteFile)
	return err
}

func (t *SSHTransport) GetFileSystem() core.FileSystem {
	return t.sftpFS
}

func (t *SSHTransport) GetOS(ctx context.Context) (string, error) {
	osName, _, err := t.GetRemoteSystemInfo(ctx)
	return osName, err
}

func (t *SSHTransport) GetRemoteSystemInfo(ctx context.Context) (string, string, error) {
	session, err := t.client.NewSession()
	if err != nil {
		return "", "", err
	}
	defer session.Close()

	// OS ve Arch bilgisini öğren
	out, err := session.Output("uname -s -m")
	if err != nil {
		return "", "", err
	}
	parts := strings.Fields(string(out))
	if len(parts) < 2 {
		return "", "", fmt.Errorf("bilinmeyen sistem çıktısı: %s", string(out))
	}

	osName := strings.ToLower(parts[0]) // Linux -> linux
	arch := strings.ToLower(parts[1])   // x86_64 -> amd64

	// Mimari isimlerini Go standartlarına çevir
	if arch == "x86_64" {
		arch = "amd64"
	} else if arch == "aarch64" {
		arch = "arm64"
	}

	return osName, arch, nil
}
