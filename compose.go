package main

import (
	"fmt"
	"os"
	"path/filepath"
)

// ════════════════════════════════════════════════════════════════
// CLONE & DIRECTORIES
// ════════════════════════════════════════════════════════════════

func cloneRepository(cfg *Config) {
	if dirExists(cfg.InstallDir) {
		// Обновляем существующий
		runShellSilent(fmt.Sprintf("cd %s && git pull origin main 2>/dev/null || true", cfg.InstallDir))
		globalProgress.done("Репозиторий обновлён")
		return
	}

	// Клонируем новый
	_, err := runCmdSilent("git", "clone", repoURL, cfg.InstallDir)
	if err != nil {
		globalProgress.fail("Ошибка клонирования: " + err.Error())
		os.Exit(1)
	}
	globalProgress.done("Репозиторий клонирован")
}

func createDirectories(cfg *Config) {
	dirs := []string{"logs", "data", "data/backups", "data/referral_qr", "locales"}
	for _, d := range dirs {
		os.MkdirAll(filepath.Join(cfg.InstallDir, d), 0777)
	}
	// Даём полные права чтобы Docker контейнер мог писать
	runShellSilent(fmt.Sprintf("chmod -R 777 %s/logs %s/data 2>/dev/null || true", cfg.InstallDir, cfg.InstallDir))
	runShellSilent(fmt.Sprintf("chmod -R 755 %s/locales 2>/dev/null || true", cfg.InstallDir))
}

// ════════════════════════════════════════════════════════════════
// DOCKER COMPOSE FILES
// ════════════════════════════════════════════════════════════════

func createStandaloneCompose(cfg *Config) {
	content := `services:
  postgres:
    image: postgres:15-alpine
    container_name: remnawave_bot_db
    restart: unless-stopped
    environment:
      POSTGRES_DB: ${POSTGRES_DB:-remnawave_bot}
      POSTGRES_USER: ${POSTGRES_USER:-remnawave_user}
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD:-secure_password_123}
      POSTGRES_INITDB_ARGS: "--encoding=UTF8 --locale=C"
    volumes:
      - postgres_data:/var/lib/postgresql/data
    networks:
      - bot_network
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U ${POSTGRES_USER:-remnawave_user} -d ${POSTGRES_DB:-remnawave_bot}"]
      interval: 30s
      timeout: 5s
      retries: 5
      start_period: 30s

  redis:
    image: redis:7-alpine
    container_name: remnawave_bot_redis
    restart: unless-stopped
    command: redis-server --appendonly yes --maxmemory 256mb --maxmemory-policy allkeys-lru
    volumes:
      - redis_data:/data
    networks:
      - bot_network
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 30s
      timeout: 10s
      retries: 3

  bot:
    build: .
    container_name: remnawave_bot
    restart: unless-stopped
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
    env_file:
      - .env
    environment:
      DOCKER_ENV: "true"
      DATABASE_MODE: "auto"
      POSTGRES_HOST: "postgres"
      POSTGRES_PORT: "5432"
      POSTGRES_DB: "${POSTGRES_DB:-remnawave_bot}"
      POSTGRES_USER: "${POSTGRES_USER:-remnawave_user}"
      POSTGRES_PASSWORD: "${POSTGRES_PASSWORD:-secure_password_123}"
      REDIS_URL: "redis://redis:6379/0"
      TZ: "Europe/Moscow"
      LOCALES_PATH: "${LOCALES_PATH:-/app/locales}"
    volumes:
      - ./logs:/app/logs:rw
      - ./data:/app/data:rw
      - ./locales:/app/locales:rw
      - /etc/timezone:/etc/timezone:ro
      - /etc/localtime:/etc/localtime:ro
      - ./vpn_logo.png:/app/vpn_logo.png:ro
    ports:
      - "${WEB_API_PORT:-8080}:8080"
    networks:
      - bot_network
    healthcheck:
      test: ["CMD-SHELL", "wget -q --spider http://localhost:8080/health || exit 1"]
      interval: 60s
      timeout: 10s
      retries: 3
      start_period: 30s

volumes:
  postgres_data:
    driver: local
  redis_data:
    driver: local

networks:
  bot_network:
    name: remnawave_bot_network
    driver: bridge
`
	os.WriteFile(filepath.Join(cfg.InstallDir, "docker-compose.yml"), []byte(content), 0644)
}

func createLocalCompose(cfg *Config) {
	networkName := cfg.DockerNetwork
	if networkName == "" {
		networkName = "remnawave-network"
	}
	content := fmt.Sprintf(`services:
  postgres:
    image: postgres:15-alpine
    container_name: remnawave_bot_db
    restart: unless-stopped
    environment:
      POSTGRES_DB: ${POSTGRES_DB:-remnawave_bot}
      POSTGRES_USER: ${POSTGRES_USER:-remnawave_user}
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD:-secure_password_123}
      POSTGRES_INITDB_ARGS: "--encoding=UTF8 --locale=C"
    volumes:
      - postgres_data:/var/lib/postgresql/data
    networks:
      - bot_network
      - remnawave_network
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U ${POSTGRES_USER:-remnawave_user} -d ${POSTGRES_DB:-remnawave_bot}"]
      interval: 30s
      timeout: 5s
      retries: 5
      start_period: 30s

  redis:
    image: redis:7-alpine
    container_name: remnawave_bot_redis
    restart: unless-stopped
    command: redis-server --appendonly yes --maxmemory 256mb --maxmemory-policy allkeys-lru
    volumes:
      - redis_data:/data
    networks:
      - bot_network
      - remnawave_network
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 30s
      timeout: 10s
      retries: 3

  bot:
    build: .
    container_name: remnawave_bot
    restart: unless-stopped
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
    env_file:
      - .env
    environment:
      DOCKER_ENV: "true"
      DATABASE_MODE: "auto"
      POSTGRES_HOST: "postgres"
      POSTGRES_PORT: "5432"
      POSTGRES_DB: "${POSTGRES_DB:-remnawave_bot}"
      POSTGRES_USER: "${POSTGRES_USER:-remnawave_user}"
      POSTGRES_PASSWORD: "${POSTGRES_PASSWORD:-secure_password_123}"
      REDIS_URL: "redis://redis:6379/0"
      TZ: "Europe/Moscow"
      LOCALES_PATH: "${LOCALES_PATH:-/app/locales}"
    volumes:
      - ./logs:/app/logs:rw
      - ./data:/app/data:rw
      - ./locales:/app/locales:rw
      - /etc/timezone:/etc/timezone:ro
      - /etc/localtime:/etc/localtime:ro
      - ./vpn_logo.png:/app/vpn_logo.png:ro
    ports:
      - "${WEB_API_PORT:-8080}:8080"
    networks:
      - bot_network
      - remnawave_network
    healthcheck:
      test: ["CMD-SHELL", "wget -q --spider http://localhost:8080/health || exit 1"]
      interval: 60s
      timeout: 10s
      retries: 3
      start_period: 30s

volumes:
  postgres_data:
    driver: local
  redis_data:
    driver: local

networks:
  bot_network:
    name: remnawave_bot_network
    driver: bridge
  remnawave_network:
    name: %s
    external: true
`, networkName)
	os.WriteFile(filepath.Join(cfg.InstallDir, "docker-compose.local.yml"), []byte(content), 0644)
}
