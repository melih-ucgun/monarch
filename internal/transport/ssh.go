package transport

import (
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/melih-ucgun/monarch/internal/config"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"golang.org/x/crypto/ssh/knownhosts"
)

type SSHTransport struct {
	Host   config.Host
	Client *ssh.Client
}

func NewSSHTransport(h config.Host) (*SSHTransport, error) {
	var authMethods []ssh.AuthMethod

	if socket := os.Getenv("SSH_AUTH_SOCK"); socket != "" {
		if conn, err := net.Dial("unix", socket); err == nil {
			authMethods = append(authMethods, ssh.PublicKeysCallback(agent.NewClient(conn).Signers))
		}
	}

	if h.KeyPath != "" {
		expanded := h.KeyPath
		if len(h.KeyPath) > 0 && h.KeyPath[0] == '~' {
			home, _ := os.UserHomeDir()
			expanded = filepath.Join(home, h.KeyPath[1:])
		}
		if key, err := os.ReadFile(expanded); err == nil {
			var signer ssh.Signer
			if h.Passphrase != "" {
				signer, _ = ssh.ParsePrivateKeyWithPassphrase(key, []byte(h.Passphrase))
			} else {
				signer, _ = ssh.ParsePrivateKey(key)
			}
			if signer != nil {
				authMethods = append(authMethods, ssh.PublicKeys(signer))
			}
		}
	}

	if h.Password != "" {
		authMethods = append(authMethods, ssh.Password(h.Password))
	}

	hostKeyCallback, _ := getHostKeyCallback()
	clientConfig := &ssh.ClientConfig{
		User: h.User, Auth: authMethods, HostKeyCallback: hostKeyCallback, Timeout: 15 * time.Second,
	}

	addr := h.Address
	if !strings.Contains(addr, ":") {
		addr += ":22"
	}

	client, err := ssh.Dial("tcp", addr, clientConfig)
	if err != nil {
		return nil, err
	}

	return &SSHTransport{Host: h, Client: client}, nil
}

// RunRemoteSecure, sudo şifresi gerektiren durumlarda şifreyi otomatik gönderir.
func (s *SSHTransport) RunRemoteSecure(command string, sudoPassword string) error {
	session, err := s.Client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	modes := ssh.TerminalModes{ssh.ECHO: 0}
	if err := session.RequestPty("xterm", 80, 40, modes); err != nil {
		return err
	}

	in, _ := session.StdinPipe()
	out, _ := session.StdoutPipe()
	session.Stderr = os.Stderr

	if err := session.Start(command); err != nil {
		return err
	}

	go func() {
		buf := make([]byte, 1024)
		for {
			n, err := out.Read(buf)
			if err != nil {
				return
			}
			output := string(buf[:n])
			os.Stdout.Write(buf[:n])
			if strings.Contains(strings.ToLower(output), "password") && sudoPassword != "" {
				fmt.Fprintln(in, sudoPassword)
			}
		}
	}()
	return session.Wait()
}

func (s *SSHTransport) CopyFile(localPath, remotePath string) error {
	file, _ := os.Open(localPath)
	defer file.Close()
	stat, _ := file.Stat()
	session, _ := s.Client.NewSession()
	defer session.Close()

	go func() {
		w, _ := session.StdinPipe()
		defer w.Close()
		fmt.Fprintf(w, "C%04o %d %s\n", 0755, stat.Size(), filepath.Base(localPath))
		io.Copy(w, file)
		fmt.Fprint(w, "\x00")
	}()
	return session.Run(fmt.Sprintf("/usr/bin/scp -t %s", filepath.Dir(remotePath)))
}

func getHostKeyCallback() (ssh.HostKeyCallback, error) {
	home, _ := os.UserHomeDir()
	path := filepath.Join(home, ".ssh", "known_hosts")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return ssh.InsecureIgnoreHostKey(), nil
	}
	return knownhosts.New(path)
}
