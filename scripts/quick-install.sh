#!/bin/bash
# ╔══════════════════════════════════════════════════════════════╗
# ║   BEDOLAGA BOT — БЫСТРАЯ УСТАНОВКА                          ║
# ╚══════════════════════════════════════════════════════════════╝
# 
# Использование:
#   curl -fsSL https://raw.githubusercontent.com/wrx861/Bedolaga_insta/main/scripts/quick-install.sh | bash
#

set -e

# Сразу переходим в существующую директорию
cd /root 2>/dev/null || cd /tmp

# Цвета
P='\033[0;35m'
G='\033[0;32m'
R='\033[0;31m'
Y='\033[1;33m'
C='\033[0;36m'
D='\033[0;90m'
A='\033[38;5;214m'
NC='\033[0m'

# Баннер
echo -e "${P}"
echo "    ____  __________  ____  __    ___   _________  "
echo "   / __ )/ ____/ __ \/ __ \/ /   /   | / ____/   | "
echo "  / __  / __/ / / / / / / / /   / /| |/ / __/ /| | "
echo " / /_/ / /___/ /_/ / /_/ / /___/ ___ / /_/ / ___ | "
echo "/_____/_____/_____/\____/_____/_/  |_\____/_/  |_| "
echo -e "${NC}"
echo -e "${A}  УСТАНОВЩИК BEDOLAGA BOT${NC}  ${D}v2.2.0${NC}"
echo -e "${D}  ─────────────────────────────────────────────${NC}"
echo

# Проверка root
if [ "$EUID" -ne 0 ]; then
    echo -e "${R}  ✗ Этот скрипт должен быть запущен от root!${NC}"
    echo -e "${D}    Используйте: sudo bash или войдите как root${NC}"
    exit 1
fi

INSTALL_PATH="/usr/local/bin/bedolaga_installer"
REPO_URL="https://github.com/wrx861/Bedolaga_insta.git"
TMP_DIR="/tmp/bedolaga_installer_build"

# Проверяем наличие Go
install_go() {
    echo -e "${C}  ↓${NC} Установка Go..."
    
    ARCH=$(uname -m)
    case "$ARCH" in
        x86_64)
            GO_ARCH="amd64"
            ;;
        aarch64|arm64)
            GO_ARCH="arm64"
            ;;
        *)
            echo -e "${R}  ✗ Неподдерживаемая архитектура: $ARCH${NC}"
            exit 1
            ;;
    esac
    
    GO_VERSION="1.24.4"
    GO_URL="https://go.dev/dl/go${GO_VERSION}.linux-${GO_ARCH}.tar.gz"
    
    curl -fsSL "$GO_URL" -o /tmp/go.tar.gz
    rm -rf /usr/local/go
    tar -C /usr/local -xzf /tmp/go.tar.gz
    rm /tmp/go.tar.gz
    
    export PATH=/usr/local/go/bin:$PATH
    echo -e "${G}  ✓${NC} Go ${GO_VERSION} установлен"
}

# Проверяем Go (минимум 1.24)
GO_MIN_MAJOR=1
GO_MIN_MINOR=24
need_install=false

if command -v go &> /dev/null; then
    GO_VER=$(go version | awk '{print $3}' | sed 's/go//')
    GO_MAJOR=$(echo "$GO_VER" | cut -d. -f1)
    GO_MINOR=$(echo "$GO_VER" | cut -d. -f2)
    if [ "$GO_MAJOR" -lt "$GO_MIN_MAJOR" ] || ([ "$GO_MAJOR" -eq "$GO_MIN_MAJOR" ] && [ "$GO_MINOR" -lt "$GO_MIN_MINOR" ]); then
        echo -e "${Y}  ⚠${NC} Go ${GO_VER} устарел (нужен ≥1.24)"
        need_install=true
    else
        echo -e "${G}  ✓${NC} Go: ${C}${GO_VER}${NC}"
        export PATH=/usr/local/go/bin:$PATH
    fi
else
    need_install=true
fi

if [ "$need_install" = true ]; then
    install_go
fi

# Проверяем git
if ! command -v git &> /dev/null; then
    echo -e "${C}  ↓${NC} Установка git..."
    DEBIAN_FRONTEND=noninteractive apt-get update -qq && DEBIAN_FRONTEND=noninteractive apt-get install -y -qq git > /dev/null 2>&1
    echo -e "${G}  ✓${NC} Git установлен"
fi

# Клонируем и собираем
echo -e "${C}  ↓${NC} Клонирование репозитория..."
rm -rf "$TMP_DIR"
git clone --depth 1 -q "$REPO_URL" "$TMP_DIR"
echo -e "${G}  ✓${NC} Репозиторий клонирован"

echo -e "${C}  ⚙${NC} Сборка установщика..."
cd "$TMP_DIR"
export PATH=/usr/local/go/bin:$PATH
go build -o "$INSTALL_PATH" . 2>/dev/null
chmod +x "$INSTALL_PATH"
echo -e "${G}  ✓${NC} Установщик собран: ${C}${INSTALL_PATH}${NC}"

# Переходим в домашнюю директорию ПЕРЕД удалением tmp
cd /root

# Очистка
rm -rf "$TMP_DIR"

echo
echo -e "${D}  ─────────────────────────────────────────────${NC}"
echo

# Запуск установщика
exec "$INSTALL_PATH" install
