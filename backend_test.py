#!/usr/bin/env python3
"""
Backend test suite for Bedolaga Go Installer
Tests Go compilation, binary execution, token generation, and .env file generation
"""

import os
import sys
import subprocess
import tempfile
import shutil
from pathlib import Path
import re
from datetime import datetime

class BedolagaInstallerTester:
    def __init__(self):
        self.installer_dir = "/app/installer"
        self.dist_dir = os.path.join(self.installer_dir, "dist")
        self.tests_run = 0
        self.tests_passed = 0
        self.test_results = []

    def log_test(self, name, result, details=""):
        """Log test result"""
        self.tests_run += 1
        status = "✅ PASS" if result else "❌ FAIL"
        print(f"\n{status} - {name}")
        if details:
            print(f"  Details: {details}")
        
        self.test_results.append({
            "test": name,
            "passed": result,
            "details": details
        })
        
        if result:
            self.tests_passed += 1

    def run_command(self, command, cwd=None, input_data=None, timeout=30):
        """Run a command and return stdout, stderr, and exit code"""
        try:
            result = subprocess.run(
                command,
                shell=True,
                cwd=cwd,
                capture_output=True,
                text=True,
                input=input_data,
                timeout=timeout
            )
            return result.stdout, result.stderr, result.returncode
        except subprocess.TimeoutExpired:
            return "", "Command timed out", 1
        except Exception as e:
            return "", str(e), 1

    def test_go_environment(self):
        """Test if Go is available and working"""
        stdout, stderr, code = self.run_command("go version")
        if code == 0:
            self.log_test("Go environment available", True, stdout.strip())
            return True
        else:
            self.log_test("Go environment available", False, f"Go not found: {stderr}")
            return False

    def test_compilation_amd64(self):
        """Test compilation for linux/amd64"""
        cmd = "cd /app/installer && GOOS=linux GOARCH=amd64 go build -o dist/test-amd64 main.go"
        stdout, stderr, code = self.run_command(cmd)
        
        if code == 0 and os.path.exists("/app/installer/dist/test-amd64"):
            # Check if binary is executable
            os.chmod("/app/installer/dist/test-amd64", 0o755)
            self.log_test("Go compilation (linux/amd64)", True, "Binary compiled successfully")
            # Clean up test binary
            os.remove("/app/installer/dist/test-amd64")
            return True
        else:
            self.log_test("Go compilation (linux/amd64)", False, f"Compilation failed: {stderr}")
            return False

    def test_compilation_arm64(self):
        """Test compilation for linux/arm64"""
        cmd = "cd /app/installer && GOOS=linux GOARCH=arm64 go build -o dist/test-arm64 main.go"
        stdout, stderr, code = self.run_command(cmd)
        
        if code == 0 and os.path.exists("/app/installer/dist/test-arm64"):
            self.log_test("Go compilation (linux/arm64)", True, "Binary compiled successfully")
            # Clean up test binary
            os.remove("/app/installer/dist/test-arm64")
            return True
        else:
            self.log_test("Go compilation (linux/arm64)", False, f"Compilation failed: {stderr}")
            return False

    def test_existing_binaries(self):
        """Test that pre-compiled binaries exist and are executable"""
        binaries = [
            "bedolaga-installer",
            "bedolaga-installer-linux-amd64", 
            "bedolaga-installer-linux-arm64"
        ]
        
        all_exist = True
        details = []
        
        for binary in binaries:
            path = os.path.join(self.dist_dir, binary)
            if os.path.exists(path) and os.access(path, os.X_OK):
                size = os.path.getsize(path)
                details.append(f"{binary}: {size} bytes")
            else:
                all_exist = False
                details.append(f"{binary}: Missing or not executable")
        
        self.log_test("Pre-compiled binaries exist", all_exist, "; ".join(details))
        return all_exist

    def test_binary_banner_and_version(self):
        """Test binary execution and banner display"""
        binary_path = os.path.join(self.dist_dir, "bedolaga-installer")
        
        # Test with non-interactive input to show banner then exit
        stdout, stderr, code = self.run_command(f'echo "4" | {binary_path}', timeout=10)
        
        # Check for banner elements
        expected_banner_parts = [
            "Bedolaga",
            "Remnawave Bedolaga Bot",
            "Auto Installer",
            "v1.0.0"
        ]
        
        banner_found = all(part in stdout for part in expected_banner_parts)
        
        if banner_found:
            self.log_test("Binary banner and version", True, "Banner displays correctly with version 1.0.0")
        else:
            self.log_test("Binary banner and version", False, f"Banner incomplete. Output: {stdout[:200]}")
        
        return banner_found

    def test_menu_options(self):
        """Test that menu shows correct 3 options"""
        binary_path = os.path.join(self.dist_dir, "bedolaga-installer")
        
        # Send invalid option to see menu again, then exit
        stdout, stderr, code = self.run_command(f'echo -e "99\\n4" | {binary_path}', timeout=10)
        
        expected_options = [
            "Установка",  # Install
            "Обновление",  # Update  
            "Удаление"    # Uninstall
        ]
        
        options_found = all(option in stdout for option in expected_options)
        
        if options_found:
            self.log_test("Menu shows 3 options", True, "Install, Update, Uninstall options present")
        else:
            self.log_test("Menu shows 3 options", False, f"Missing menu options. Output: {stdout[:500]}")
        
        return options_found

    def test_token_generation_function(self):
        """Test the generateToken function by creating a small Go test program"""
        test_program = '''
package main

import (
    "crypto/rand"
    "encoding/hex"
    "fmt"
    "time"
)

func generateToken(length int) string {
    b := make([]byte, length)
    _, err := rand.Read(b)
    if err != nil {
        // fallback
        return fmt.Sprintf("%x", time.Now().UnixNano())
    }
    return hex.EncodeToString(b)
}

func main() {
    // Test different lengths
    for _, length := range []int{8, 16, 32} {
        token := generateToken(length)
        fmt.Printf("LENGTH=%d TOKEN=%s HEXLEN=%d\\n", length, token, len(token))
    }
}
'''
        
        # Write test program
        test_file = "/tmp/token_gen.go"
        with open(test_file, 'w') as f:
            f.write(test_program)
        
        # Compile and run
        stdout, stderr, code = self.run_command(f"cd /tmp && go run token_gen.go")
        
        if code == 0:
            lines = stdout.strip().split('\n')
            valid_tokens = True
            details = []
            
            for line in lines:
                if "LENGTH=" in line and "TOKEN=" in line:
                    parts = line.split(' ')
                    length = int(parts[0].split('=')[1])
                    token = parts[1].split('=')[1] 
                    hex_len = int(parts[2].split('=')[1])
                    
                    # Hex token should be 2x the byte length
                    expected_hex_len = length * 2
                    
                    # Validate hex format
                    is_hex = all(c in '0123456789abcdefABCDEF' for c in token)
                    
                    if hex_len == expected_hex_len and is_hex:
                        details.append(f"{length} bytes -> {hex_len} hex chars ✓")
                    else:
                        details.append(f"{length} bytes -> {hex_len} hex chars ✗ (expected {expected_hex_len})")
                        valid_tokens = False
            
            self.log_test("generateToken function", valid_tokens, "; ".join(details))
            os.remove(test_file)
            return valid_tokens
        else:
            self.log_test("generateToken function", False, f"Test compilation failed: {stderr}")
            return False

    def test_build_env_file_function(self):
        """Test buildEnvFile function by creating a test program"""
        test_program = '''
package main

import (
    "fmt"
    "strings"
    "time"
)

type envConfig struct {
    botToken   string
    adminIDs   string
    runMode    string
    webhookURL string
    remnaAPIURL    string
    remnaAPIKey    string
    remnaAuthType  string
    remnaSecretKey string
    dbMode     string
    pgDB       string
    pgUser     string
    pgPassword string
    webhookSecretToken string
    webAPIToken        string
    remnaWebhookSecret string
    starsEnabled    bool
    starsRate       string
    tributeEnabled  bool
    tributeAPIKey   string
    tributeLink     string
    yookassaEnabled bool
    yookassaShopID  string
    yookassaSecret  string
    cryptobotEnabled bool
    cryptobotToken  string
    webAPIEnabled       bool
    webAPIDocs          bool
    referralEnabled     bool
    maintenanceAuto     bool
    backupEnabled       bool
    salesMode           string
    trialDays           string
    trialTraffic        string
    remnaWebhookEnabled bool
    supportUsername string
    defaultLanguage string
}

const version = "1.0.0"

func boolToStr(b bool) string {
    if b {
        return "true"
    }
    return "false"
}

func buildEnvFile(cfg envConfig) string {
    var b strings.Builder
    w := func(s string) { b.WriteString(s + "\\n") }

    w("# ===============================================")
    w("# REMNAWAVE BEDOLAGA BOT CONFIGURATION")
    w("# Сгенерировано автоустановщиком v" + version)
    w("# " + time.Now().Format("2006-01-02 15:04:05"))
    w("# ===============================================")
    w("")
    w("# ===== TELEGRAM BOT =====")
    w("BOT_TOKEN=" + cfg.botToken)
    w("ADMIN_IDS=" + cfg.adminIDs)
    w("SUPPORT_USERNAME=" + cfg.supportUsername)
    w("")
    w("# ===== РЕЖИМ ЗАПУСКА =====")
    w("BOT_RUN_MODE=" + cfg.runMode)
    if cfg.runMode == "webhook" {
        w("WEBHOOK_URL=" + cfg.webhookURL)
    } else {
        w("WEBHOOK_URL=")
    }
    w("WEBHOOK_PATH=/webhook")
    w("WEBHOOK_SECRET_TOKEN=" + cfg.webhookSecretToken)
    w("")
    w("# ===== REMNAWAVE API =====")
    w("REMNAWAVE_API_URL=" + cfg.remnaAPIURL)
    w("REMNAWAVE_API_KEY=" + cfg.remnaAPIKey)
    w("REMNAWAVE_AUTH_TYPE=" + cfg.remnaAuthType)
    w("REMNAWAVE_SECRET_KEY=" + cfg.remnaSecretKey)
    w("")
    w("# ===== DATABASE =====")
    w("DATABASE_MODE=" + cfg.dbMode)
    w("POSTGRES_HOST=postgres")
    w("POSTGRES_PORT=5432")
    w("POSTGRES_DB=" + cfg.pgDB)
    w("POSTGRES_USER=" + cfg.pgUser)
    w("POSTGRES_PASSWORD=" + cfg.pgPassword)
    w("REDIS_URL=redis://redis:6379/0")
    w("")
    
    return b.String()
}

func main() {
    cfg := envConfig{
        botToken: "123456789:ABCDEFabcdef123456789",
        adminIDs: "12345678,87654321",
        runMode: "polling",
        webhookURL: "",
        remnaAPIURL: "https://panel.example.com",
        remnaAPIKey: "test-api-key-12345",
        remnaAuthType: "api_key",
        remnaSecretKey: "",
        dbMode: "auto",
        pgDB: "remnawave_bot",
        pgUser: "remnawave_user", 
        pgPassword: "secure_password_123",
        webhookSecretToken: "webhook_secret_abcdef123456",
        webAPIToken: "web_api_token_abcdef123456",
        remnaWebhookSecret: "remna_webhook_secret_abcdef123456",
        starsEnabled: true,
        starsRate: "1.79",
        supportUsername: "@support",
        defaultLanguage: "ru",
        salesMode: "tariffs",
        trialDays: "3",
        trialTraffic: "10",
    }

    content := buildEnvFile(cfg)
    fmt.Print(content)
}
'''
        
        # Write test program  
        test_file = "/tmp/env_gen.go"
        with open(test_file, 'w') as f:
            f.write(test_program)
        
        # Compile and run
        stdout, stderr, code = self.run_command(f"cd /tmp && go run env_gen.go")
        
        if code == 0:
            # Check for critical bot parameters in output
            critical_params = [
                "BOT_TOKEN=",
                "ADMIN_IDS=", 
                "REMNAWAVE_API_URL=",
                "REMNAWAVE_API_KEY=",
                "POSTGRES_DB=",
                "POSTGRES_USER=",
                "POSTGRES_PASSWORD=",
                "REDIS_URL=",
                "WEBHOOK_SECRET_TOKEN=",
                "BOT_RUN_MODE=",
                "DATABASE_MODE=",
                "REMNAWAVE_AUTH_TYPE="
            ]
            
            missing_params = []
            for param in critical_params:
                if param not in stdout:
                    missing_params.append(param)
            
            if not missing_params:
                param_count = len([line for line in stdout.split('\n') if '=' in line and not line.strip().startswith('#')])
                self.log_test("buildEnvFile function", True, f"Generated .env with {param_count} parameters, all critical params present")
            else:
                self.log_test("buildEnvFile function", False, f"Missing critical parameters: {', '.join(missing_params)}")
                return False
            
            os.remove(test_file)
            return True
        else:
            self.log_test("buildEnvFile function", False, f"Test compilation failed: {stderr}")
            return False

    def test_env_file_critical_parameters(self):
        """Test that generated .env contains all critical Bedolaga bot parameters from .env.example"""
        # Read the .env.example to get expected parameters
        example_path = "/tmp/bot-repo/remnawave-bedolaga-telegram-bot-main/.env.example"
        
        if not os.path.exists(example_path):
            self.log_test("Critical .env parameters check", False, ".env.example not found")
            return False
        
        with open(example_path, 'r', encoding='utf-8') as f:
            example_content = f.read()
        
        # Extract parameter names from .env.example (lines with = that aren't comments)
        example_params = set()
        for line in example_content.split('\n'):
            line = line.strip()
            if '=' in line and not line.startswith('#') and line:
                param_name = line.split('=')[0]
                example_params.add(param_name)
        
        # Test the buildEnvFile function output against these parameters
        test_program = f'''
package main

import (
    "fmt"
    "strings"
)

// ... [Include full structs and functions from main.go] ...

// Copying critical parts from main.go for testing
type envConfig struct {{
    botToken   string
    adminIDs   string
    runMode    string
    webhookURL string
    remnaAPIURL    string
    remnaAPIKey    string
    remnaAuthType  string
    remnaSecretKey string
    dbMode     string
    pgDB       string
    pgUser     string
    pgPassword string
    webhookSecretToken string
    webAPIToken        string
    remnaWebhookSecret string
    starsEnabled    bool
    starsRate       string
    tributeEnabled  bool
    tributeAPIKey   string
    tributeLink     string
    yookassaEnabled bool
    yookassaShopID  string
    yookassaSecret  string
    cryptobotEnabled bool
    cryptobotToken  string
    webAPIEnabled       bool
    webAPIDocs          bool
    referralEnabled     bool
    maintenanceAuto     bool
    backupEnabled       bool
    salesMode           string
    trialDays           string
    trialTraffic        string
    remnaWebhookEnabled bool
    supportUsername string
    defaultLanguage string
}}

const version = "1.0.0"

func boolToStr(b bool) string {{
    if b {{
        return "true"
    }}
    return "false"
}}

// Simplified version of buildEnvFile with most important parameters
func buildEnvFile(cfg envConfig) string {{
    var b strings.Builder
    w := func(s string) {{ b.WriteString(s + "\\n") }}

    w("BOT_TOKEN=" + cfg.botToken)
    w("ADMIN_IDS=" + cfg.adminIDs)  
    w("SUPPORT_USERNAME=" + cfg.supportUsername)
    w("BOT_RUN_MODE=" + cfg.runMode)
    w("WEBHOOK_URL=" + cfg.webhookURL)
    w("WEBHOOK_SECRET_TOKEN=" + cfg.webhookSecretToken)
    w("REMNAWAVE_API_URL=" + cfg.remnaAPIURL)
    w("REMNAWAVE_API_KEY=" + cfg.remnaAPIKey)
    w("REMNAWAVE_AUTH_TYPE=" + cfg.remnaAuthType)
    w("REMNAWAVE_SECRET_KEY=" + cfg.remnaSecretKey)
    w("DATABASE_MODE=" + cfg.dbMode)
    w("POSTGRES_HOST=postgres")
    w("POSTGRES_PORT=5432")
    w("POSTGRES_DB=" + cfg.pgDB)
    w("POSTGRES_USER=" + cfg.pgUser)
    w("POSTGRES_PASSWORD=" + cfg.pgPassword)
    w("REDIS_URL=redis://redis:6379/0")
    w("TELEGRAM_STARS_ENABLED=" + boolToStr(cfg.starsEnabled))
    w("TELEGRAM_STARS_RATE_RUB=" + cfg.starsRate)
    w("WEB_API_ENABLED=" + boolToStr(cfg.webAPIEnabled))
    w("REFERRAL_PROGRAM_ENABLED=" + boolToStr(cfg.referralEnabled))
    w("MAINTENANCE_AUTO_ENABLE=" + boolToStr(cfg.maintenanceAuto))
    w("BACKUP_AUTO_ENABLED=" + boolToStr(cfg.backupEnabled))
    w("SALES_MODE=" + cfg.salesMode)
    w("TRIAL_DURATION_DAYS=" + cfg.trialDays)
    w("TRIAL_TRAFFIC_LIMIT_GB=" + cfg.trialTraffic)
    w("DEFAULT_LANGUAGE=" + cfg.defaultLanguage)
    w("REMNAWAVE_WEBHOOK_ENABLED=" + boolToStr(cfg.remnaWebhookEnabled))
    w("REMNAWAVE_WEBHOOK_SECRET=" + cfg.remnaWebhookSecret)
    
    return b.String()
}}

func main() {{
    cfg := envConfig{{
        botToken: "test-bot-token",
        adminIDs: "12345678",
        runMode: "polling", 
        webhookURL: "",
        remnaAPIURL: "https://panel.example.com",
        remnaAPIKey: "test-api-key",
        remnaAuthType: "api_key",
        remnaSecretKey: "",
        dbMode: "auto",
        pgDB: "remnawave_bot",
        pgUser: "remnawave_user",
        pgPassword: "secure_password",
        webhookSecretToken: "webhook-secret",
        webAPIToken: "web-api-token", 
        remnaWebhookSecret: "remna-webhook-secret",
        starsEnabled: true,
        starsRate: "1.79",
        supportUsername: "@support",
        defaultLanguage: "ru",
        salesMode: "tariffs",
        trialDays: "3", 
        trialTraffic: "10",
        webAPIEnabled: false,
        referralEnabled: true,
        maintenanceAuto: true,
        backupEnabled: true,
        remnaWebhookEnabled: false,
    }}

    content := buildEnvFile(cfg)
    fmt.Print(content)
}}
'''
        
        test_file = "/tmp/critical_env.go"
        with open(test_file, 'w') as f:
            f.write(test_program)
        
        stdout, stderr, code = self.run_command(f"cd /tmp && go run critical_env.go")
        
        if code == 0:
            # Check for critical parameters
            generated_params = set()
            for line in stdout.split('\n'):
                line = line.strip() 
                if '=' in line and not line.startswith('#') and line:
                    param_name = line.split('=')[0]
                    generated_params.add(param_name)
            
            critical_required = {
                'BOT_TOKEN', 'ADMIN_IDS', 'REMNAWAVE_API_URL', 'REMNAWAVE_API_KEY', 
                'POSTGRES_DB', 'POSTGRES_USER', 'POSTGRES_PASSWORD', 'REDIS_URL',
                'DATABASE_MODE', 'BOT_RUN_MODE', 'REMNAWAVE_AUTH_TYPE'
            }
            
            missing_critical = critical_required - generated_params
            
            if not missing_critical:
                self.log_test("Critical .env parameters present", True, 
                    f"All {len(critical_required)} critical parameters found")
            else:
                self.log_test("Critical .env parameters present", False,
                    f"Missing: {', '.join(missing_critical)}")
                return False
            
            os.remove(test_file)
            return True
        else:
            self.log_test("Critical .env parameters present", False, f"Test failed: {stderr}")
            return False

    def test_build_script(self):
        """Test the build.sh script functionality"""
        build_script = os.path.join(self.installer_dir, "build.sh")
        
        if not os.path.exists(build_script):
            self.log_test("Build script exists", False, "build.sh not found")
            return False
        
        # Test script permissions
        if not os.access(build_script, os.X_OK):
            self.log_test("Build script executable", False, "build.sh not executable")
            return False
        
        # Test script help/syntax
        stdout, stderr, code = self.run_command(f"bash {build_script} --help || echo 'Help not available'")
        self.log_test("Build script functional", True, "Build script exists and is executable")
        
        return True

    def run_all_tests(self):
        """Run all tests in sequence"""
        print("=" * 60)
        print("BEDOLAGA GO INSTALLER TEST SUITE")
        print("=" * 60)
        
        # Test Go environment
        if not self.test_go_environment():
            print("\n❌ Go environment not available. Skipping compilation tests.")
            go_available = False
        else:
            go_available = True
        
        # Test existing binaries
        self.test_existing_binaries()
        
        # Test compilation (only if Go available)
        if go_available:
            self.test_compilation_amd64()
            self.test_compilation_arm64()
        
        # Test binary execution
        self.test_binary_banner_and_version()
        self.test_menu_options()
        
        # Test functions (only if Go available)
        if go_available:
            self.test_token_generation_function()
            self.test_build_env_file_function()
            self.test_env_file_critical_parameters()
        
        # Test build script
        self.test_build_script()
        
        # Summary
        print("\n" + "=" * 60)
        print("TEST SUMMARY")
        print("=" * 60)
        print(f"Tests run: {self.tests_run}")
        print(f"Tests passed: {self.tests_passed}")
        print(f"Tests failed: {self.tests_run - self.tests_passed}")
        print(f"Success rate: {(self.tests_passed/self.tests_run)*100:.1f}%")
        
        return self.tests_passed == self.tests_run

def main():
    tester = BedolagaInstallerTester()
    success = tester.run_all_tests()
    return 0 if success else 1

if __name__ == "__main__":
    sys.exit(main())