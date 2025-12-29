package resources

import (
	"context"
	"fmt"
	"os/exec"
)

type ExecResource struct {
	CanonicalID string
	Command     string
	Unless      string
	OnlyIf      string
	Creates     string
	RunAsUser   string
}

func (e *ExecResource) ID() string {
	if e.CanonicalID != "" {
		return e.CanonicalID
	}
	return fmt.Sprintf("exec:%s", e.Command)
}

func (e *ExecResource) Check() (bool, error) {
	// Unless ve OnlyIf kontrolleri de RunAsUser bağlamında çalışmalı mı?
	// Genellikle evet. Ancak basitlik adına şimdilik root olarak kontrol ediyoruz.
	// İhtiyaç olursa burası da sudo -u ile sarmalanabilir.

	if e.Unless != "" {
		if err := exec.Command("sh", "-c", e.Unless).Run(); err == nil {
			return true, nil // Unless başarılıysa (exit 0), işlem yapma
		}
	}
	if e.OnlyIf != "" {
		if err := exec.Command("sh", "-c", e.OnlyIf).Run(); err != nil {
			return true, nil // OnlyIf başarısızsa, işlem yapma
		}
	}
	// Exec kaynağı "durum" tutmaz, her çalıştırıldığında (unless yoksa) false döner.
	return false, nil
}

func (e *ExecResource) Apply() error {
	var cmd *exec.Cmd

	if e.RunAsUser != "" {
		fmt.Printf("🚀 Çalıştırılıyor (%s): %s\n", e.RunAsUser, e.Command)
		// Kullanıcı adına geçiş yaparak çalıştır
		cmd = exec.Command("sudo", "-u", e.RunAsUser, "sh", "-c", e.Command)
	} else {
		fmt.Printf("🚀 Çalıştırılıyor: %s\n", e.Command)
		cmd = exec.Command("sh", "-c", e.Command)
	}

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("exec hatası: %s\nÇıktı: %s", err, string(out))
	}
	return nil
}

func (e *ExecResource) Diff() (string, error) {
	userMsg := ""
	if e.RunAsUser != "" {
		userMsg = fmt.Sprintf(" (User: %s)", e.RunAsUser)
	}
	return fmt.Sprintf("! exec: %s%s", e.Command, userMsg), nil
}

func (e *ExecResource) Undo(ctx context.Context) error {
	// Exec için genel bir undo yoktur.
	return nil
}
