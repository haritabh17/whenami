package launchagent

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/haritabh17/theirtime/internal/config"
)

const (
	menubarLabel     = "dev.theirtime.menubar"
	legacyTickLabel  = "dev.whenami"
	legacyWatchLabel = "dev.whenami.watch"
	legacyMenubar    = "dev.whenami.menubar"
)

func menubarPlistPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "Library", "LaunchAgents", menubarLabel+".plist"), nil
}

func logDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, "Library", "Logs", "theirtime")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	return dir, nil
}

func Install(binaryPath string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	if err := removeLegacyAgents(); err != nil {
		return err
	}
	if len(cfg.Team) == 0 {
		return uninstallMenubarAgent()
	}
	logs, err := logDir()
	if err != nil {
		return err
	}
	return installMenubarAgent(binaryPath, logs)
}

func installMenubarAgent(binaryPath, logs string) error {
	menubarPath, err := menubarPlistPath()
	if err != nil {
		return err
	}
	menubarPlist := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>Label</key>
  <string>%s</string>
  <key>ProgramArguments</key>
  <array>
    <string>%s</string>
    <string>menubar</string>
  </array>
  <key>RunAtLoad</key>
  <true/>
  <key>KeepAlive</key>
  <true/>
  <key>LSAppNapIsDisabled</key>
  <true/>
  <key>StandardOutPath</key>
  <string>%s/menubar.log</string>
  <key>StandardErrorPath</key>
  <string>%s/menubar.error.log</string>
</dict>
</plist>
`, menubarLabel, binaryPath, logs, logs)

	if err := os.MkdirAll(filepath.Dir(menubarPath), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(menubarPath, []byte(menubarPlist), 0o644); err != nil {
		return err
	}
	return loadAgent(menubarPath, menubarLabel)
}

func uninstallMenubarAgent() error {
	menubarPath, err := menubarPlistPath()
	if err != nil {
		return err
	}
	_ = bootout(menubarPath)
	if err := os.Remove(menubarPath); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func removeLegacyAgents() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	agentsDir := filepath.Join(home, "Library", "LaunchAgents")
	legacy := []struct {
		label string
		name  string
	}{
		{legacyTickLabel, legacyTickLabel + ".plist"},
		{legacyWatchLabel, legacyWatchLabel + ".plist"},
		{legacyMenubar, legacyMenubar + ".plist"},
	}
	for _, l := range legacy {
		path := filepath.Join(agentsDir, l.name)
		_ = bootout(path)
		_ = bootoutByLabel(l.label)
		_ = os.Remove(path)
	}
	return nil
}

func loadAgent(plistPath, label string) error {
	_ = bootout(plistPath)

	uid := os.Getuid()
	domain := fmt.Sprintf("gui/%d", uid)
	out, err := exec.Command("launchctl", "bootstrap", domain, plistPath).CombinedOutput()
	if err != nil {
		msg := strings.TrimSpace(string(out))
		if strings.Contains(msg, "already") || strings.Contains(msg, "Input/output error") {
			_ = bootoutByLabel(label)
			out, err = exec.Command("launchctl", "bootstrap", domain, plistPath).CombinedOutput()
			if err != nil {
				return fmt.Errorf("launchctl bootstrap: %s: %w", strings.TrimSpace(string(out)), err)
			}
		} else {
			return fmt.Errorf("launchctl bootstrap: %s: %w", msg, err)
		}
	}
	return kickstart(label)
}

func bootoutByLabel(label string) error {
	uid := os.Getuid()
	target := fmt.Sprintf("gui/%d/%s", uid, label)
	out, err := exec.Command("launchctl", "bootout", target).CombinedOutput()
	if err != nil && !isBootoutBenign(string(out)) {
		return fmt.Errorf("launchctl bootout: %s: %w", strings.TrimSpace(string(out)), err)
	}
	return nil
}

func isBootoutBenign(out string) bool {
	return strings.Contains(out, "No such process") ||
		strings.Contains(out, "Could not find service") ||
		strings.Contains(out, "not found")
}

func Uninstall() error {
	if err := removeLegacyAgents(); err != nil {
		return err
	}
	return uninstallMenubarAgent()
}

func IsMenubarInstalled() bool {
	path, err := menubarPlistPath()
	if err != nil {
		return false
	}
	_, err = os.Stat(path)
	return err == nil
}

func bootout(plistPath string) error {
	uid := os.Getuid()
	domain := fmt.Sprintf("gui/%d", uid)
	out, err := exec.Command("launchctl", "bootout", domain, plistPath).CombinedOutput()
	if err != nil && !isBootoutBenign(string(out)) {
		return fmt.Errorf("launchctl bootout: %s: %w", strings.TrimSpace(string(out)), err)
	}
	return nil
}

func kickstart(label string) error {
	uid := os.Getuid()
	target := fmt.Sprintf("gui/%d/%s", uid, label)
	out, err := exec.Command("launchctl", "kickstart", "-k", target).CombinedOutput()
	if err != nil {
		return fmt.Errorf("launchctl kickstart: %s: %w", strings.TrimSpace(string(out)), err)
	}
	return nil
}

func CurrentBinary() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.EvalSymlinks(exe)
}
