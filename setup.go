package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"bedolaga-installer/pkg/ui"
)

// ════════════════════════════════════════════════════════════════
// INTERACTIVE SETUP STEPS
// ════════════════════════════════════════════════════════════════

func selectInstallDir(cfg *Config) {
	idx := ui.SelectOption("Каталог установки", []ui.SelectItem{
		{Title: "/opt/remnawave-bedolaga-telegram-bot", Description: "Рекомендуемое расположение"},
		{Title: "/root/remnawave-bedolaga-telegram-bot", Description: "Домашний каталог"},
		{Title: "Свой путь", Description: "Указать свой путь"},
	})
	switch idx {
	case 0:
		cfg.InstallDir = "/opt/remnawave-bedolaga-telegram-bot"
	case 1:
		cfg.InstallDir = "/root/remnawave-bedolaga-telegram-bot"
	case 2:
		cfg.InstallDir = ui.InputText("Путь установки", "/opt/my-bot", "Введите полный путь", true)
	}
	globalProgress.info("Каталог: " + ui.HighlightStyle.Render(cfg.InstallDir))
}

func checkRemnawavePanel(cfg *Config) {
	idx := ui.SelectOption("Расположение панели", []ui.SelectItem{
		{Title: "Панель на этом сервере", Description: "Бот подключается через Docker-сеть"},
		{Title: "Панель на другом сервере", Description: "Бот подключается через внешний URL"},
	})
	switch idx {
	case 0:
		cfg.PanelInstalledLocally = true
		setupLocalPanel(cfg)
	case 1:
		cfg.PanelInstalledLocally = false
		globalProgress.info("Автономный режим — укажите внешний URL при настройке")
	}
}

func setupLocalPanel(cfg *Config) {
	idx := ui.SelectOption("Каталог панели", []ui.SelectItem{
		{Title: "/opt/remnawave", Description: "Стандартный путь установки"},
		{Title: "/root/remnawave", Description: "Домашний каталог"},
		{Title: "Свой путь", Description: "Указать свой путь"},
	})
	switch idx {
	case 0:
		cfg.PanelDir = "/opt/remnawave"
	case 1:
		cfg.PanelDir = "/root/remnawave"
	case 2:
		cfg.PanelDir = ui.InputText("Путь к каталогу панели", "/opt/remnawave", "", true)
	}

	if !dirExists(cfg.PanelDir) {
		globalProgress.warn("Каталог " + cfg.PanelDir + " не найден")
		globalProgress.info("Переключаемся на режим внешней панели — укажите URL позже")
		cfg.PanelInstalledLocally = false
		cfg.PanelDir = ""
		cfg.DockerNetwork = ""
		return
	}

	globalProgress.done("Панель найдена: " + cfg.PanelDir)
	detectPanelNetwork(cfg)
}

func detectPanelNetwork(cfg *Config) {
	found := false

	if out, err := runShellSilent(`docker inspect remnawave --format '{{range $net, $config := .NetworkSettings.Networks}}{{$net}}{{"\n"}}{{end}}' 2>/dev/null | grep -v "^$" | grep -v "host" | grep -v "none" | head -1`); err == nil && out != "" {
		cfg.DockerNetwork = out
		found = true
	}

	if !found {
		known := []string{"remnawave-network", "remnawave_default", "remnawave_network", "remnawave", "remnawave-panel_default"}
		for _, n := range known {
			if _, err := runShellSilent(fmt.Sprintf("docker network inspect %s 2>/dev/null", n)); err == nil {
				cfg.DockerNetwork = n
				found = true
				break
			}
		}
	}

	if !found {
		if out, err := runShellSilent(`docker network ls --format '{{.Name}}' | grep -i "remnawave" | grep -v "bedolaga" | grep -v "bot" | head -1`); err == nil && out != "" {
			cfg.DockerNetwork = out
			found = true
		}
	}

	if found {
		globalProgress.done("Docker-сеть: " + cfg.DockerNetwork)
	} else {
		globalProgress.warn("Docker-сеть не найдена автоматически")
		cfg.DockerNetwork = "remnawave-network"
		globalProgress.info("Используется сеть по умолчанию: " + cfg.DockerNetwork)
	}
}

func inputDomainSafe(label, hint string) string {
	for {
		val := ui.InputText(label, "bot.example.com", hint, false)
		if val == "" {
			return ""
		}
		val = cleanDomain(val)
		if !validateDomain(val) {
			ui.PrintError("Неверный формат домена: " + val)
			ui.PrintDim("Ожидаемый формат: bot.example.com")
			if !ui.IsInteractive() {
				return ""
			}
			idx := ui.SelectOption("Что делать?", []ui.SelectItem{
				{Title: "Попробовать снова", Description: "Ввести другой домен"},
				{Title: "Использовать всё равно", Description: "Продолжить с этим значением"},
				{Title: "Пропустить", Description: "Не настраивать этот домен"},
			})
			switch idx {
			case 0:
				continue
			case 1:
				return val
			case 2:
				return ""
			}
		}
		ui.PrintInfo("Проверка DNS...")
		if !checkDomainDNS(val) {
			ui.PrintWarning("DNS не указывает на этот сервер")
			if !ui.IsInteractive() {
				return val
			}
			idx := ui.SelectOption("Что делать?", []ui.SelectItem{
				{Title: "Попробовать другой домен", Description: "Ввести другой домен"},
				{Title: "Продолжить с этим доменом", Description: "DNS можно настроить позже"},
				{Title: "Пропустить", Description: "Не настраивать этот домен"},
			})
			switch idx {
			case 0:
				continue
			case 1:
				return val
			case 2:
				return ""
			}
		}
		return val
	}
}

func checkPostgresVolume(cfg *Config) {
	cfg.KeepExistingVolumes = false
	cfg.OldPostgresPassword = ""

	if fileExists(filepath.Join(cfg.InstallDir, ".env")) {
		if out, err := runShellSilent(fmt.Sprintf(`grep "^POSTGRES_PASSWORD=" "%s/.env" 2>/dev/null | cut -d'=' -f2- | tr -d '"' | tr -d "'"`, cfg.InstallDir)); err == nil && out != "" {
			cfg.OldPostgresPassword = out
		}
	}

	foundVolumes, _ := runShellSilent(`docker volume ls -q 2>/dev/null | grep -E "(postgres|bot)" | grep -v "remnawave_postgres" || true`)
	if strings.TrimSpace(foundVolumes) == "" {
		globalProgress.info("Чистая установка — существующих томов нет")
		return
	}

	globalProgress.warn("Найдены существующие Docker-тома")
	if cfg.OldPostgresPassword != "" {
		globalProgress.done("Найден старый пароль PostgreSQL")
		cfg.KeepExistingVolumes = true
	} else {
		runShellSilent(fmt.Sprintf("cd %s 2>/dev/null && docker compose down -v 2>/dev/null || true", cfg.InstallDir))
	}
}

func interactiveSetup(cfg *Config) {
	ui.PrintBox("⚙️  Интерактивная настройка",
		"Введите необходимые данные для настройки бота.\n"+
			ui.DimStyle.Render("Необязательные поля можно пропустить клавишей Esc."))

	cfg.BotToken = ui.InputText("BOT_TOKEN", "123456:ABC-DEF...", "Получить у @BotFather в Telegram", true)
	cfg.AdminIDs = ui.InputText("ADMIN_IDS", "123456789", "Ваш Telegram ID (несколько: 123,456). Узнать у @userinfobot", true)

	if cfg.PanelInstalledLocally && cfg.DockerNetwork != "" {
		ui.PrintInfo("Локальная панель — используется внутренний Docker-адрес")
		val := ui.InputText("REMNAWAVE_API_URL", "http://remnawave:3000", "Внутренний адрес для локальной панели", false)
		if val == "" {
			val = "http://remnawave:3000"
		}
		cfg.RemnawaveAPIURL = val
	} else {
		cfg.RemnawaveAPIURL = ui.InputText("REMNAWAVE_API_URL", "https://panel.yourdomain.com", "Внешний URL панели Remnawave", true)
	}

	cfg.RemnawaveAPIKey = ui.InputText("REMNAWAVE_API_KEY", "", "Получить в настройках панели Remnawave", true)

	idx := ui.SelectOption("Тип авторизации", []ui.SelectItem{
		{Title: "API Key", Description: "По умолчанию — только API-ключ"},
		{Title: "Basic Auth", Description: "Авторизация по логину и паролю"},
	})
	cfg.RemnawaveAuthType = "api_key"
	if idx == 1 {
		cfg.RemnawaveAuthType = "basic_auth"
		cfg.RemnawaveUsername = ui.InputText("REMNAWAVE_USERNAME", "", "", false)
		cfg.RemnawavePassword = ui.InputText("REMNAWAVE_PASSWORD", "", "", false)
	}

	cfg.WebhookDomain = inputDomainSafe("Домен вебхука (необязательно)", "Для режима webhook. Оставьте пустым для polling.")
	cfg.MiniappDomain = inputDomainSafe("Домен Mini App (необязательно)", "Домен для Telegram Mini App")
	cfg.AdminNotificationsChatID = ui.InputText("Chat ID уведомлений (необязательно)", "-1001234567890", "ID чата/группы Telegram для уведомлений администратора", false)

	if cfg.KeepExistingVolumes && cfg.OldPostgresPassword != "" {
		cfg.PostgresPassword = cfg.OldPostgresPassword
		ui.PrintSuccess("PostgreSQL: используется сохранённый пароль")
	} else {
		pw := ui.InputText("Пароль PostgreSQL (необязательно)", "", "Оставьте пустым для автогенерации безопасного пароля", false)
		if pw == "" {
			cfg.PostgresPassword = generateSafePassword(24)
			ui.PrintSuccess("Сгенерирован безопасный пароль PostgreSQL")
		} else {
			cfg.PostgresPassword = pw
		}
	}

	if cfg.WebhookDomain != "" || cfg.MiniappDomain != "" {
		proxyItems := []ui.SelectItem{
			{Title: "Nginx (системный)", Description: "Автономный nginx на сервере"},
			{Title: "Caddy", Description: "Автоматический HTTPS, простая настройка"},
			{Title: "Пропустить", Description: "Настроить вручную позже"},
		}
		if cfg.PanelInstalledLocally {
			nginxNet, _ := runShellSilent("docker inspect remnawave-nginx --format '{{.HostConfig.NetworkMode}}' 2>/dev/null")
			if strings.TrimSpace(nginxNet) == "host" {
				proxyItems = []ui.SelectItem{
					{Title: "Nginx (панели)", Description: "Добавить в nginx панели (host mode)"},
					{Title: "Nginx (системный)", Description: "Автономный nginx на сервере"},
					{Title: "Caddy", Description: "Автоматический HTTPS, простая настройка"},
					{Title: "Пропустить", Description: "Настроить вручную позже"},
				}
			}
		}

		idx := ui.SelectOption("Обратный прокси", proxyItems)
		title := proxyItems[idx].Title
		switch {
		case strings.Contains(title, "панели"):
			cfg.ReverseProxyType = "nginx_panel"
		case strings.Contains(title, "системный"):
			cfg.ReverseProxyType = "nginx_system"
		case strings.Contains(title, "Caddy"):
			cfg.ReverseProxyType = "caddy"
		default:
			cfg.ReverseProxyType = "skip"
		}
	} else {
		cfg.ReverseProxyType = "skip"
	}

	cfg.WebhookSecretToken = generateToken()
	cfg.WebAPIDefaultToken = generateToken()
	cfg.SupportUsername = "@support"

	if cfg.WebhookDomain != "" {
		cfg.BotRunMode = "webhook"
		cfg.WebhookURL = "https://" + cfg.WebhookDomain
		cfg.WebAPIEnabled = "true"
	} else {
		cfg.BotRunMode = "polling"
		cfg.WebhookURL = ""
		cfg.WebAPIEnabled = "false"
	}

	ui.PrintSuccessBox(ui.SuccessStyle.Render("Настройка завершена!"))
}
