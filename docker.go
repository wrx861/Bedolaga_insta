package main

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

// ════════════════════════════════════════════════════════════════
// DOCKER START
// ════════════════════════════════════════════════════════════════

func startDocker(cfg *Config) {
	runShellSilent(fmt.Sprintf("cd %s && docker compose down 2>/dev/null || true", cfg.InstallDir))
	runShellSilent(fmt.Sprintf("cd %s && docker compose -f docker-compose.local.yml down 2>/dev/null || true", cfg.InstallDir))

	composeFile := "docker-compose.yml"

	if cfg.PanelInstalledLocally {
		if cfg.DockerNetwork != "" {
			runShellSilent(fmt.Sprintf("docker network create %s 2>/dev/null || true", cfg.DockerNetwork))
		}
		createLocalCompose(cfg)
		composeFile = "docker-compose.local.yml"
	} else {
		if !fileExists(filepath.Join(cfg.InstallDir, "docker-compose.yml")) {
			createStandaloneCompose(cfg)
		}
	}

	runWithSpinner("Сборка и запуск контейнеров...", func() error {
		_, err := runShellSilent(fmt.Sprintf("cd %s && docker compose -f %s up -d --build 2>&1", cfg.InstallDir, composeFile))
		return err
	})

	printInfo("Ожидание контейнеров...")
	time.Sleep(8 * time.Second)

	// Show status
	out, _ := runShellSilent(fmt.Sprintf("cd %s && docker compose -f %s ps --format 'table {{.Name}}\\t{{.Status}}' 2>/dev/null", cfg.InstallDir, composeFile))
	if out != "" {
		fmt.Println()
		fmt.Println(dimStyle.Render("  " + strings.ReplaceAll(out, "\n", "\n  ")))
		fmt.Println()
	}

	if cfg.PanelInstalledLocally && cfg.DockerNetwork != "" {
		ensureNetworkConnection(cfg)
		verifyPanelConnection()
	}
}

func ensureNetworkConnection(cfg *Config) {
	net := cfg.DockerNetwork
	containers := []string{"remnawave_bot", "remnawave_bot_db", "remnawave_bot_redis"}
	for _, c := range containers {
		out, _ := runShellSilent(fmt.Sprintf("docker ps --format '{{.Names}}' | grep '^%s$'", c))
		if out == "" {
			continue
		}
		nets, _ := runShellSilent(fmt.Sprintf(`docker inspect %s --format '{{range $net, $_ := .NetworkSettings.Networks}}{{$net}} {{end}}'`, c))
		if !strings.Contains(nets, net) {
			runShellSilent(fmt.Sprintf("docker network connect %s %s 2>/dev/null", net, c))
		}
	}
}

func verifyPanelConnection() {
	time.Sleep(3 * time.Second)
	if out, err := runShellSilent("docker exec remnawave_bot getent hosts remnawave 2>/dev/null | awk '{print $1}'"); err == nil && out != "" {
		printSuccess("Подключение к панели проверено: remnawave -> " + out + ":3000")
	} else {
		printWarning("Не удаётся разрешить 'remnawave' — проверьте сетевое подключение вручную")
	}
}

// ════════════════════════════════════════════════════════════════
// FIREWALL (optional)
// ════════════════════════════════════════════════════════════════

func setupFirewall() {
	if !confirmPrompt("Настроить Firewall (UFW)?", false) {
		return
	}
	runWithSpinner("Настройка firewall...", func() error {
		if !commandExists("ufw") {
			runShellSilent("apt-get install -y ufw")
		}
		runShellSilent("ufw --force reset")
		runShellSilent("ufw default deny incoming")
		runShellSilent("ufw default allow outgoing")
		runShellSilent("ufw allow 22/tcp")
		runShellSilent("ufw allow 80/tcp")
		runShellSilent("ufw allow 443/tcp")
		runShellSilent("ufw --force enable")
		return nil
	})
}
