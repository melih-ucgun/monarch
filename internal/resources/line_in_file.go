package resources

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"
)

type LineInFileResource struct {
	CanonicalID string `mapstructure:"-"`
	Path        string `mapstructure:"path"`
	Regexp      string `mapstructure:"regexp"` // Aranan satır (Regex)
	Line        string `mapstructure:"line"`   // Olması gereken satır
}

func (r *LineInFileResource) ID() string {
	return r.CanonicalID
}

func (r *LineInFileResource) Check() (bool, error) {
	file, err := os.Open(r.Path)
	if os.IsNotExist(err) {
		return false, nil // Dosya yoksa oluşturulmalı
	}
	if err != nil {
		return false, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	// Regex varsa onu derle
	var re *regexp.Regexp
	if r.Regexp != "" {
		re, err = regexp.Compile(r.Regexp)
		if err != nil {
			return false, fmt.Errorf("geçersiz regex: %w", err)
		}
	}

	for scanner.Scan() {
		line := scanner.Text()

		// Eğer tam satır eşleşiyorsa OK (Drift yok)
		if line == r.Line {
			return true, nil
		}

		// Eğer regex eşleşiyorsa ama içerik tam olarak r.Line değilse,
		// bu "hatalı içerik" demektir ve döngüden çıkmadan devam ederiz.
		// En sonunda false döneceği için burada ekstra bir işlem yapmaya gerek yok.
		// Sadece tam eşleşme (üstteki if) true döner.
		if re != nil && re.MatchString(line) {
			// Regex tuttu ama içerik farklı, yani drift var.
			// Döngü bittiğinde false döneceğiz.
		}
	}

	// Tam eşleşme bulunamadı -> false
	return false, nil
}

func (r *LineInFileResource) Apply() error {
	// Dosya yoksa oluştur
	if _, err := os.Stat(r.Path); os.IsNotExist(err) {
		return os.WriteFile(r.Path, []byte(r.Line+"\n"), 0644)
	}

	// Dosyayı oku
	content, err := os.ReadFile(r.Path)
	if err != nil {
		return err
	}
	lines := strings.Split(string(content), "\n")

	var re *regexp.Regexp
	if r.Regexp != "" {
		re, _ = regexp.Compile(r.Regexp)
	}

	found := false
	var newLines []string

	for _, line := range lines {
		if re != nil && re.MatchString(line) {
			// Regex eşleşti, satırı güncelle
			newLines = append(newLines, r.Line)
			found = true
		} else if line == r.Line {
			// Tam eşleşme zaten var
			newLines = append(newLines, line)
			found = true
		} else {
			newLines = append(newLines, line)
		}
	}

	if !found {
		// Eşleşme yoksa sona ekle
		// Eğer dosya boşsa veya son satır boşluksa temiz bir ekleme yap
		if len(newLines) > 0 && newLines[len(newLines)-1] == "" {
			newLines = append(newLines, r.Line)
		} else {
			newLines = append(newLines, r.Line)
		}
	}

	output := strings.Join(newLines, "\n")
	return os.WriteFile(r.Path, []byte(output), 0644)
}

func (r *LineInFileResource) Undo(ctx context.Context) error {
	// Geri alma mantığı karmaşık (eski değer bilinmiyor), pas geçiyoruz.
	return nil
}

func (r *LineInFileResource) Diff() (string, error) {
	return fmt.Sprintf("Line[%s] in %s missing or incorrect", r.Line, r.Path), nil
}
