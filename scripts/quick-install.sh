#!/bin/bash
# ╔══════════════════════════════════════════════════════════════╗
# ║   BEDOLAGA BOT — БЫСТРАЯ УСТАНОВКА                          ║
# ╚══════════════════════════════════════════════════════════════╝
# 
# Использование:
#   curl -fsSL https://raw.githubusercontent.com/wrx861/Bedolaga_insta/main/scripts/quick-install.sh | bash
#

set -e

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
echo -e "${A}  УСТАНОВЩИК BEDOLAGA BOT${NC}  ${D}v2.0.1${NC}"
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
    
    GO_VERSION="1.21.11"
    GO_URL="https://go.dev/dl/go${GO_VERSION}.linux-${GO_ARCH}.tar.gz"
    
    curl -fsSL "$GO_URL" -o /tmp/go.tar.gz
    rm -rf /usr/local/go
    tar -C /usr/local -xzf /tmp/go.tar.gz
    rm /tmp/go.tar.gz
    
    export PATH=/usr/local/go/bin:$PATH
    echo -e "${G}  ✓${NC} Go ${GO_VERSION} установлен"
}

# Проверяем Go
if command -v go &> /dev/null; then
    GO_VER=$(go version | awk '{print $3}' | sed 's/go//')
    echo -e "${G}  ✓${NC} Go: ${C}${GO_VER}${NC}"
    export PATH=/usr/local/go/bin:$PATH
else
    install_go
fi

# Проверяем git
if ! command -v git &> /dev/null; then
    echo -e "${C}  ↓${NC} Установка git..."
    apt-get update -qq && apt-get install -y -qq git > /dev/null 2>&1
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
go build -o "$INSTALL_PATH" main.go 2>/dev/null
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
