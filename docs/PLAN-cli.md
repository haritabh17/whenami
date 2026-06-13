# theirtime macOS CLI — Build Plan

Menu bar teammate times with Slack avatars. No display-name writes.

## Install

```bash
curl -fsSL https://raw.githubusercontent.com/haritabh17/theirtime/main/scripts/install.sh | bash
theirtime onboard
theirtime team add bob U012ABCDEF
theirtime install-agents
```

No central theirtime Slack app. No credentials in release binaries.

## Commands

| Command | Purpose |
|---------|---------|
| `theirtime onboard` | Create Slack app + OAuth |
| `theirtime team add\|list\|remove` | Watch list for menu bar |
| `theirtime menubar` | Menu bar UI (LaunchAgent when team non-empty) |
| `theirtime install-agents` | Install/refresh menu bar LaunchAgent |
| `theirtime status` | Config, Keychain, agent state |
| `theirtime config get\|set …` | Display toggles, time format |
| `theirtime auth` | Re-authenticate |
| `theirtime offboard` | Uninstall agents, Keychain, config |
| `theirtime version` | Version / commit |

## Local paths

| Path | Contents |
|------|----------|
| Keychain `theirtime` | OAuth token + app credentials |
| `~/Library/Application Support/theirtime/config.yaml` | prefs + team list |
| `~/Library/LaunchAgents/dev.theirtime.menubar.plist` | menu bar agent |
| `~/Library/Logs/theirtime/` | menubar.log |

## Release

Tag `v*` → GoReleaser publishes `theirtime_*_darwin_all.tar.gz` on GitHub Releases.

## Layout

```
theirtime/
  cmd/theirtime/
  internal/menubar/
  internal/team/
  internal/timeformat/
  third_party/systray/
  manifest/theirtime.manifest.yaml
```
