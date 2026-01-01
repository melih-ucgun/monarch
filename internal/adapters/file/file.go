package file

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/melih-ucgun/veto/internal/core"
)

func init() {
	core.RegisterResource("file", func(name string, params map[string]interface{}, ctx *core.SystemContext) (core.Resource, error) {
		return NewFileAdapter(name, params), nil
	})
}

type FileAdapter struct {
	core.BaseResource
	Path       string
	Source     string // Kopyalanacak kaynak dosya (opsiyonel)
	Content    string // Yazılacak içerik (opsiyonel)
	Method     string // copy (default), symlink
	Mode       os.FileMode
	State      string // present, absent
	BackupPath string // Yedeklenen dosyanın yolu
}

func (r *FileAdapter) GetBackupPath() string {
	return r.BackupPath
}

func NewFileAdapter(name string, params map[string]interface{}) core.Resource {
	path, _ := params["path"].(string)
	if path == "" {
		path = name // Eğer path verilmezse name'i path olarak kullan
	}

	source, _ := params["source"].(string)
	content, _ := params["content"].(string)
	state, _ := params["state"].(string)
	if state == "" {
		state = "present"
	}

	// İzinleri ayarla (varsayılan 0644)
	mode := os.FileMode(0644)
	if m, ok := params["mode"].(int); ok {
		mode = os.FileMode(m)
	}

	method, _ := params["method"].(string)
	if method == "" {
		method = "copy"
	}

	return &FileAdapter{
		BaseResource: core.BaseResource{Name: name, Type: "file"},
		Path:         path,
		Source:       source,
		Content:      content,
		Method:       method,
		Mode:         mode,
		State:        state,
	}
}

func (r *FileAdapter) Validate(ctx *core.SystemContext) error {
	if r.Path == "" {
		return fmt.Errorf("file path is required")
	}

	if r.State != "present" && r.State != "absent" {
		return fmt.Errorf("invalid state '%s', must be 'present' or 'absent'", r.State)
	}

	if r.State == "present" {
		if r.Source == "" && r.Content == "" {
			return fmt.Errorf("either 'source' or 'content' must be provided for file resource when state is 'present'")
		}
		if r.Source != "" && r.Content != "" {
			return fmt.Errorf("cannot provide both 'source' and 'content' for file resource")
		}
		if r.Source != "" {
			// Source check - usually repo-side, but let's use ctx.FS if we assume a unified view
			if _, err := ctx.FS.Stat(r.Source); os.IsNotExist(err) {
				return fmt.Errorf("source file '%s' does not exist", r.Source)
			}
		}
		if r.Method != "copy" && r.Method != "symlink" {
			return fmt.Errorf("invalid method '%s', must be 'copy' or 'symlink'", r.Method)
		}
	}

	return nil
}

func (r *FileAdapter) Check(ctx *core.SystemContext) (bool, error) {
	info, err := ctx.FS.Stat(r.Path)

	if r.State == "absent" {
		// Dosya varsa silinmeli -> değişiklik var (true)
		return !os.IsNotExist(err), nil
	}

	// Dosya yoksa oluşturulmalı -> değişiklik var
	if os.IsNotExist(err) {
		return true, nil
	}
	if err != nil {
		return false, err
	}

	// İzin kontrolü
	if info.Mode().Perm() != r.Mode {
		return true, nil
	}

	// İçerik kontrolü
	if r.Content != "" {
		existingContent, err := ctx.FS.ReadFile(r.Path)
		if err != nil {
			return false, err
		}
		if string(existingContent) != r.Content {
			return true, nil
		}
	} else if r.Source != "" {
		if r.Method == "symlink" {
			// Check if it is a symlink and points to correct source
			if info.Mode()&os.ModeSymlink == 0 {
				return true, nil // It's not a symlink
			}

			linkDest, err := ctx.FS.Readlink(r.Path)
			if err != nil {
				return false, err
			}

			// Resolve paths for safe comparison
			absSource, _ := filepath.Abs(r.Source)
			absDest, _ := filepath.Abs(linkDest)

			if absDest != absSource {
				return true, nil
			}
		} else {
			// Copy Mode
			// Source ile hedefi karşılaştır
			same, err := r.compareFiles(ctx, r.Source, r.Path)
			if err != nil {
				return false, err
			}
			if !same {
				return true, nil
			}
		}
	}
	return false, nil
}

func (r *FileAdapter) Diff(ctx *core.SystemContext) (string, error) {
	if r.State == "absent" {
		if _, err := ctx.FS.Stat(r.Path); os.IsNotExist(err) {
			return "", nil
		}
		current, _ := ctx.FS.ReadFile(r.Path)
		return core.GenerateDiff(r.Path, string(current), ""), nil
	}

	// For "present"
	var desired string
	if r.Content != "" {
		desired = r.Content
	} else if r.Source != "" {
		s, err := ctx.FS.ReadFile(r.Source)
		if err != nil {
			return "", err
		}
		desired = string(s)
	}

	current := ""
	if _, err := ctx.FS.Stat(r.Path); err == nil {
		c, _ := ctx.FS.ReadFile(r.Path)
		current = string(c)
	}

	return core.GenerateDiff(r.Path, current, desired), nil
}

func (r *FileAdapter) Apply(ctx *core.SystemContext) (core.Result, error) {
	needsAction, _ := r.Check(ctx)
	if !needsAction {
		return core.SuccessNoChange(fmt.Sprintf("File %s is up to date", r.Path)), nil
	}

	if ctx.DryRun {
		msg := fmt.Sprintf("[DryRun] Would %s file %s", r.State, r.Path)
		if r.Content != "" {
			msg += fmt.Sprintf("\nContent Preview:\n%s", r.Content)
		}
		return core.SuccessChange(msg), nil
	}

	// YEDEKLEME
	if core.GlobalBackup != nil {
		backupPath, err := core.GlobalBackup.BackupFile(r.Path)
		if err == nil {
			r.BackupPath = backupPath
		} else {
			return core.Failure(err, "Failed to backup file"), err
		}
	}

	if r.State == "absent" {
		if err := ctx.FS.Remove(r.Path); err != nil {
			return core.Failure(err, "Failed to delete file"), err
		}
		return core.SuccessChange("File deleted"), nil
	}

	// Dizin yoksa oluştur
	dir := filepath.Dir(r.Path)
	if err := ctx.FS.MkdirAll(dir, 0755); err != nil {
		return core.Failure(err, "Failed to create directory"), err
	}

	// İçerik yazma veya kopyalama
	if r.Content != "" {
		if err := ctx.FS.WriteFile(r.Path, []byte(r.Content), r.Mode); err != nil {
			return core.Failure(err, "Failed to write content"), err
		}
	} else if r.Source != "" {
		if r.Method == "symlink" {
			// Delete existing if present (since we confirmed it's wrong in Check)
			ctx.FS.Remove(r.Path)

			// Create symlink
			if err := ctx.FS.Symlink(r.Source, r.Path); err != nil {
				return core.Failure(err, "Failed to create symlink"), err
			}
		} else {
			if err := r.copyFile(ctx, r.Source, r.Path, r.Mode); err != nil {
				return core.Failure(err, "Failed to copy file"), err
			}
		}
	}

	return core.SuccessChange(fmt.Sprintf("File %s created/updated", r.Path)), nil
}

func (r *FileAdapter) Revert(ctx *core.SystemContext) error {
	if r.BackupPath != "" {
		// Yedeği geri yükle
		return r.copyFile(ctx, r.BackupPath, r.Path, r.Mode)
	}

	if r.State == "present" {
		return ctx.FS.Remove(r.Path)
	}

	return nil
}

// copyFile basitleştirilmiş kopyalama fonksiyonu (FS üzerinden)
func (r *FileAdapter) copyFile(ctx *core.SystemContext, src, dst string, mode os.FileMode) error {
	sourceFile, err := ctx.FS.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := ctx.FS.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, sourceFile); err != nil {
		return err
	}
	return ctx.FS.Chmod(dst, mode)
}

// compareFiles logic moved to FileAdapter to use FS abstraction
func (r *FileAdapter) compareFiles(ctx *core.SystemContext, src, dst string) (bool, error) {
	s, err := ctx.FS.ReadFile(src)
	if err != nil {
		return false, err
	}
	d, err := ctx.FS.ReadFile(dst)
	if err != nil {
		return false, err
	}
	return string(s) == string(d), nil
}
