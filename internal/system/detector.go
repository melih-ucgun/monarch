package system

import (
	"bufio"
	"os"
	"strings"

	"github.com/melih-ucgun/monarch/internal/core"
)

// Detect, mevcut sistemi analiz eder ve bir SystemContext doldurur.
func Detect(dryRun bool) *core.SystemContext {
	ctx := core.NewSystemContext(dryRun)

	// Basit Linux tespiti (/etc/os-release okuyarak)
	// İleride burası macOS (Darwin) ve Windows desteği ile genişletilebilir.
	info := readOSRelease()

	ctx.OS = "linux" // Şimdilik varsayılan
	ctx.Distro = info["ID"]
	ctx.Version = info["VERSION_ID"]

	// CachyOS, Arch tabanlıdır. ID bazen "cachyos", ID_LIKE "arch" olabilir.
	// Paket yöneticisi kararı için bu bilgi önemli.
	if val, ok := info["ID_LIKE"]; ok {
		// ID_LIKE birden fazla değer içerebilir "arch cachyos" gibi
		if strings.Contains(val, "arch") {
			// Arch türevi olarak işaretleyebiliriz veya Distro bilgisini olduğu gibi bırakırız
		}
	}

	return ctx
}

func readOSRelease() map[string]string {
	info := make(map[string]string)
	f, err := os.Open("/etc/os-release")
	if err != nil {
		return info
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if parts := strings.SplitN(line, "=", 2); len(parts) == 2 {
			key := parts[0]
			// Tırnak işaretlerini temizle
			val := strings.Trim(parts[1], "\"")
			info[key] = val
		}
	}
	return info
}
