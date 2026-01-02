package system

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"

	"github.com/melih-ucgun/veto/internal/core"
)

// Detect, mevcut sistemi analiz eder ve SystemContext'i doldurur.
func Detect(ctx *core.SystemContext) {
	// 1. Temel OS Bilgileri
	info := readOSRelease(ctx)
	ctx.OS = "linux"
	ctx.Distro = info["ID"]
	ctx.Version = info["VERSION_ID"]

	hostname, _ := ctx.Transport.Execute(ctx.Context, "hostname")
	ctx.Hostname = strings.TrimSpace(hostname)

	ctx.InitSystem = detectInitSystem(ctx)
	ctx.Kernel = detectKernel(ctx)

	// Arch tabanlı veya versiyon bilgisi olmayan sistemler için Rolling Release kontrolü
	if ctx.Version == "" {
		rollingDistros := []string{"arch", "cachyos", "manjaro", "endeavouros"}
		for _, d := range rollingDistros {
			if strings.Contains(strings.ToLower(ctx.Distro), d) {
				ctx.Version = "Rolling Release"
				break
			}
		}
	}

	if val, ok := info["ID_LIKE"]; ok {
		if strings.Contains(val, "arch") && ctx.Version == "" {
			ctx.Version = "Rolling Release"
		}
	}

	// 2. Kullanıcı Bilgileri
	if username, err := ctx.Transport.Execute(ctx.Context, "id -u -n"); err == nil {
		ctx.User = strings.TrimSpace(username)
	}
	if home, err := ctx.Transport.Execute(ctx.Context, "echo $HOME"); err == nil {
		ctx.HomeDir = strings.TrimSpace(home)
	}
	if uid, err := ctx.Transport.Execute(ctx.Context, "id -u"); err == nil {
		ctx.UID = strings.TrimSpace(uid)
	}
	if gid, err := ctx.Transport.Execute(ctx.Context, "id -g"); err == nil {
		ctx.GID = strings.TrimSpace(gid)
	}

	// 3. Donanım Tespiti
	ctx.Hardware = detectHardware(ctx)

	// 4. Çevresel Değişkenler
	ctx.Env = detectEnv(ctx)

	// 5. Dosya Sistemi
	ctx.FSInfo = detectFS(ctx, "/")
}

func readOSRelease(ctx *core.SystemContext) map[string]string {
	info := make(map[string]string)
	f, err := ctx.FS.Open("/etc/os-release")
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

func detectHardware(ctx *core.SystemContext) core.SystemHardware {
	hw := core.SystemHardware{
		CPUModel:  "Unknown CPU",
		GPUVendor: "Unknown GPU",
	}

	// CPU
	if f, err := ctx.FS.Open("/proc/cpuinfo"); err == nil {
		defer f.Close()
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.Contains(line, "model name") {
				parts := strings.Split(line, ":")
				if len(parts) > 1 {
					hw.CPUModel = strings.TrimSpace(parts[1])
				}
				break
			}
		}
	}

	if out, err := ctx.Transport.Execute(ctx.Context, "nproc"); err == nil {
		hw.CPUCore, _ = strconv.Atoi(strings.TrimSpace(out))
	}

	// RAM
	if f, err := ctx.FS.Open("/proc/meminfo"); err == nil {
		defer f.Close()
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "MemTotal:") {
				parts := strings.Fields(line)
				if len(parts) >= 2 {
					if kb, err := strconv.Atoi(parts[1]); err == nil {
						gb := float64(kb) / (1024 * 1024)
						hw.RAMTotal = fmt.Sprintf("%.1f GB", gb)
					} else {
						hw.RAMTotal = parts[1] + " " + parts[2] // Fallback
					}
				}
				break
			}
		}
	}

	// GPU (lspci check)
	if out, err := ctx.Transport.Execute(ctx.Context, "lspci"); err == nil {
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

func detectEnv(ctx *core.SystemContext) core.SystemEnv {
	shell, _ := ctx.Transport.Execute(ctx.Context, "echo $SHELL")
	lang, _ := ctx.Transport.Execute(ctx.Context, "echo $LANG")
	term, _ := ctx.Transport.Execute(ctx.Context, "echo $TERM")

	return core.SystemEnv{
		Shell:    strings.TrimSpace(shell),
		Lang:     strings.TrimSpace(lang),
		Term:     strings.TrimSpace(term),
		Timezone: detectTimezone(ctx),
	}
}

func detectTimezone(ctx *core.SystemContext) string {
	// /etc/timezone dosyasını oku veya sembolik linki kontrol et
	if content, err := ctx.FS.ReadFile("/etc/timezone"); err == nil {
		return strings.TrimSpace(string(content))
	}
	// Fallback: readlink /etc/localtime
	if link, err := ctx.FS.Readlink("/etc/localtime"); err == nil {
		// /usr/share/zoneinfo/Europe/Istanbul -> Europe/Istanbul
		parts := strings.Split(link, "zoneinfo/")
		if len(parts) > 1 {
			return parts[1]
		}
	}
	return "UTC"
}

func detectFS(ctx *core.SystemContext, path string) core.SystemFS {
	// mount komutunun çıktısı veya /proc/mounts
	// Basitçe root path için /proc/mounts taraması
	fs := core.SystemFS{RootFSType: "unknown"}

	f, err := ctx.FS.Open("/proc/mounts")
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

func detectKernel(ctx *core.SystemContext) string {
	out, err := ctx.Transport.Execute(ctx.Context, "uname -r")
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(out)
}
