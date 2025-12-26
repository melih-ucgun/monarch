# Monarch 🏰

**Your System, Your Rules. Enforced by Monarch.**

Monarch is a declarative, agentless configuration management tool written in Go.  
Designed to manage everything from your personal desktop to remote servers with a single configuration.

> Write once, deploy anywhere — laptop, homelab, VPS, containers.

## 🚀 Vision
- Declarative state management (like NixOS, but for any Linux distro)
- Agentless (SSH only)
- Single static binary (no dependencies)
- Built-in secret management with age
- Drift detection and auto-healing

## 📦 Current Status
**Early development** – MVP in progress (v0.1.0 target: Q1 2026)

Implemented so far:
- [x] CLI structure (Cobra)
- [x] YAML config parsing
- [ ] File resource
- [ ] Package resource
- [ ] Service resource
- [ ] Remote SSH execution
- [ ] Watch mode

## 🛠 Planned Roadmap
- v0.1: Core resources + local/remote apply
- v0.2: Templating + more resources
- v1.0: Monarch Studio (GUI config editor)
- Future: Monarch Hub (central fleet management)

## 🧑‍💻 Contributing
Contributions welcome! See [CONTRIBUTING.md](#) for details.

## 📄 License
AGPL-3.0 — see [LICENSE](LICENSE)

---

Made with ❤️ for sovereign systems.
