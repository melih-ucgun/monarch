package system

import (
	"os"
	"os/exec"
	"strings"
)

func detectInitSystem() string {
	// 1. Check PID 1 (most reliable)
	// /proc/1/comm usually contains "systemd" or "init"
	if comm, err := os.ReadFile("/proc/1/comm"); err == nil {
		s := strings.TrimSpace(string(comm))
		if s == "systemd" {
			return "systemd"
		}
	}

	// 2. Check /run/systemd/system (Standard way to check if booted with systemd)
	if _, err := os.Stat("/run/systemd/system"); err == nil {
		return "systemd"
	}

	// 3. OpenRC checks
	if _, err := os.Stat("/run/openrc"); err == nil {
		return "openrc"
	}
	if _, err := exec.LookPath("rc-service"); err == nil {
		return "openrc"
	}

	// 4. SysVinit checks (if /etc/init.d exists and no systemd/openrc detected)
	if _, err := os.Stat("/etc/init.d"); err == nil {
		return "sysvinit"
	}

	return "unknown"
}
