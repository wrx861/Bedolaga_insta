# 🤖 Bedolaga Bot Installer

<p align="center">
  <img src="https://img.shields.io/badge/version-2.0.1-violet?style=for-the-badge" alt="Version">
  <img src="https://img.shields.io/badge/go-1.21+-00ADD8?style=for-the-badge&logo=go" alt="Go">
  <img src="https://img.shields.io/badge/platform-linux-FCC624?style=for-the-badge&logo=linux" alt="Linux">
  <img src="https://img.shields.io/badge/язык-русский-blue?style=for-the-badge" alt="Russian">
</p>

<p align="center">
  <b>Автоматический установщик Remnawave Bedolaga Telegram Bot</b><br>
  Премиум TUI интерфейс • Полная русская локализация • Один клик установка
</p>

---

## ⚡ Быстрая установка

```bash
curl -fsSL https://raw.githubusercontent.com/wrx861/Bedolaga_insta/main/scripts/quick-install.sh | bash
```

Или скачайте бинарник напрямую:

```bash
# Для Linux AMD64
curl -fsSL https://github.com/wrx861/Bedolaga_insta/releases/latest/download/bedolaga-installer-linux-amd64 -o bedolaga_installer
chmod +x bedolaga_installer
./bedolaga_installer
```

```bash
# Для Linux ARM64 (Raspberry Pi, Oracle Cloud, etc.)
curl -fsSL https://github.com/wrx861/Bedolaga_insta/releases/latest/download/bedolaga-installer-linux-arm64 -o bedolaga_installer
chmod +x bedolaga_installer
./bedolaga_installer
```

---

## ✨ Возможности

### 🎨 Премиум интерфейс
- ASCII арт баннер
- Прогресс-бар с процентами (12 этапов)
- Навигация стрелками ↑↓
- Анимированные спиннеры
- Цветовая палитра Violet/Gold/Emerald

### 🛡️ Безопасность
- Защита от Ctrl+C — подтверждение перед выходом
- Восстановление при ошибках — не сбрасывает установку
- Автоопределение паролей PostgreSQL из существующих томов

### 🔧 Функционал
- **2 режима установки**: с панелью / автономно
- **Автонастройка**: Nginx (системный/панели) или Caddy
- **Полный .env**: 200+ переменных конфигурации
- **SSL сертификаты**: Let's Encrypt через certbot
- **Управление**: команда `bot` с интерактивным меню

---

## 📋 Требования

- **ОС**: Ubuntu 20.04+ / Debian 11+ (рекомендуется)
- **Доступ**: root пользователь
- **Память**: минимум 1 ГБ RAM
- **Диск**: минимум 5 ГБ свободного места

### Перед установкой подготовьте:
1. 🔑 **BOT_TOKEN** — получить у [@BotFather](https://t.me/BotFather)
2. 🆔 **Ваш Telegram ID** — узнать у [@userinfobot](https://t.me/userinfobot)
3. 🔐 **REMNAWAVE_API_KEY** — из настроек панели Remnawave
4. 🌐 **Домены** (опционально) — для webhook и Mini App

---

## 🚀 Использование

### Установка
```bash
./bedolaga_installer install
```

### Обновление бота
```bash
./bedolaga_installer update
```

### Удаление
```bash
./bedolaga_installer uninstall
```

### Справка
```bash
./bedolaga_installer help
```

---

## 🎛️ Команда управления `bot`

После установки доступна команда `bot`:

```bash
bot              # Интерактивное меню
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

## 📁 Структура установки

```
/opt/remnawave-bedolaga-telegram-bot/
├── .env                    # Конфигурация
├── docker-compose.yml      # Docker Compose (автономный режим)
├── docker-compose.local.yml # Docker Compose (с панелью)
├── logs/                   # Логи бота
├── data/                   # Данные
│   └── backups/           # Резервные копии
└── locales/               # Файлы локализации
```

---

## 🔄 Режимы установки

### С панелью Remnawave (на одном сервере)
- Бот подключается к панели через Docker-сеть
- Автоопределение сети панели
- Внутренний адрес: `http://remnawave:3000`

### Автономный режим
- Бот работает независимо
- Подключение к панели по внешнему URL
- Собственные контейнеры PostgreSQL и Redis

---

## 🛠️ Сборка из исходников

```bash
# Клонировать репозиторий
git clone https://github.com/wrx861/Bedolaga_insta.git
cd bedolaga_auto_install/installer

# Установить зависимости
go mod tidy

# Собрать
go build -o bedolaga_installer main.go

# Кросс-компиляция
GOOS=linux GOARCH=amd64 go build -o dist/bedolaga-installer-linux-amd64 main.go
GOOS=linux GOARCH=arm64 go build -o dist/bedolaga-installer-linux-arm64 main.go
```

---

## 📝 Changelog

### v2.0.1
- ✅ Полная русская локализация UI
- ✅ Исправлены все английские строки

### v2.0.0
- ✅ Премиум TUI интерфейс (bubbletea)
- ✅ Защита от Ctrl+C
- ✅ Восстановление при ошибках ввода
- ✅ Команда управления `bot`

---

## 🤝 Поддержка

- 📱 Telegram: [@bedolaga_support](https://t.me/bedolaga_support)
- 🐛 Issues: [GitHub Issues](https://github.com/wrx861/Bedolaga_insta/issues)

---

## 📄 Лицензия

MIT License © 2024-2025 Bedolaga Dev Team
