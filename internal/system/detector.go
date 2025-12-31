package system

import (
	"bufio"
	"os"
	"os/exec"
	"os/user"
	"runtime"
	"strings"

	"github.com/melih-ucgun/monarch/internal/core"
)

// Detect, mevcut sistemi analiz eder ve bir SystemContext doldurur.
func Detect(dryRun bool) *core.SystemContext {
	ctx := core.NewSystemContext(dryRun)

	// 1. Temel OS Bilgileri
	info := readOSRelease()
	ctx.OS = "linux"
	ctx.Distro = info["ID"]
	ctx.Version = info["VERSION_ID"]
	ctx.Hostname, _ = os.Hostname()

	if val, ok := info["ID_LIKE"]; ok {
		if strings.Contains(val, "arch") {
			// Arch tabanlı sistemleri gerektiğinde işaretleyebiliriz
		}
	}

	// 2. Kullanıcı Bilgileri
	currentUser, err := user.Current()
	if err == nil {
		ctx.User = currentUser.Username
		ctx.HomeDir = currentUser.HomeDir
		ctx.UID = currentUser.Uid
		ctx.GID = currentUser.Gid
	}

	// 3. Donanım Tespiti
	ctx.Hardware = detectHardware()

	// 4. Çevresel Değişkenler
	ctx.Env = detectEnv()

	// 5. Dosya Sistemi
	ctx.FS = detectFS("/")

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
			val := strings.Trim(parts[1], "\"")
			info[key] = val
		}
	}
	return info
}

func detectHardware() core.SystemHardware {
	hw := core.SystemHardware{
		CPUModel:  "Unknown CPU",
		GPUVendor: "Unknown GPU",
	}

	// CPU
	if f, err := os.Open("/proc/cpuinfo"); err == nil {
		defer f.Close()
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.Contains(line, "model name") {
				parts := strings.Split(line, ":")
				if len(parts) > 1 {
					hw.CPUModel = strings.TrimSpace(parts[1])
				}
				// Tekirdek sayısını runtime.NumCPU() ile almak daha doğru olabilir ama
				// /proc/cpuinfo'dan okuyorsak count yapabiliriz.
				// Şimdilik sadece model adını ilk bulduğumuzda alalım.
				break
			}
		}
	}
	hw.CPUCore = runtime.NumCPU()

	// RAM
	if f, err := os.Open("/proc/meminfo"); err == nil {
		defer f.Close()
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "MemTotal:") {
				parts := strings.Fields(line)
				if len(parts) >= 2 {
					// KB cinsinden gelir, GB'a çevirelim (kabaca)
					// parts[1] is e.g. "32655120"
					// Basit tutmak için string olarak raw alıyorum, ileride parse edilebilir.
					hw.RAMTotal = parts[1] + " " + parts[2]
				}
				break
			}
		}
	}

	// GPU (lspci check)
	// lspci -mm (machine readable) çıktısını kullanmak daha güvenli olabilir
	if out, err := exec.Command("lspci").Output(); err == nil {
		output := string(out)
		lowerOut := strings.ToLower(output)
		if strings.Contains(lowerOut, "nvidia") {
			hw.GPUVendor = "NVIDIA"
			hw.GPUModel = extractGPUModel(output, "NVIDIA")
		} else if strings.Contains(lowerOut, "amd") || strings.Contains(lowerOut, "radeon") {
			hw.GPUVendor = "AMD"
			hw.GPUModel = extractGPUModel(output, "AMD")
		} else if strings.Contains(lowerOut, "intel") {
			hw.GPUVendor = "Intel"
			hw.GPUModel = extractGPUModel(output, "Intel")
		}
	}

	return hw
}

func extractGPUModel(lspciOut, vendor string) string {
	// Basit bir parser; vendor ismini içeren satırı bulup VGA uyumlu olanı döndürür.
	lines := strings.Split(lspciOut, "\n")
	for _, line := range lines {
		if strings.Contains(line, "VGA compatible controller") && strings.Contains(strings.ToLower(line), strings.ToLower(vendor)) {
			// Ör: "01:00.0 VGA compatible controller: NVIDIA Corporation GA104 [GeForce RTX 3070] (rev a1)"
			// Sadece parantez içi veya ":" sonrası kısmı alabiliriz.
			parts := strings.Split(line, ":")
			if len(parts) >= 3 {
				return strLimit(strings.TrimSpace(parts[2]), 40)
			}
		}
	}
	return "Unknown Model"
}

func detectEnv() core.SystemEnv {
	return core.SystemEnv{
		Shell:    os.Getenv("SHELL"),
		Lang:     os.Getenv("LANG"),
		Term:     os.Getenv("TERM"),
		Timezone: detectTimezone(),
	}
}

func detectTimezone() string {
	// /etc/timezone dosyasını oku veya sembolik linki kontrol et
	if content, err := os.ReadFile("/etc/timezone"); err == nil {
		return strings.TrimSpace(string(content))
	}
	// Fallback: readlink /etc/localtime
	if link, err := os.Readlink("/etc/localtime"); err == nil {
		// /usr/share/zoneinfo/Europe/Istanbul -> Europe/Istanbul
		parts := strings.Split(link, "zoneinfo/")
		if len(parts) > 1 {
			return parts[1]
		}
	}
	return "UTC"
}

func detectFS(path string) core.SystemFS {
	// mount komutunun çıktısı veya /proc/mounts
	// Basitçe root path için /proc/mounts taraması
	fs := core.SystemFS{RootFSType: "unknown"}

	f, err := os.Open("/proc/mounts")
	if err != nil {
		return fs
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) >= 3 {
			mountPoint := fields[1]
			fsType := fields[2]
			if mountPoint == path {
				fs.RootFSType = fsType
				break
			}
		}
	}
	return fs
}

func strLimit(s string, limit int) string {
	if len(s) > limit {
		return s[:limit] + "..."
	}
	return s
}
