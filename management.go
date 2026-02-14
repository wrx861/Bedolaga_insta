package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// ════════════════════════════════════════════════════════════════
// MANAGEMENT SCRIPT
// ════════════════════════════════════════════════════════════════

func createManagementScript(cfg *Config) {
	composeFile := "docker-compose.yml"
	if cfg.PanelInstalledLocally {
		composeFile = "docker-compose.local.yml"
	}

	script := fmt.Sprintf(`#!/bin/bash
# ╔══════════════════════════════════════════════════════════════╗
# ║   REMNAWAVE BEDOLAGA BOT — УПРАВЛЕНИЕ                       ║
# ╚══════════════════════════════════════════════════════════════╝

INSTALL_DIR="%s"
COMPOSE_FILE="%s"

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

check_dir() { [ ! -d "$INSTALL_DIR" ] && echo -e "${R}  ✗ Бот не найден: $INSTALL_DIR${NC}" && exit 1; cd "$INSTALL_DIR"; }

do_logs()    { check_dir; echo -e "${C}  Логи (Ctrl+C для выхода)...${NC}"; docker compose -f "$COMPOSE_FILE" logs -f --tail=150 bot; }
do_status()  { check_dir; echo; echo -e "${W}  Статус контейнеров${NC}"; echo -e "${D}  --------------------------------------------------${NC}"; docker compose -f "$COMPOSE_FILE" ps --format "table {{.Name}}\t{{.Status}}\t{{.Ports}}" 2>/dev/null; echo; echo -e "${W}  Ресурсы${NC}"; echo -e "${D}  --------------------------------------------------${NC}"; docker stats --no-stream --format "table {{.Name}}\t{{.CPUPerc}}\t{{.MemUsage}}" 2>/dev/null | grep -E "remnawave|postgres|redis" || echo "  Нет данных"; }
do_restart() { check_dir; echo -e "${C}  Перезапуск...${NC}"; docker compose -f "$COMPOSE_FILE" restart && echo -e "${G}  ✓ Перезапущено${NC}"; }
do_start()   { check_dir; echo -e "${C}  Запуск...${NC}"; docker compose -f "$COMPOSE_FILE" up -d && echo -e "${G}  ✓ Запущено${NC}"; }
do_stop()    { check_dir; echo -e "${C}  Остановка...${NC}"; docker compose -f "$COMPOSE_FILE" down && echo -e "${G}  ✓ Остановлено${NC}"; }

do_update() {
    check_dir
    echo -e "${A}  Обновление бота...${NC}"
    cp .env ".env.backup_$(date +%%Y%%m%%d_%%H%%M%%S)" 2>/dev/null
    echo -e "${C}  Загрузка обновлений...${NC}"
    git pull origin main
    echo -e "${C}  Пересборка контейнеров...${NC}"
    docker compose -f "$COMPOSE_FILE" down && docker compose -f "$COMPOSE_FILE" up -d --build && docker compose -f "$COMPOSE_FILE" logs -f -t
}

do_backup() {
    check_dir
    local BK="$INSTALL_DIR/data/backups/$(date +%%Y%%m%%d_%%H%%M%%S)"
    mkdir -p "$BK"
    echo -e "${C}  Создание бэкапа...${NC}"
    docker compose -f "$COMPOSE_FILE" exec -T postgres pg_dump -U remnawave_user remnawave_bot > "$BK/database.sql" 2>/dev/null && echo -e "${G}  ✓ database.sql${NC}" || echo -e "${R}  ✗ Ошибка бэкапа БД${NC}"
    cp .env "$BK/.env" 2>/dev/null && echo -e "${G}  ✓ .env${NC}"
    cp docker-compose*.yml "$BK/" 2>/dev/null
    echo -e "${G}  ✓ Бэкап: $BK${NC}"
    ls -dt "$INSTALL_DIR/data/backups"/*/  2>/dev/null | tail -n +6 | xargs -r rm -rf
}

do_health() {
    check_dir
    echo
    echo -e "${A}  ╔══════════════════════════════════╗${NC}"
    echo -e "${A}  ║       ДИАГНОСТИКА СИСТЕМЫ         ║${NC}"
    echo -e "${A}  ╚══════════════════════════════════╝${NC}"
    echo
    docker ps --format '{{.Names}}' | grep -q "remnawave_bot$" && echo -e "${G}  ✓ Бот: работает${NC}" || echo -e "${R}  ✗ Бот: не запущен${NC}"
    docker compose -f "$COMPOSE_FILE" exec -T postgres pg_isready -U remnawave_user >/dev/null 2>&1 && echo -e "${G}  ✓ PostgreSQL: работает${NC}" || echo -e "${R}  ✗ PostgreSQL: не доступен${NC}"
    docker compose -f "$COMPOSE_FILE" exec -T redis redis-cli ping >/dev/null 2>&1 && echo -e "${G}  ✓ Redis: работает${NC}" || echo -e "${R}  ✗ Redis: не доступен${NC}"
    echo
    echo -e "${D}  Последние 10 строк логов:${NC}"
    docker compose -f "$COMPOSE_FILE" logs --tail=10 bot 2>/dev/null
}

do_config()    { check_dir; ${EDITOR:-nano} "$INSTALL_DIR/.env"; echo -e "${Y}  Перезапустите для применения: bot restart${NC}"; }

do_uninstall() {
    check_dir
    echo -e "${R}  ╔══════════════════════════════════╗${NC}"
    echo -e "${R}  ║         УДАЛЕНИЕ БОТА             ║${NC}"
    echo -e "${R}  ╚══════════════════════════════════╝${NC}"
    echo -e "${Y}  Это остановит и удалит контейнеры бота.${NC}"
    read -p "  Введите 'yes' для подтверждения: " CONFIRM
    [ "$CONFIRM" != "yes" ] && echo "  Отменено" && return
    docker compose -f "$COMPOSE_FILE" down
    read -p "  Удалить данные (тома)? (y/n): " -n 1 -r; echo
    [[ $REPLY =~ ^[Yy]$ ]] && docker compose -f "$COMPOSE_FILE" down -v
    [ -f "/usr/local/bin/bot" ] && rm -f /usr/local/bin/bot && echo -e "${G}  ✓ Команда 'bot' удалена${NC}"
    echo -e "${G}  ✓ Удаление завершено${NC}"
    echo -e "${Y}  Каталог сохранён: $INSTALL_DIR${NC}"
}

show_menu() {
    clear
    echo -e "${P}"
    echo "  ╔══════════════════════════════════════════════════════╗"
    echo "  ║                                                      ║"
    echo -e "  ║  ${A} REMNAWAVE BEDOLAGA BOT${P}                          ║"
    echo "  ║                                                      ║"
    echo "  ╚══════════════════════════════════════════════════════╝"
    echo -e "${NC}"
    echo -e "  ${D}Каталог:${NC} ${C}$INSTALL_DIR${NC}"
    docker ps --format '{{.Names}}' | grep -q "remnawave_bot$" && echo -e "  ${D}Статус:${NC}  ${G}● Работает${NC}" || echo -e "  ${D}Статус:${NC}  ${R}○ Остановлен${NC}"
    echo
    echo -e "  ${D}─────────────────────────────────────────${NC}"
    echo
    echo -e "  ${A}1${NC}  Логи              ${A}6${NC}  Бэкап"
    echo -e "  ${A}2${NC}  Статус            ${A}7${NC}  Диагностика"
    echo -e "  ${A}3${NC}  Перезапуск        ${A}8${NC}  Редактировать .env"
    echo -e "  ${A}4${NC}  Запуск            ${A}9${NC}  Обновить бота"
    echo -e "  ${A}5${NC}  Остановка         ${A}0${NC}  Удаление"
    echo
    echo -e "  ${D}─────────────────────────────────────────${NC}"
    echo -e "  ${D}q${NC}  Выход"
    echo
}

interactive_menu() {
    while true; do
        show_menu
        read -p "  $(echo -e ${A})>${NC} " choice
        case $choice in
            1) do_logs ;; 2) do_status; read -p "  Нажмите Enter..." ;; 3) do_restart; read -p "  Нажмите Enter..." ;;
            4) do_start; read -p "  Нажмите Enter..." ;; 5) do_stop; read -p "  Нажмите Enter..." ;; 6) do_backup; read -p "  Нажмите Enter..." ;;
            7) do_health; read -p "  Нажмите Enter..." ;; 8) do_config ;; 9) do_update; read -p "  Нажмите Enter..." ;;
            0) do_uninstall; break ;; q|Q) echo -e "  ${D}Пока!${NC}"; exit 0 ;; *) echo -e "${R}  Неверный выбор${NC}"; sleep 0.5 ;;
        esac
    done
}

show_help() {
    echo -e "${A}  REMNAWAVE BEDOLAGA BOT${NC}"
    echo -e "  ${D}Использование: bot [команда]${NC}"
    echo
    echo -e "  ${W}(без аргументов)${NC}  Интерактивное меню"
    echo -e "  ${W}logs${NC}              Просмотр логов"
    echo -e "  ${W}status${NC}            Статус контейнеров"
    echo -e "  ${W}restart${NC}           Перезапуск"
    echo -e "  ${W}start${NC}             Запуск"
    echo -e "  ${W}stop${NC}              Остановка"
    echo -e "  ${W}update${NC}            Обновление (git pull + пересборка)"
    echo -e "  ${W}backup${NC}            Создать бэкап"
    echo -e "  ${W}health${NC}            Диагностика системы"
    echo -e "  ${W}config${NC}            Редактировать .env"
    echo -e "  ${W}uninstall${NC}         Удалить бота"
}

case "$1" in
    logs) do_logs ;; status) do_status ;; restart) do_restart ;; start) do_start ;; stop) do_stop ;;
    update|upgrade) do_update ;; backup) do_backup ;; health|check) do_health ;; config|edit) do_config ;;
    uninstall|remove) do_uninstall ;; help|--help|-h) show_help ;; "") interactive_menu ;;
    *) echo -e "${R}  Неизвестная команда: $1${NC}"; echo "  Используйте: bot help"; exit 1 ;;
esac
`, cfg.InstallDir, composeFile)

	os.WriteFile("/usr/local/bin/bot", []byte(script), 0755)
	globalProgress.done("Команда 'bot' установлена")
}

// ════════════════════════════════════════════════════════════════
// FINAL INFO
// ════════════════════════════════════════════════════════════════

func printFinalInfo(cfg *Config) {
	sep := lipgloss.NewStyle().Foreground(colorAccent).Render(strings.Repeat("═", 56))

	var b strings.Builder
	b.WriteString(successStyle.Render("  УСТАНОВКА ЗАВЕРШЕНА") + "\n\n")
	b.WriteString(highlightStyle.Render("  Каталог: ") + infoStyle.Render(cfg.InstallDir) + "\n")
	b.WriteString(highlightStyle.Render("  Конфиг:  ") + infoStyle.Render(cfg.InstallDir+"/.env") + "\n\n")

	b.WriteString(highlightStyle.Render("  Управление:") + "\n")
	b.WriteString(dimStyle.Render("    bot          ") + "Интерактивное меню\n")
	b.WriteString(dimStyle.Render("    bot logs     ") + "Просмотр логов\n")
	b.WriteString(dimStyle.Render("    bot status   ") + "Статус контейнеров\n")
	b.WriteString(dimStyle.Render("    bot update   ") + "Обновить бота\n\n")

	if cfg.WebhookDomain != "" {
		b.WriteString(highlightStyle.Render("  Webhook: ") + infoStyle.Render("https://"+cfg.WebhookDomain) + "\n")
	}
	if cfg.MiniappDomain != "" {
		b.WriteString(highlightStyle.Render("  MiniApp: ") + infoStyle.Render("https://"+cfg.MiniappDomain) + "\n")
	}

	fmt.Println("\n  " + sep)
	fmt.Println(successBoxStyle.Render(b.String()))
	fmt.Println("  " + sep)
}
