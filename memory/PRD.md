# Bedolaga Bot Installer — PRD

## Original Problem Statement
Пользователь попросил изучить репозиторий от А до Я и затем разбить монолитный `main.go` на модули.

## Architecture
- **Язык**: Go 1.24.2
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
- Скрипт управления `bot` с интерактивным меню
- Команды update/uninstall

## What's Been Implemented
### 2025-02-14: Модульная структура v2.1.0
- Разбит `main.go` (2543 строки) на 15 модулей
- Все 11 тестов проходят
- Сборка `go build` и `go vet` успешны
- README обновлён

## Prioritized Backlog
### P0 (Critical)
- (нет — основной функционал работает)

### P1 (Important)
- Добавить тесты для setup, proxy, compose модулей
- CI/CD через GitHub Actions + GoReleaser

### P2 (Nice to have)
- Вынести UI-компоненты в пакет `pkg/ui/`
- Добавить конфигурационный файл для дефолтов
- Поддержка --non-interactive флага для автоматизации

## Next Tasks
1. Расширить покрытие тестами
2. Настроить CI/CD
