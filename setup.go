package main

import (
	"fmt"
	"path/filepath"
	"strings"
)

// ════════════════════════════════════════════════════════════════
// INTERACTIVE SETUP STEPS
// ════════════════════════════════════════════════════════════════

func selectInstallDir(cfg *Config) {
	idx := selectOption("Каталог установки", []selectItem{
		{title: "/opt/remnawave-bedolaga-telegram-bot", description: "Рекомендуемое расположение"},
		{title: "/root/remnawave-bedolaga-telegram-bot", description: "Домашний каталог"},
		{title: "Свой путь", description: "Указать свой путь"},
	})
	switch idx {
	case 0:
		cfg.InstallDir = "/opt/remnawave-bedolaga-telegram-bot"
	case 1:
		cfg.InstallDir = "/root/remnawave-bedolaga-telegram-bot"
	case 2:
		cfg.InstallDir = inputText("Путь установки", "/opt/my-bot", "Введите полный путь", true)
	}
	globalProgress.info("Каталог: " + highlightStyle.Render(cfg.InstallDir))
}

func checkRemnawavePanel(cfg *Config) {
	idx := selectOption("Расположение панели", []selectItem{
		{title: "Панель на этом сервере", description: "Бот подключается через Docker-сеть"},
		{title: "Панель на другом сервере", description: "Бот подключается через внешний URL"},
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
	idx := selectOption("Каталог панели", []selectItem{
		{title: "/opt/remnawave", description: "Стандартный путь установки"},
		{title: "/root/remnawave", description: "Домашний каталог"},
		{title: "Свой путь", description: "Указать свой путь"},
	})
	switch idx {
	case 0:
		cfg.PanelDir = "/opt/remnawave"
	case 1:
		cfg.PanelDir = "/root/remnawave"
	case 2:
		cfg.PanelDir = inputText("Путь к каталогу панели", "/opt/remnawave", "", true)
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

	// Method 1: by running container
	if out, err := runShellSilent(`docker inspect remnawave --format '{{range $net, $config := .NetworkSettings.Networks}}{{$net}}{{"\n"}}{{end}}' 2>/dev/null | grep -v "^$" | grep -v "host" | grep -v "none" | head -1`); err == nil && out != "" {
		cfg.DockerNetwork = out
		found = true
	}

	// Method 2: known names
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

	// Method 3: grep
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
		// Используем дефолтную сеть
		cfg.DockerNetwork = "remnawave-network"
		globalProgress.info("Используется сеть по умолчанию: " + cfg.DockerNetwork)
	}
}

func inputDomainSafe(label, hint string) string {
	for {
		val := inputText(label, "bot.example.com", hint, false)
		if val == "" {
			return ""
		}
		val = cleanDomain(val)
		if !validateDomain(val) {
			printError("Неверный формат домена: " + val)
			printDim("Ожидаемый формат: bot.example.com")
			if !isInteractive() {
				// В неинтерактивном режиме пропускаем
				return ""
			}
			idx := selectOption("Что делать?", []selectItem{
				{title: "Попробовать снова", description: "Ввести другой домен"},
				{title: "Использовать всё равно", description: "Продолжить с этим значением"},
				{title: "Пропустить", description: "Не настраивать этот домен"},
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
		printInfo("Проверка DNS...")
		if !checkDomainDNS(val) {
			printWarning("DNS не указывает на этот сервер")
			if !isInteractive() {
				// В неинтерактивном режиме продолжаем с этим доменом
				return val
			}
			idx := selectOption("Что делать?", []selectItem{
				{title: "Попробовать другой домен", description: "Ввести другой домен"},
				{title: "Продолжить с этим доменом", description: "DNS можно настроить позже"},
				{title: "Пропустить", description: "Не настраивать этот домен"},
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
	printBox("⚙️  Интерактивная настройка",
		"Введите необходимые данные для настройки бота.\n"+
			dimStyle.Render("Необязательные поля можно пропустить клавишей Esc."))

	// 1. BOT_TOKEN
	cfg.BotToken = inputText("BOT_TOKEN", "123456:ABC-DEF...", "Получить у @BotFather в Telegram", true)

	// 2. ADMIN_IDS
	cfg.AdminIDs = inputText("ADMIN_IDS", "123456789", "Ваш Telegram ID (несколько: 123,456). Узнать у @userinfobot", true)

	// 3. REMNAWAVE_API_URL
	if cfg.PanelInstalledLocally && cfg.DockerNetwork != "" {
		printInfo("Локальная панель — используется внутренний Docker-адрес")
		val := inputText("REMNAWAVE_API_URL", "http://remnawave:3000", "Внутренний адрес для локальной панели", false)
		if val == "" {
			val = "http://remnawave:3000"
		}
		cfg.RemnawaveAPIURL = val
	} else {
		cfg.RemnawaveAPIURL = inputText("REMNAWAVE_API_URL", "https://panel.yourdomain.com", "Внешний URL панели Remnawave", true)
	}

	// 4. REMNAWAVE_API_KEY
	cfg.RemnawaveAPIKey = inputText("REMNAWAVE_API_KEY", "", "Получить в настройках панели Remnawave", true)

	// 5. Auth type
	idx := selectOption("Тип авторизации", []selectItem{
		{title: "API Key", description: "По умолчанию — только API-ключ"},
		{title: "Basic Auth", description: "Авторизация по логину и паролю"},
	})
	cfg.RemnawaveAuthType = "api_key"
	if idx == 1 {
		cfg.RemnawaveAuthType = "basic_auth"
		cfg.RemnawaveUsername = inputText("REMNAWAVE_USERNAME", "", "", false)
		cfg.RemnawavePassword = inputText("REMNAWAVE_PASSWORD", "", "", false)
	}

	// 6. Webhook domain
	cfg.WebhookDomain = inputDomainSafe("Домен вебхука (необязательно)", "Для режима webhook. Оставьте пустым для polling.")

	// 8. Miniapp domain
	cfg.MiniappDomain = inputDomainSafe("Домен Mini App (необязательно)", "Домен для Telegram Mini App")

	// 9. Notifications
	cfg.AdminNotificationsChatID = inputText("Chat ID уведомлений (необязательно)", "-1001234567890", "ID чата/группы Telegram для уведомлений администратора", false)

	// 10. PostgreSQL password
	if cfg.KeepExistingVolumes && cfg.OldPostgresPassword != "" {
		cfg.PostgresPassword = cfg.OldPostgresPassword
		printSuccess("PostgreSQL: используется сохранённый пароль")
	} else {
		pw := inputText("Пароль PostgreSQL (необязательно)", "", "Оставьте пустым для автогенерации безопасного пароля", false)
		if pw == "" {
			cfg.PostgresPassword = generateSafePassword(24)
			printSuccess("Сгенерирован безопасный пароль PostgreSQL")
		} else {
			cfg.PostgresPassword = pw
		}
	}

	// 11. Reverse proxy
	if cfg.WebhookDomain != "" || cfg.MiniappDomain != "" {
		proxyItems := []selectItem{
			{title: "Nginx (системный)", description: "Автономный nginx на сервере"},
			{title: "Caddy", description: "Автоматический HTTPS, простая настройка"},
			{title: "Пропустить", description: "Настроить вручную позже"},
		}
		if cfg.PanelInstalledLocally {
			nginxNet, _ := runShellSilent("docker inspect remnawave-nginx --format '{{.HostConfig.NetworkMode}}' 2>/dev/null")
			if strings.TrimSpace(nginxNet) == "host" {
				proxyItems = []selectItem{
					{title: "Nginx (панели)", description: "Добавить в nginx панели (host mode)"},
					{title: "Nginx (системный)", description: "Автономный nginx на сервере"},
					{title: "Caddy", description: "Автоматический HTTPS, простая настройка"},
					{title: "Пропустить", description: "Настроить вручную позже"},
				}
			}
		}

		idx := selectOption("Обратный прокси", proxyItems)
		title := proxyItems[idx].title
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

	// Generate tokens
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

	printSuccessBox(successStyle.Render("Настройка завершена!"))
}
