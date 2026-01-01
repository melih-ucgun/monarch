# 👑 Veto

> **The Sovereign System Orchestrator**
> _"Don't just manage your OS. Rule it."_

**Veto** transforms Linux system management from a chaotic, irreversible process into a **modular, self-healing Lego experience**. It treats your system not as a monolithic entity, but as a collection of attachable and detachable **Profiles**.

Whether you are running a minimal setup with Hyprland on Arch or a server fleet on Fedora, Veto gives you the power of immutable systems with the flexibility of a rolling release.

---

## ⚡ The Problem

Modern Linux setups are fragmented. You install a package with `pacman`, manage configs with `stow`, enable services with `systemctl`, and fix permissions manually. **One change makes the system "dirty," and undoing it becomes nearly impossible.**

Veto solves this by providing a unified, context-aware interface for all system resources.

## 🆚 Veto vs. The Ecosystem

| **Feature** | **👑 Veto** | **❄️ NixOS** | **🐍 Ansible** | **📂 Chezmoi / Stow** |
| :--- | :--- | :--- | :--- | :--- |
| **Primary Goal** | Modular Desktop Orchestration | Reproducible OS | Server Configuration | Dotfile Management |
| **"Undo" Button** | ✅ **Native** (Atomic Revert + BTRFS) | ⚠️ Rollback (Whole OS) | ❌ Manual Playbooks | ❌ None |
| **OS Requirement** | Any (Arch, Fedora, Debian, etc.) | Must use NixOS | Any | Any |
| **Scope** | Pkgs + Configs + Svcs + Secrets | Everything | Everything | Config Files Only |
| **Drift Detection** | ✅ **Auto-Repair** | ⚠️ Read-only Store | ❌ Overwrite on run | ❌ Overwrite on run |
| **Learning Curve** | 🟢 **Low** (Simple YAML) | 🔴 Very High | 🟡 Medium | 🟢 Low |

---

## 🚀 Current Features

### 1. **Context-Aware Intelligence**
Veto analyzes your hardware (CPU, GPU), distribution, and environment variables. It can intelligently select the right drivers and configurations based on your system's "personality."

### 2. **System Discovery & Import**
Moving from a "dirty" system to Veto is easy. The `veto import` command scans your installed packages, enabled services, and common configuration files, generating a baseline YAML for you.

### 3. **Secret Management**
Veto has built-in encryption for sensitive values. Generate a master key and use `veto secret encrypt` to store passwords or tokens securely in your configuration files.

### 4. **Atomic Snapshots & Rollback**
Before applying changes, Veto can automatically create BTRFS snapshots (via Snapper or Timeshift). If something goes wrong, you can `veto rollback` to return to the exact state before the operation.

### 5. **Live Watch Mode**
Run `veto watch` to monitor your configuration files. As soon as you save a change, Veto automatically applies it to your system—perfect for iterative styling of your desktop environment.

### 6. **Comprehensive Resource Support**
Veto is powered by a modular adapter system supporting:
- **Package Managers:** 13+ managers including `pacman`, `yay`, `apt`, `dnf`, `brew`, `flatpak`, `snap`, etc.
- **Service Managers:** `systemd`, `openrc`, `sysvinit`.
- **Identity Management:** `user` and `group` creation and modification.
- **Filesystem:** Templates, symlinks, line-in-file edits, and archive extraction.
- **Automation:** Git repository management and custom shell execution.

---

## 🏗️ Supported Adapters

Veto is designed to be distro-agnostic. It currently supports:

| Category | Supported Technologies / Types |
| :--- | :--- |
| **Pkg Managers** | `pacman`, `yay`, `paru`, `apt`, `dnf`, `yum`, `zypper`, `brew`, `apk`, `flatpak`, `snap` |
| **Services** | `systemd`, `openrc`, `sysvinit` |
| **Files** | `file` (create/delete), `symlink`, `template` (Go templates), `line_in_file`, `archive`, `download` |
| **System** | `user`, `group`, `git`, `shell` (exec) |
| **Snapshots** | `snapper`, `timeshift` (BTRFS) |

---

## 📦 Installation & Quick Start

Veto is a single binary written in Go. No external dependencies are required for the core engine.

### 1. Initialize System Context
This scans your hardware and OS to configure the local registry.
```bash
veto init
```

### 2. Import Current State (The "Sovereign" Path)
Discovers your current setup and creates a baseline configuration.
```bash
veto import my-system.yaml
```

### 3. Apply Configuration
Ensure your system state matches your YAML definition.
```bash
# Preview changes first
veto apply my-system.yaml --dry-run

# Apply for real (automatically creates a snapshot)
veto apply my-system.yaml
```

### 4. Watch for Changes
Iterate on your setup in real-time.
```bash
veto watch my-system.yaml
```

---

## 🛠️ Configuration Example

```yaml
# my-system.yaml
resources:
  - name: base-packages
    type: pkg
    params:
      names: [git, curl, vim]
      state: present

  - name: enable-docker
    type: service
    name: docker
    state: running
    params:
      enabled: true
    depends_on: [pkg:docker]

  - name: dotfiles
    type: git
    params:
      url: "https://github.com/user/dotfiles.git"
      path: "/home/user/.dotfiles"

  - name: zshrc
    type: symlink
    params:
      source: "/home/user/.dotfiles/.zshrc"
      target: "/home/user/.zshrc"
```

---

## 🗺️ Roadmap

- [x] **Core Engine:** Resource adapters, State management, Undo logic.
- [x] **System Awareness:** Auto-detection of Hardware/OS.
- [x] **Secrets System:** Built-in encryption for sensitive data.
- [x] **Import/Export:** System discovery for easy migration.
- [ ] **Veto Hub:** A decentralized registry for sharing rulesets.
- [ ] **Veto Studio:** A GUI Dashboard for visual system orchestration.

---

## 🤝 Contributing

Veto is community-driven. We welcome contributions to adapters, core logic, or the upcoming Hub.

1. Fork the Project.
2. Create your Feature Branch (`git checkout -b feature/AmazingFeature`).
3. Commit your Changes (`git commit -m 'Add some AmazingFeature'`).
4. Push to the Branch (`git push origin feature/AmazingFeature`).
5. Open a Pull Request.

---

## 📜 License

Distributed under the Apache 2.0 License. See `LICENSE` for more information.

**Veto** © 2025 Developed by **Melih Uçgun**
_"We don't just configure systems—we build sovereign infrastructure."_
