# Bedolaga Bot Installer — PRD

## Original Problem Statement
Пользователь попросил изучить репозиторий от А до Я, затем разбить монолитный `main.go` на модули,
затем переделать скрипт `bot` на TUI со стрелками, затем вынести UI-компоненты в `pkg/ui/`.

## Architecture
- **Язык**: Go 1.24.4
- **UI фреймворк**: Charm ecosystem (bubbletea v1.3.10, lipgloss v1.1.0, bubbles v1.0.0)
- **UI пакет**: `pkg/ui/` — 9 файлов (styles, banner, helpers, spinner, select, input, confirm, progress_bar, utils)
- **Платформа**: Linux (Ubuntu 20.04+/Debian 11+), root
- **Бинарники**: linux/amd64 и linux/arm64

## User Personas
- DevOps/сисадмины устанавливающие Remnawave Bedolaga Telegram Bot на VPS
- Разработчики Bedolaga, поддерживающие кодовую базу установщика

## Core Requirements (Static)
- TUI установщик с 12-шаговым визардом
- Поддержка 2 режимов: с панелью / автономно
- Настройка Nginx/Caddy + SSL
- Генерация .env с 200+ переменными
- Команда `manage` с TUI-меню (стрелки, bubbletea)
- Скрипт `bot` как wrapper: `exec bedolaga_installer manage`
- Команды update/uninstall

## What's Been Implemented

### 2025-02-14: Модульная структура v2.1.0
- Разбит `main.go` (2543 строки) на 15 модулей
- Все 11 тестов проходят

### 2026-02-xx: TUI-управление bot v2.2.0
- Добавлена команда `manage` в `main.go`
- `management.go`: bash-скрипт → wrapper `exec bedolaga_installer manage "$@"`
- `manage.go`: TUI-панель управления (11 пунктов, стрелки)
- UI-компоненты вынесены в `pkg/ui/` (9 файлов)
- Удалены старые `styles.go`, `helpers.go`, `ui.go`
- Все 11 тестов + go vet + go build успешны
- README.md обновлён

### 2026-12-xx: Caddy Docker-based fix v2.3.0
- **Исправлен критический баг**: Caddy теперь работает как Docker-контейнер вместо хост-сервиса
- Caddy находится в `bot_network` и проксирует на `remnawave_bot:8080` (внутренний Docker DNS)
- Генерируются файлы: `caddy/Caddyfile` + `docker-compose.caddy.yml`
- Автоматический HTTPS работает корректно
- Удалена зависимость от системного Caddy (`installCaddy()` не вызывается)

## Prioritized Backlog

### P1 (Important)
- Добавить тесты для `pkg/ui/` пакета
- Добавить тесты для setup, proxy, compose модулей
- CI/CD через GitHub Actions + GoReleaser

### P2 (Nice to have)
- Добавить конфигурационный файл для дефолтов
- Поддержка --non-interactive флага для автоматизации

## File Structure
```
/app/
├── main.go               # Точка входа (install, manage, update, uninstall)
├── commands.go            # Wizard + update + uninstall
├── manage.go              # TUI-панель управления ботом
├── management.go          # Генерация wrapper-скрипта bot
├── config.go              # Структура Config
├── progress.go            # Прогресс-трекер + сигналы
├── utils.go               # Системные утилиты
├── system.go              # ОС + пакеты
├── setup.go               # Интерактивная настройка
├── envfile.go             # .env генерация
├── compose.go             # Docker Compose
├── proxy.go               # Nginx/Caddy/SSL
├── docker.go              # Docker + firewall
├── main_test.go           # 11 тестов
├── pkg/ui/                # UI-пакет
│   ├── styles.go          # Цвета + стили
│   ├── banner.go          # Баннер
│   ├── helpers.go         # Print-хелперы
│   ├── spinner.go         # Спиннер
│   ├── select.go          # Список выбора
│   ├── input.go           # Текстовый ввод
│   ├── confirm.go         # Подтверждение
│   ├── progress_bar.go    # Прогресс-бар
│   └── utils.go           # IsInteractive
└── scripts/quick-install.sh
```
