package snapshot

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/pterm/pterm"
)

// Timeshift manages integration with the timeshift tool
type Timeshift struct{}

func NewTimeshift() *Timeshift {
	return &Timeshift{}
}

func (t *Timeshift) Name() string {
	return "Timeshift"
}

func (t *Timeshift) IsAvailable() bool {
	path, err := exec.LookPath("timeshift")
	// Ayrıca root yetkisi gerekebilir ama şimdilik binary kontrolü yeterli
	return err == nil && path != ""
}

func (t *Timeshift) CreateSnapshot(description string) error {
	// timeshift --create --comments "..." --tags D
	// D: Daily, O: Ondemand (genelde O kullanılır manuel için ama timeshift cli bazen tag ister)
	// --script flag'i non-interactive mod için önemli
	pterm.Info.Println("Creating Timeshift snapshot (this might take a while)...")

	// Timeshift needs to run as root usually. Veto assumes it has permissions or sudo.
	cmd := exec.Command("timeshift", "--create", "--comments", description, "--tags", "O", "--script")

	// Output'u yakalamak debug için iyi olabilir
	if output, err := cmd.CombinedOutput(); err != nil {
		// Hata durumunda output'un son kısmını göster
		outStr := string(output)
		lines := strings.Split(outStr, "\n")
		lastLines := ""
		if len(lines) > 3 {
			lastLines = strings.Join(lines[len(lines)-3:], "\n")
		} else {
			lastLines = outStr
		}
		return fmt.Errorf("timeshift failed: %v\nOutput: %s", err, lastLines)
	}

	pterm.Success.Println("Timeshift snapshot created")
	return nil
}

// Timeshift doesn't have a native "pre-post" pairing exposed easily in CLI like Snapper.
// So we treat Pre as just taking a snapshot, and Post as no-op (or another snapshot if really desired, but usually overkill).

func (t *Timeshift) CreatePreSnapshot(description string) (string, error) {
	err := t.CreateSnapshot(description)
	if err != nil {
		return "", err
	}
	return "done", nil // Return a dummy ID to signal success
}

func (t *Timeshift) CreatePostSnapshot(id string, description string) error {
	// Timeshift operations are usually heavy (especially in rsync mode).
	// Taking TWO snapshots (pre and post) might be too slow.
	// For Timeshift, "Pre" snapshot is the most critical one for rollback.
	// We will skip Post snapshot to save time, as we already have the state BEFORE changes.
	return nil
}
