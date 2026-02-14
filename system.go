package main

import (
	"fmt"
	"os"

	"bedolaga-installer/pkg/ui"
)

// ════════════════════════════════════════════════════════════════
// SYSTEM CHECKS
// ════════════════════════════════════════════════════════════════

func checkRoot() {
	if os.Getuid() != 0 {
		ui.PrintErrorBox(ui.ErrorStyle.Render("Этот скрипт должен быть запущен от root!"))
		os.Exit(1)
	}
}

func detectOS() string {
	out, _ := runShellSilent("cat /etc/os-release 2>/dev/null | grep ^ID= | cut -d= -f2 | tr -d '\"'")
	prettyName, _ := runShellSilent("cat /etc/os-release 2>/dev/null | grep ^PRETTY_NAME= | cut -d= -f2 | tr -d '\"'")
	if prettyName != "" {
		globalProgress.info("ОС: " + prettyName)
	}
	switch out {
	case "ubuntu", "debian":
		return out
	default:
		if out != "" {
			globalProgress.warn("Оптимизировано для Ubuntu/Debian. Обнаружено: " + out)
			if !ui.ConfirmPrompt("Продолжить на неподдерживаемой ОС?", false) {
				os.Exit(0)
			}
		}
		return out
	}
}

// ════════════════════════════════════════════════════════════════
// PACKAGE INSTALLATION
// ════════════════════════════════════════════════════════════════

func updateSystem() {
	runShellSilent("DEBIAN_FRONTEND=noninteractive apt-get update -y -qq 2>/dev/null")
}

func installBasePackages() {
	packages := []string{
		"curl wget git",
		"nano htop",
		"make openssl ca-certificates gnupg",
		"lsb-release dnsutils",
	}
	for _, pkg := range packages {
		runShellSilent(fmt.Sprintf("DEBIAN_FRONTEND=noninteractive apt-get install -y -qq %s 2>/dev/null || true", pkg))
	}
	runShellSilent("DEBIAN_FRONTEND=noninteractive apt-get install -y -qq certbot python3-certbot-nginx 2>/dev/null || true")
}

func installDocker() {
	if commandExists("docker") {
		ver, _ := runShellSilent("docker --version")
		globalProgress.done("Docker: " + ver)
	} else {
		runShellSilent("DEBIAN_FRONTEND=noninteractive curl -fsSL https://get.docker.com | sh")
		runShellSilent("systemctl enable docker 2>/dev/null || true")
		runShellSilent("systemctl start docker 2>/dev/null || true")

		if !commandExists("docker") {
			globalProgress.fail("Не удалось установить Docker!")
			globalProgress.info("Попробуйте установить Docker вручную: curl -fsSL https://get.docker.com | sh")
			os.Exit(1)
		}
		ver, _ := runShellSilent("docker --version")
		globalProgress.done("Docker установлен: " + ver)
	}

	if out, err := runShellSilent("docker compose version 2>/dev/null"); err == nil && out != "" {
		globalProgress.done("Docker Compose: " + out)
	} else if out, err := runShellSilent("docker-compose --version 2>/dev/null"); err == nil && out != "" {
		globalProgress.done("Docker Compose (standalone): " + out)
	} else {
		globalProgress.fail("Docker Compose не найден!")
		globalProgress.info("Установите Docker Compose: apt install docker-compose-plugin")
		os.Exit(1)
	}
}

func installNginx() {
	if commandExists("nginx") {
		return
	}
	ui.RunWithSpinner("Установка Nginx...", func() error {
		runShellSilent("DEBIAN_FRONTEND=noninteractive apt-get install -y -qq nginx")
		runShellSilent("systemctl enable nginx")
		runShellSilent("systemctl start nginx")
		return nil
	})
}

func installCaddy() {
	if commandExists("caddy") {
		return
	}
	ui.RunWithSpinner("Установка Caddy...", func() error {
		runShellSilent("DEBIAN_FRONTEND=noninteractive apt-get install -y -qq debian-keyring debian-archive-keyring apt-transport-https curl")
		runShellSilent("curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' | gpg --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg 2>/dev/null || true")
		runShellSilent("curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' | tee /etc/apt/sources.list.d/caddy-stable.list > /dev/null")
		runShellSilent("DEBIAN_FRONTEND=noninteractive apt-get update -y -qq")
		runShellSilent("DEBIAN_FRONTEND=noninteractive apt-get install -y -qq caddy")
		return nil
	})
}
