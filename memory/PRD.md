# PRD: Bedolaga Installer (Go)

## Problem Statement
Create a Go-based CLI installer to automate the deployment of Remnawave Bedolaga Telegram Bot on Linux servers.

## Core Requirements
1. Interactive wizard for full bot configuration
2. Support 2 installation types: with panel (local Docker network) / standalone
3. Auto-configure Nginx (system + panel host mode) and Caddy
4. Full `.env` generation matching official `.env.example` (900+ lines)
5. No `sudo` — always runs as root
6. eGames SECRET_KEY support
7. Update command: `docker compose down && docker compose up -d --build && docker compose logs -f -t`
8. Management script (`bot` command) with interactive menu
9. Cross-platform builds (Linux/macOS, amd64/arm64)
10. Premium TUI design with bubbletea framework
11. Ctrl+C protection — confirmation before exit
12. Graceful error recovery — no installation reset on bad domain/DNS input
13. **Full Russian localization** — all UI text in Russian

## What's Implemented (Feb 13, 2026)

### v2.0.1 — Full Russian Localization
- [x] **Complete Russian UI** — all prompts, messages, menus, confirmations translated
  - "Really exit?" → "Точно выйти?"
  - "Yes/No" → "Да/Нет"
  - All step names, error messages, hints in Russian
  - Help text and commands in Russian
  - Banner text localized

### v2.0.0 — Premium TUI Redesign
- [x] **Premium UI** — bubbletea + lipgloss + bubbles framework
  - ASCII art banner with branded styling
  - Progress bar (step X/12 with percentage)
  - Arrow-key navigation menus (bubbletea list component)
  - Styled text inputs with placeholders
  - Yes/No toggle confirmations (← → keys)
  - Animated spinners for long operations
  - Styled boxes for info/success/error messages
  - Violet/Gold/Emerald color palette
- [x] **Ctrl+C protection** — SIGINT handler asks "Точно выйти?" before quitting
- [x] **Graceful error recovery** — bad domain/DNS offers "Попробовать снова / Использовать всё равно / Пропустить"
- [x] **Premium management script** (`/usr/local/bin/bot`) with styled interactive menu

### Core Functionality
- [x] Full Go installer (`main.go`, ~2500 lines with bubbletea)
- [x] Interactive wizard: install dir, panel check, Docker network auto-detect
- [x] Full interactive setup: BOT_TOKEN, ADMIN_IDS, API URL/KEY, auth type, eGames, webhook, miniapp, notifications, PostgreSQL
- [x] Full `.env` generation (all variables from official `.env.example`)
- [x] Docker Compose generation: standalone + local panel (with external network)
- [x] Nginx system mode + Nginx panel mode + Caddy mode
- [x] SSL via certbot (for nginx); Caddy auto-handles SSL
- [x] Firewall setup (UFW, optional)
- [x] Update command with correct docker compose sequence
- [x] Uninstall with backup option
- [x] DNS validation for domains
- [x] PostgreSQL volume detection and password recovery
- [x] Cross-compilation: Linux (amd64/arm64)

## Tech Stack
- Go 1.21+ (auto-upgraded to 1.25 toolchain)
- bubbletea v1.3.10 (TUI framework)
- lipgloss v1.1.0 (terminal styling)
- bubbles v1.0.0 (UI components: list, spinner, textinput, progress)

## Build Artifacts
- `/app/installer/dist/bedolaga-installer-linux-amd64`
- `/app/installer/dist/bedolaga-installer-linux-arm64`

## P1 Backlog
- Health check improvements
- Backup to Telegram integration

## P2 Backlog
- Windows/macOS build support
- Migration tool from old shell installer
