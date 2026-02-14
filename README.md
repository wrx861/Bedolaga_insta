# Bedolaga Bot Installer

<p align="center">
  <img src="https://img.shields.io/badge/version-2.2.0-violet?style=for-the-badge" alt="Version">
  <img src="https://img.shields.io/badge/go-1.24+-00ADD8?style=for-the-badge&logo=go" alt="Go">
  <img src="https://img.shields.io/badge/platform-linux-FCC624?style=for-the-badge&logo=linux" alt="Linux">
  <img src="https://img.shields.io/badge/язык-русский-blue?style=for-the-badge" alt="Russian">
</p>

<p align="center">
  <b>Автоматический установщик Remnawave Bedolaga Telegram Bot</b><br>
  Премиум TUI интерфейс • Полная русская локализация • Один клик установка
</p>

---

## Быстрая установка

```bash
curl -fsSL https://raw.githubusercontent.com/wrx861/Bedolaga_insta/main/scripts/quick-install.sh | bash
```

Скрипт автоматически:
- Установит Go (если нужно)
- Скачает и соберёт установщик
- Запустит мастер установки

---

## Возможности

### Премиум интерфейс
- ASCII арт баннер
- Прогресс-бар с процентами (12 этапов)
- Навигация стрелками (bubbletea)
- Анимированные спиннеры
- Цветовая палитра Violet/Gold/Emerald

### Безопасность
- Защита от Ctrl+C — подтверждение перед выходом
- Восстановление при ошибках — не сбрасывает установку
- Автоопределение паролей PostgreSQL из существующих томов

### Функционал
- **2 режима установки**: с панелью / автономно
- **Автонастройка**: Nginx (системный/панели) или Caddy
- **Полный .env**: 200+ переменных конфигурации
- **SSL сертификаты**: Let's Encrypt через certbot
- **Управление**: команда `bot` с TUI-меню (стрелки)

---

## Требования

- **ОС**: Ubuntu 20.04+ / Debian 11+ (рекомендуется)
- **Доступ**: root пользователь
- **Память**: минимум 1 ГБ RAM
- **Диск**: минимум 5 ГБ свободного места

### Перед установкой подготовьте:
1. **BOT_TOKEN** — получить у [@BotFather](https://t.me/BotFather)
2. **Ваш Telegram ID** — узнать у [@userinfobot](https://t.me/userinfobot)
3. **REMNAWAVE_API_KEY** — из настроек панели Remnawave
4. **Домены** (опционально) — для webhook и Mini App

---

## Использование

### Установка
```bash
bedolaga_installer install
```

### Управление ботом (TUI)
```bash
bedolaga_installer manage
```

### Обновление бота
```bash
bedolaga_installer update
```

### Удаление
```bash
bedolaga_installer uninstall
```

---

## Команда управления `bot`

После установки доступна команда `bot` (wrapper над `bedolaga_installer manage`):

```bash
bot              # Интерактивное TUI-меню со стрелками
bot logs         # Просмотр логов
bot status       # Статус контейнеров
bot restart      # Перезапуск
bot start        # Запуск
bot stop         # Остановка
bot update       # Обновление
bot backup       # Создать бэкап
bot health       # Диагностика системы
bot config       # Редактировать .env
bot uninstall    # Удаление
```

---

## Структура исходного кода

```
├── main.go               # Точка входа + CLI-роутинг
├── commands.go            # Wizard + update + uninstall
├── manage.go              # TUI-панель управления ботом
├── management.go          # Генерация wrapper-скрипта bot
├── config.go              # Структура Config
├── progress.go            # Прогресс-трекер + обработка сигналов
├── utils.go               # Системные утилиты
├── system.go              # Проверки ОС + установка пакетов
├── setup.go               # Интерактивная настройка (12 шагов)
├── envfile.go             # Генерация .env (200+ переменных)
├── compose.go             # Docker Compose + клонирование репозитория
├── proxy.go               # Nginx + Caddy + SSL
├── docker.go              # Запуск Docker + firewall
├── main_test.go           # Unit-тесты (11 тестов)
├── go.mod / go.sum        # Go-модули
├── pkg/
│   └── ui/                # UI-пакет (переиспользуемый)
│       ├── styles.go      # Цвета + стили lipgloss
│       ├── banner.go      # ASCII баннер
│       ├── helpers.go     # Print-хелперы
│       ├── spinner.go     # Спиннер (bubbletea)
│       ├── select.go      # Выбор из списка (стрелки)
│       ├── input.go       # Текстовый ввод
│       ├── confirm.go     # Диалог подтверждения (Да/Нет)
│       ├── progress_bar.go # Прогресс-бар
│       └── utils.go       # IsInteractive()
├── scripts/               # Скрипт быстрой установки
└── dist/                  # Предсобранные бинарники
```

---

## Ручная сборка

```bash
git clone https://github.com/wrx861/Bedolaga_insta.git
cd Bedolaga_insta
go mod tidy
go build -o bedolaga_installer .
./bedolaga_installer install
```

---

## Changelog

### v2.2.0
- UI-компоненты вынесены в пакет `pkg/ui/` (9 файлов)
- Команда `manage` добавлена в CLI
- Скрипт `bot` теперь wrapper: `exec bedolaga_installer manage`
- TUI-панель управления с навигацией стрелками

### v2.1.0
- Модульная структура: 15 Go-файлов вместо одного монолита
- Каждый модуль отвечает за свою область (UI, Docker, Proxy, Config...)

### v2.0.1
- Полная русская локализация UI

### v2.0.0
- Премиум TUI интерфейс (bubbletea)
- Защита от Ctrl+C
- Команда управления `bot`

---

## Поддержка

- Telegram: [@bedolaga_support](https://t.me/bedolaga_support)
- Issues: [GitHub Issues](https://github.com/wrx861/Bedolaga_insta/issues)

---

## Лицензия

MIT License © 2024-2025 Bedolaga Dev Team
