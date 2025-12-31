# 👑 Monarch

> **The Sovereign System Orchestrator** > _"Don't just manage your OS. Rule it."_

**Monarch** transforms Linux system management from a chaotic, irreversible process into a **modular, self-healing Lego experience**. It treats your system not as a monolithic entity, but as a collection of attachable and detachable **Profiles**.

Whether you are running a minimal **CachyOS** setup with Hyprland or a full-blown server fleet, Monarch gives you the power of immutable systems with the flexibility of a rolling release.

## ⚡ The Problem

Modern Linux setups are fragmented. You install a package with `pacman`, manage configs with `stow`, enable services with `systemctl`, and fix permissions manually. **One change makes the system "dirty," and undoing it becomes nearly impossible.**

## 🆚 Monarch vs. The Ecosystem

Why build a new tool? Because existing solutions force you to choose between **flexibility** and **stability**.

|   |   |   |   |   |
|---|---|---|---|---|
|**Feature**|**👑 Monarch**|**❄️ NixOS**|**🐍 Ansible**|**📂 Chezmoi / Stow**|
|**Primary Goal**|Modular Desktop Orchestration|Reproducible OS|Server Configuration|Dotfile Management|
|**"Undo" Button**|✅ **Native** (Atomic Revert)|⚠️ Rollback (Whole OS)|❌ Manual Playbooks|❌ None|
|**OS Requirement**|Any (Arch/Cachy, Fedora, etc.)|Must use NixOS|Any|Any|
|**Scope**|Packages + Configs + Services|Everything|Everything|Config Files Only|
|**Drift Detection**|✅ **Auto-Repair**|⚠️ Read-only Store|❌ Overwrite on run|❌ Overwrite on run|
|**Learning Curve**|🟢 **Low** (Simple YAML)|🔴 Very High|🟡 Medium|🟢 Low|

> **The Verdict:**
> 
> - **Use NixOS** if you are willing to replace your entire OS and learn a new language.
>     
> - **Use Ansible** if you are managing 1000 servers and don't care about "undoing" changes on a laptop.
>     
> - **Use Monarch** if you want the stability of NixOS on your favorite distro (like CachyOS) with the ease of use of a Lego set.
>     

## 🚀 Key Features

### 1. Profile-First Architecture

Define your entire system personality in a single, human-readable YAML file. Switch between "Work Mode", "Gaming Mode", or "Minimal Mode" in seconds.

```
# ~/.config/monarch/profiles/workstation.yaml
name: "Dev Workstation"
description: "High performance setup for Go & Rust dev"
rulesets:
  - official:base-devel
  - official:hyprland:latest
  - community:vscode:insiders
  - community:docker:rootless
self_healing:
  enabled: true
```

### 2. Context-Aware Intelligence

Monarch doesn't blindly run scripts. It analyzes the host first.

> _Example:_ If you apply a gaming profile on a **Ryzen 7 7730u** laptop, Monarch intelligently selects `mesa` and `vulkan-radeon` instead of forcing NVIDIA drivers.

### 3. True Undo Capability

Most package managers remove the binary but leave the chaos. Monarch tracks every file created, every permission changed, and every service enabled.

```
monarch profile disable gaming
# Result: System returns to the exact state before the profile was applied.
```

## 📦 Installation & Quick Start

Monarch is a single binary written in Go. No dependencies required.

### 1. Install

```
curl -L [https://monarch.sh/install](https://monarch.sh/install) | sudo bash
```

### 2. Initialize (System Detection)

This step scans your hardware (CPU, GPU) and OS (e.g., CachyOS, Arch, Fedora) to configure the local registry.

```
monarch init
```

### 3. Search & Apply

Find what you need in the Hub and apply it.

```
# Find a window manager setup
monarch hub search hyprland

# Apply a pre-made profile (Dry run first!)
monarch profile apply minimal-desktop --dry-run
monarch profile apply minimal-desktop
```

## 🎮 Real-World Scenarios

### Scenario A: The Modern Developer (Your Setup)

Hardware: Lenovo IdeaPad (Ryzen 7, AMD Graphics)

OS: CachyOS

Goal: A clean, keyboard-driven development environment.

```
# 1. Create a fresh profile
monarch profile create dev-laptop

# 2. Add rulesets (Monarch auto-detects AMD GPU context)
monarch profile add ruleset dev-laptop official:hyprland
monarch profile add ruleset dev-laptop community:waybar-custom
monarch profile add ruleset dev-laptop community:rofi-lbonn

# 3. Apply
monarch profile apply dev-laptop
```

_Result: A fully configured Hyprland environment with Waybar and Rofi, optimized for AMD integrated graphics._

### Scenario B: The "Just for Tonight" Gamer

**Goal:** Install heavy gaming tools, play for the weekend, and remove them completely for work on Monday.

```
# Friday Night:
monarch profile apply hardcore-gaming

# Monday Morning:
monarch profile disable hardcore-gaming
```

_Result: Steam, Wine, Proton, and 20GB of dependencies are gone. No residual config files. No background services._

## 🏗️ Architecture

Monarch is built on the **Holy Trinity** of modern system orchestration:

1. **The Engine (CLI):** * Written in **Go**.
    
    - Distro-agnostic core with adapters for `pacman`, `apt`, and `dnf`.
        
    - Manages state using local JSON tracking and checksums.
        
2. **The Hub:**
    
    - A decentralized registry of Rulesets (hosted on GitHub).
        
    - Includes compatibility scoring (e.g., "This ruleset is 100% compatible with Wayland").
        
3. **The Studio (Coming Soon):**
    
    - A Wails-based GUI to visually construct profiles.
        

## 🗺️ Roadmap

|   |   |   |
|---|---|---|
|**Phase**|**Status**|**Focus Area**|
|**1**|✅|**Core Engine:** Resource adapters, State management, Undo logic.|
|**1.5**|🚧|**System Awareness:** Auto-detection of Hardware/OS, Profile Sync.|
|**2**|⏳|**Hub Ecosystem:** Compatibility Scoring, Community Repository.|
|**3**|🔮|**Monarch Studio:** GUI Dashboard & Visual Builder.|

## 🤝 Contributing

Monarch is designed to be community-driven.

- **Rule Creators:** Submit your custom Hyprland configs or Dev environments as Rulesets.
    
- **Go Developers:** Help us improve the resource adapters for non-Arch distros.
    

1. Fork the Project
    
2. Create your Feature Branch (`git checkout -b feature/AmazingFeature`)
    
3. Commit your Changes (`git commit -m 'Add some AmazingFeature'`)
    
4. Push to the Branch (`git push origin feature/AmazingFeature`)
    
5. Open a Pull Request
    

## 📜 License

Distributed under the Apache 2.0 License. See `LICENSE` for more information.

**Monarch** © 2025 Developed by **Melih Uçgun** _"We don't just configure systems—we build sovereign infrastructure."_
