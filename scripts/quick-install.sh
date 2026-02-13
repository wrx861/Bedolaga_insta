#!/bin/bash
# ╔══════════════════════════════════════════════════════════════╗
# ║   BEDOLAGA BOT — БЫСТРАЯ УСТАНОВКА                          ║
# ╚══════════════════════════════════════════════════════════════╝
# 
# Использование:
#   curl -fsSL https://raw.githubusercontent.com/wrx861/bedolaga_auto_install/main/scripts/quick-install.sh | bash
#

set -e

# Цвета
P='\033[0;35m'   # Purple
G='\033[0;32m'   # Green
R='\033[0;31m'   # Red
Y='\033[1;33m'   # Yellow
C='\033[0;36m'   # Cyan
W='\033[1;37m'   # White
D='\033[0;90m'   # Dim
A='\033[38;5;214m' # Amber
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

# Определение архитектуры
ARCH=$(uname -m)
case "$ARCH" in
    x86_64)
        BINARY="bedolaga-installer-linux-amd64"
        echo -e "${G}  ✓${NC} Архитектура: ${C}amd64${NC}"
        ;;
    aarch64|arm64)
        BINARY="bedolaga-installer-linux-arm64"
        echo -e "${G}  ✓${NC} Архитектура: ${C}arm64${NC}"
        ;;
    *)
        echo -e "${R}  ✗ Неподдерживаемая архитектура: $ARCH${NC}"
        exit 1
        ;;
esac

# URL для загрузки
DOWNLOAD_URL="https://github.com/wrx861/Bedolaga_insta/releases/latest/download/${BINARY}"
INSTALL_PATH="/usr/local/bin/bedolaga_installer"

echo -e "${C}  ↓${NC} Загрузка установщика..."

# Загрузка
if command -v curl &> /dev/null; then
    curl -fsSL "$DOWNLOAD_URL" -o "$INSTALL_PATH"
elif command -v wget &> /dev/null; then
    wget -q "$DOWNLOAD_URL" -O "$INSTALL_PATH"
else
    echo -e "${R}  ✗ Требуется curl или wget${NC}"
    echo -e "${D}    Установите: apt-get install -y curl${NC}"
    exit 1
fi

# Права на выполнение
chmod +x "$INSTALL_PATH"

echo -e "${G}  ✓${NC} Установщик загружен: ${C}${INSTALL_PATH}${NC}"
echo
echo -e "${D}  ─────────────────────────────────────────────${NC}"
echo

# Запуск установщика
exec "$INSTALL_PATH" install
