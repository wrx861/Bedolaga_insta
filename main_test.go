package main

import (
	"strings"
	"testing"
)

func TestGenerateToken(t *testing.T) {
	token := generateToken()
	if len(token) != 64 {
		t.Errorf("Expected 64 chars hex token, got %d", len(token))
	}
	token2 := generateToken()
	if token == token2 {
		t.Error("Two tokens should not be identical")
	}
}

func TestGenerateSafePassword(t *testing.T) {
	pw := generateSafePassword(24)
	if len(pw) != 24 {
		t.Errorf("Expected 24 chars password, got %d", len(pw))
	}
	for _, c := range pw {
		if !((c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9')) {
			t.Errorf("Password contains invalid char: %c", c)
		}
	}
}

func TestValidateDomain(t *testing.T) {
	tests := []struct {
		domain string
		valid  bool
	}{
		{"bot.example.com", true},
		{"example.com", true},
		{"sub.domain.example.com", true},
		{"notadomain", false},
		{"http://example.com", true},  // protocol is stripped internally
		{"https://example.com", true}, // protocol is stripped internally
		{".bad.com", false},
	}
	for _, tt := range tests {
		got := validateDomain(tt.domain)
		if got != tt.valid {
			t.Errorf("validateDomain(%q) = %v, want %v", tt.domain, got, tt.valid)
		}
	}
}

func TestCleanDomain(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"https://bot.example.com/", "bot.example.com"},
		{"http://bot.example.com", "bot.example.com"},
		{"bot.example.com", "bot.example.com"},
	}
	for _, tt := range tests {
		got := cleanDomain(tt.input)
		if got != tt.expected {
			t.Errorf("cleanDomain(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestFindInstallDir(t *testing.T) {
	dir := findInstallDir()
	_ = dir
}

func TestDetectComposeFile(t *testing.T) {
	result := detectComposeFile("/nonexistent")
	if result != "docker-compose.yml" {
		t.Errorf("Expected docker-compose.yml, got %s", result)
	}
}

func TestConfigDefaults(t *testing.T) {
	cfg := &Config{}
	if cfg.BotRunMode != "" {
		t.Error("Expected empty default")
	}
	if cfg.PanelInstalledLocally != false {
		t.Error("Expected false default")
	}
}

func TestAppVersion(t *testing.T) {
	if !strings.HasPrefix(appVersion, "2.") {
		t.Errorf("Expected version 2.x, got %s", appVersion)
	}
}

func TestFileExists(t *testing.T) {
	if !fileExists("/dev/null") {
		t.Error("Expected /dev/null to exist")
	}
	if fileExists("/nonexistent/path/abc") {
		t.Error("Expected nonexistent path to not exist")
	}
}

func TestDirExists(t *testing.T) {
	if !dirExists("/tmp") {
		t.Error("Expected /tmp to exist")
	}
	if dirExists("/nonexistent/dir") {
		t.Error("Expected nonexistent dir to not exist")
	}
}

func TestCommandExists(t *testing.T) {
	if !commandExists("bash") {
		t.Error("Expected bash to exist")
	}
	if commandExists("nonexistent_command_xyz") {
		t.Error("Expected nonexistent command to not exist")
	}
}
