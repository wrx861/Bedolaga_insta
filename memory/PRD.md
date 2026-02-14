# Bedolaga Bot Installer — PRD

## Original Problem Statement
Пользователь попросил изучить репозиторий от А до Я и затем разбить монолитный `main.go` на модули.
Далее — переработать скрипт `bot` для управления: навигация стрелками вместо цифр, продакшн-оформление.

## Architecture
- **Язык**: Go 1.24.4
- **UI фреймворк**: Charm ecosystem (bubbletea v1.3.10, lipgloss v1.1.0, bubbles v1.0.0)
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
- Скрипт управления `bot` с TUI-меню (стрелки, bubbletea)
- Команды update/uninstall/manage

## What's Been Implemented

### 2025-02-14: Модульная структура v2.1.0
- Разбит `main.go` (2543 строки) на 15 модулей
- Все 11 тестов проходят
- Сборка `go build` и `go vet` успешны
- README обновлён

### 2026-02-xx: TUI-управление bot v2.2.0
- Добавлена команда `manage` в `main.go` → вызов `manageBot()` из `manage.go`
- `management.go`: bash-скрипт заменён на однострочный wrapper `exec bedolaga_installer manage "$@"`
- `manage.go`: полноценная TUI-панель управления (11 пунктов, навигация стрелками, bubbletea)
- Help команда работает без установленного бота
- Все 11 тестов пройдены, сборка и go vet успешны

## File Structure
```
/app/
├── main.go               # Точка входа (install, manage, update, uninstall, version, help)
├── manage.go             # TUI-панель управления ботом (bubbletea)
├── management.go         # Генерация wrapper-скрипта bot + printFinalInfo
├── config.go             # Структура Config
├── styles.go             # Палитра цветов и стили lipgloss
├── progress.go           # Трекинг прогресса установки
├── ui.go                 # UI-компоненты (spinner, select, input, confirm, progress)
├── helpers.go            # Print-хелперы
├── utils.go              # Системные утилиты
├── system.go             # Проверки ОС и установка пакетов
├── setup.go              # Интерактивная настройка
├── envfile.go            # Генерация .env
├── compose.go            # Docker Compose логика
├── proxy.go              # Nginx/Caddy прокси
├── docker.go             # Docker команды и файрвол
├── commands.go           # Команды wizard, update, uninstall
├── main_test.go          # 11 юнит-тестов
├── go.mod / go.sum       # Зависимости Go
└── README.md             # Документация
```

## Prioritized Backlog

### P1 (Important)
- Добавить тесты для setup, proxy, compose модулей
- CI/CD через GitHub Actions + GoReleaser

### P2 (Nice to have)
- Вынести UI-компоненты в пакет `pkg/ui/`
- Добавить конфигурационный файл для дефолтов
- Поддержка --non-interactive флага для автоматизации
