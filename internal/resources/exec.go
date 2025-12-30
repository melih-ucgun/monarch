package resources

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

type ExecResource struct {
	CanonicalID string `mapstructure:"-"`
	Command     string `mapstructure:"command"`
	Unless      string `mapstructure:"unless"` // Eğer bu komut 0 dönerse, asıl komutu çalıştırma
	RunAsUser   string `mapstructure:"run_as"` // TODO: User switch logic (su/sudo) eklenebilir
}

func (r *ExecResource) ID() string {
	return r.CanonicalID
}

func (r *ExecResource) Check() (bool, error) {
	// Eğer 'Unless' komutu tanımlıysa, önce onu çalıştır.
	// Unless komutu BAŞARILI (exit 0) dönerse, asıl komutun çalışmasına gerek yok demektir -> return true
	if r.Unless != "" {
		cmd := exec.Command("sh", "-c", r.Unless)
		if err := cmd.Run(); err == nil {
			return true, nil
		}
	}

	// Eğer Unless yoksa veya başarısız olduysa, Command her zaman çalışmalıdır.
	// Ancak idempotent olmayan komutlar için Check() her zaman false döner (her run'da çalışır).
	// Eğer bir durum kontrolü isteniyorsa 'Unless' kullanılmalıdır.
	return false, nil
}

func (r *ExecResource) Apply() error {
	trimmedCmd := strings.TrimSpace(r.Command)
	if trimmedCmd == "" {
		return nil
	}

	// Shell üzerinden çalıştır ki pipe/redirect gibi özellikler kullanılabilsin
	cmd := exec.Command("sh", "-c", r.Command)

	// Çıktıyı yakala ki hata durumunda loglayabilelim
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("komut hatası: %s, output: \n%s", err, string(output))
	}
	return nil
}

func (r *ExecResource) Undo(ctx context.Context) error {
	// Shell komutlarının otomatik geri alınması imkansızdır.
	// İleride 'undo_command' parametresi eklenebilir.
	return nil
}

func (r *ExecResource) Diff() (string, error) {
	return fmt.Sprintf("Command needs to run: %s", r.Command), nil
}
