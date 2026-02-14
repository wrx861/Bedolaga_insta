package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"bedolaga-installer/pkg/ui"
)

// ════════════════════════════════════════════════════════════════
// NGINX SETUP
// ════════════════════════════════════════════════════════════════

func setupNginxSystem(cfg *Config) {
	installNginx()

	nginxAvail := "/etc/nginx/sites-available"
	nginxEnabled := "/etc/nginx/sites-enabled"
	os.MkdirAll(nginxAvail, 0755)
	os.MkdirAll(nginxEnabled, 0755)

	if cfg.WebhookDomain != "" {
		conf := fmt.Sprintf(`server {
    listen 80;
    server_name %s;
    client_max_body_size 32m;

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_read_timeout 120s;
        proxy_send_timeout 120s;
        proxy_buffering off;
        proxy_request_buffering off;
    }

    location = /app-config.json {
        add_header Access-Control-Allow-Origin "*";
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
`, cfg.WebhookDomain)
		os.WriteFile(filepath.Join(nginxAvail, "bedolaga-webhook"), []byte(conf), 0644)
		os.Remove(filepath.Join(nginxEnabled, "bedolaga-webhook"))
		os.Symlink(filepath.Join(nginxAvail, "bedolaga-webhook"), filepath.Join(nginxEnabled, "bedolaga-webhook"))
	}

	if cfg.MiniappDomain != "" {
		conf := fmt.Sprintf(`server {
    listen 80;
    server_name %s;
    client_max_body_size 32m;
    root %s/miniapp;
    index index.html;

    location / {
        try_files $uri $uri/ /index.html;
        expires 1h;
        add_header Cache-Control "public";
    }

    location /miniapp/ {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_read_timeout 120s;
        proxy_send_timeout 120s;
    }

    location = /app-config.json {
        add_header Access-Control-Allow-Origin "*";
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
`, cfg.MiniappDomain, cfg.InstallDir)
		os.WriteFile(filepath.Join(nginxAvail, "bedolaga-miniapp"), []byte(conf), 0644)
		os.Remove(filepath.Join(nginxEnabled, "bedolaga-miniapp"))
		os.Symlink(filepath.Join(nginxAvail, "bedolaga-miniapp"), filepath.Join(nginxEnabled, "bedolaga-miniapp"))
	}

	runShellSilent("nginx -t && systemctl reload nginx")
	ui.PrintSuccess("Nginx настроен")
}

func setupNginxPanel(cfg *Config) {
	panelNginxConf := filepath.Join(cfg.PanelDir, "nginx.conf")
	if !fileExists(panelNginxConf) {
		ui.PrintWarning("nginx.conf панели не найден, переключаемся на системный nginx")
		setupNginxSystem(cfg)
		return
	}

	runShellSilent(fmt.Sprintf(`cp "%s" "%s.backup.$(date +%%Y%%m%%d_%%H%%M%%S)"`, panelNginxConf, panelNginxConf))
	runShellSilent(fmt.Sprintf(`sed -i '/# === BEGIN Bedolaga Bot ===/,/# === END Bedolaga Bot ===/d' "%s"`, panelNginxConf))

	block := "\n# === BEGIN Bedolaga Bot ===\n"
	if cfg.WebhookDomain != "" {
		block += fmt.Sprintf(`server {
    server_name %s;
    listen 443 ssl;
    http2 on;
    ssl_certificate "/etc/letsencrypt/live/%s/fullchain.pem";
    ssl_certificate_key "/etc/letsencrypt/live/%s/privkey.pem";
    client_max_body_size 32m;
    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection $connection_upgrade;
        proxy_read_timeout 120s;
        proxy_send_timeout 120s;
        proxy_buffering off;
    }
}
`, cfg.WebhookDomain, cfg.WebhookDomain, cfg.WebhookDomain)
	}
	if cfg.MiniappDomain != "" {
		block += fmt.Sprintf(`server {
    server_name %s;
    listen 443 ssl;
    http2 on;
    ssl_certificate "/etc/letsencrypt/live/%s/fullchain.pem";
    ssl_certificate_key "/etc/letsencrypt/live/%s/privkey.pem";
    client_max_body_size 32m;
    location /miniapp/ {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_read_timeout 120s;
        proxy_send_timeout 120s;
    }
    location = /app-config.json {
        add_header Access-Control-Allow-Origin "*";
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
    location / {
        root /var/www/remnawave-miniapp;
        try_files $uri $uri/ /index.html;
        expires 1h;
        add_header Cache-Control "public, immutable";
    }
}
`, cfg.MiniappDomain, cfg.MiniappDomain, cfg.MiniappDomain)
	}
	block += "# === END Bedolaga Bot ===\n"

	f, err := os.OpenFile(panelNginxConf, os.O_APPEND|os.O_WRONLY, 0644)
	if err == nil {
		f.WriteString(block)
		f.Close()
	}

	runShellSilent(fmt.Sprintf("cd %s && docker compose up -d remnawave-nginx 2>/dev/null || docker restart remnawave-nginx 2>/dev/null || true", cfg.PanelDir))
	ui.PrintSuccess("Nginx панели обновлён")
}

// ════════════════════════════════════════════════════════════════
// CADDY SETUP (Docker-based)
// ════════════════════════════════════════════════════════════════

func setupCaddy(cfg *Config) {
	// Останавливаем nginx/apache если запущены (они занимают порт 80)
	runShellSilent("systemctl stop nginx 2>/dev/null || true")
	runShellSilent("systemctl disable nginx 2>/dev/null || true")
	runShellSilent("systemctl stop apache2 2>/dev/null || true")
	runShellSilent("systemctl disable apache2 2>/dev/null || true")

	// Создаём Caddyfile для Docker-контейнера
	createCaddyfile(cfg)
	// Создаём docker-compose для Caddy
	createCaddyCompose(cfg)
	ui.PrintSuccess("Caddy настроен (Docker-контейнер с автоматическим HTTPS)")
}

func createCaddyfile(cfg *Config) {
	caddyDir := filepath.Join(cfg.InstallDir, "caddy")
	os.MkdirAll(caddyDir, 0755)

	var content string

	// Используем 127.0.0.1:8080 т.к. Caddy работает в host network mode
	if cfg.WebhookDomain != "" {
		content += fmt.Sprintf(`%s {
    reverse_proxy 127.0.0.1:8080 {
        flush_interval -1
    }
}

`, cfg.WebhookDomain)
	}

	if cfg.MiniappDomain != "" {
		content += fmt.Sprintf(`%s {
    @api path /miniapp/*
    reverse_proxy @api 127.0.0.1:8080 {
        flush_interval -1
    }
    @config path /app-config.json
    reverse_proxy @config 127.0.0.1:8080
    header @config Access-Control-Allow-Origin *
    root * %s/miniapp
    try_files {path} {path}/ /index.html
    file_server
}

`, cfg.MiniappDomain, cfg.InstallDir)
	}

	os.WriteFile(filepath.Join(caddyDir, "Caddyfile"), []byte(content), 0644)
}

func createCaddyCompose(cfg *Config) {
	// Caddy использует host network mode для доступа к интернету (Let's Encrypt)
	// и к боту на 127.0.0.1:8080
	miniappVolume := ""
	if cfg.MiniappDomain != "" {
		miniappVolume = fmt.Sprintf(`
      - %s/miniapp:/srv/miniapp:ro`, cfg.InstallDir)
	}

	content := fmt.Sprintf(`services:
  caddy:
    image: caddy:2-alpine
    container_name: remnawave_caddy
    restart: unless-stopped
    network_mode: host
    volumes:
      - ./caddy/Caddyfile:/etc/caddy/Caddyfile:ro
      - caddy_data:/data
      - caddy_config:/config%s

volumes:
  caddy_data:
    driver: local
  caddy_config:
    driver: local
`, miniappVolume)

	os.WriteFile(filepath.Join(cfg.InstallDir, "docker-compose.caddy.yml"), []byte(content), 0644)
}

// ════════════════════════════════════════════════════════════════
// SSL SETUP
// ════════════════════════════════════════════════════════════════

func setupSSL(cfg *Config) {
	if cfg.ReverseProxyType == "caddy" || cfg.ReverseProxyType == "skip" {
		return
	}
	if cfg.WebhookDomain == "" && cfg.MiniappDomain == "" {
		return
	}

	if !ui.ConfirmPrompt("Получить SSL-сертификаты сейчас?", true) {
		ui.PrintInfo("Вы можете получить сертификаты позже: certbot --nginx -d yourdomain.com")
		return
	}

	cfg.SSLEmail = ui.InputText("Email Let's Encrypt", "admin@example.com", "Email для уведомлений о SSL-сертификатах", true)

	isPanelMode := cfg.ReverseProxyType == "nginx_panel"

	for _, domain := range []string{cfg.WebhookDomain, cfg.MiniappDomain} {
		if domain == "" {
			continue
		}
		ui.RunWithSpinner("Получение SSL для "+domain+"...", func() error {
			if isPanelMode {
				runShellSilent("docker stop remnawave-nginx 2>/dev/null || true")
				runShellSilent("systemctl stop nginx 2>/dev/null || true")
				time.Sleep(2 * time.Second)
				err := runShell(fmt.Sprintf("certbot certonly --standalone -d %s --email %s --agree-tos --non-interactive", domain, cfg.SSLEmail))
				runShellSilent("docker start remnawave-nginx 2>/dev/null || true")
				runShellSilent("systemctl start nginx 2>/dev/null || true")
				return err
			}
			return runShell(fmt.Sprintf("certbot --nginx -d %s --email %s --agree-tos --non-interactive", domain, cfg.SSLEmail))
		})
	}

	runShellSilent("systemctl enable certbot.timer 2>/dev/null || true")
	runShellSilent("systemctl start certbot.timer 2>/dev/null || true")
}
