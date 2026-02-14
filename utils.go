package main

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"bedolaga-installer/pkg/ui"
)

// ════════════════════════════════════════════════════════════════
// SYSTEM UTILS
// ════════════════════════════════════════════════════════════════

var reader = bufio.NewReader(os.Stdin)

func runCmd(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = "/root"
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

func runCmdSilent(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	cmd.Dir = "/root"
	out, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

func runShell(command string) error {
	cmd := exec.Command("bash", "-c", command)
	cmd.Dir = "/root"
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

func runShellSilent(command string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, "bash", "-c", command)
	cmd.Dir = "/root"
	cmd.Env = append(os.Environ(), "DEBIAN_FRONTEND=noninteractive")
	out, err := cmd.CombinedOutput()
	if ctx.Err() == context.DeadlineExceeded {
		return "", fmt.Errorf("команда превысила таймаут: %s", command)
	}
	return strings.TrimSpace(string(out)), err
}

func commandExists(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

func generateToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func generateSafePassword(length int) string {
	const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, length)
	rand.Read(b)
	for i := range b {
		b[i] = chars[int(b[i])%len(chars)]
	}
	return string(b)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func validateDomain(domain string) bool {
	domain = strings.TrimPrefix(domain, "https://")
	domain = strings.TrimPrefix(domain, "http://")
	domain = strings.TrimSuffix(domain, "/")
	if !strings.Contains(domain, ".") {
		return false
	}
	re := regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9.\-]+[a-zA-Z0-9]$`)
	return re.MatchString(domain)
}

func cleanDomain(input string) string {
	d := strings.TrimPrefix(input, "https://")
	d = strings.TrimPrefix(d, "http://")
	d = strings.TrimSuffix(d, "/")
	return d
}

func checkDomainDNS(domain string) bool {
	serverIP, err := runShellSilent("curl -4 -s --connect-timeout 5 ifconfig.me 2>/dev/null || curl -4 -s --connect-timeout 5 icanhazip.com 2>/dev/null")
	if err != nil || serverIP == "" {
		ui.PrintWarning("Не удалось определить IP сервера")
		return false
	}
	ips, err := net.LookupHost(domain)
	if err != nil || len(ips) == 0 {
		ui.PrintWarning(fmt.Sprintf("DNS-запись для %s не найдена", domain))
		return false
	}
	for _, ip := range ips {
		if ip == strings.TrimSpace(serverIP) {
			ui.PrintSuccess(fmt.Sprintf("DNS %s → %s", domain, ip))
			return true
		}
	}
	ui.PrintWarning(fmt.Sprintf("Домен %s → %s, IP сервера %s", domain, ips[0], serverIP))
	return false
}
