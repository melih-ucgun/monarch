package transport

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/melih-ucgun/monarch/internal/config"
	"golang.org/x/crypto/ssh"
)

type SSHTransport struct {
	Host   config.Host
	Client *ssh.Client
}

func NewSSHTransport(h config.Host) (*SSHTransport, error) {
	config := &ssh.ClientConfig{
		User: h.User,
		Auth: []ssh.AuthMethod{
			ssh.Password(h.Password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // MVP için basit tutulmuştur
	}

	client, err := ssh.Dial("tcp", h.Address, config)
	if err != nil {
		return nil, fmt.Errorf("SSH bağlantısı başarısız: %w", err)
	}

	return &SSHTransport{Host: h, Client: client}, nil
}

// RunRemote, uzak sunucuda bir komut çalıştırır ve çıktısını yerel stdout'a bağlar.
func (s *SSHTransport) RunRemote(command string) error {
	session, err := s.Client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	session.Stdout = os.Stdout
	session.Stderr = os.Stderr

	return session.Run(command)
}

// CopyFile, yerel bir dosyayı uzak sunucudaki hedef yola kopyalar.
func (s *SSHTransport) CopyFile(localPath, remotePath string) error {
	file, err := os.Open(localPath)
	if err != nil {
		return err
	}
	defer file.Close()

	stat, _ := file.Stat()

	session, err := s.Client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	go func() {
		w, _ := session.StdinPipe()
		defer w.Close()
		fmt.Fprintf(w, "C%04o %d %s\n", 0755, stat.Size(), filepath.Base(localPath))
		io.Copy(w, file)
		fmt.Fprint(w, "\x00")
	}()

	return session.Run("/usr/bin/scp -t " + filepath.Dir(remotePath))
}
