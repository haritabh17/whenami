package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/haritabh17/theirtime/internal/timeformat"
	"gopkg.in/yaml.v3"
)

const (
	DirName  = "theirtime"
	FileName = "config.yaml"
)

var fileMu sync.Mutex

// Config holds preferences — never secrets.
type Config struct {
	SlackUserID   string       `yaml:"slack_user_id,omitempty"`
	Format24h     bool         `yaml:"format_24h,omitempty"`
	ShowAvatar    *bool        `yaml:"show_avatar,omitempty"`
	ShowName      *bool        `yaml:"show_name,omitempty"`
	ShowTime      *bool        `yaml:"show_time,omitempty"`
	TimePrecision string       `yaml:"time_precision,omitempty"` // hours | minutes
	Team          []TeamMember `yaml:"team,omitempty"`
}

func ShowAvatar(c *Config) bool {
	if c == nil || c.ShowAvatar == nil {
		return true
	}
	return *c.ShowAvatar
}

func ShowName(c *Config) bool {
	if c == nil || c.ShowName == nil {
		return false
	}
	return *c.ShowName
}

func ShowTime(c *Config) bool {
	if c == nil || c.ShowTime == nil {
		return true
	}
	return *c.ShowTime
}

func TimePrecision(c *Config) string {
	if c == nil || c.TimePrecision == "" {
		return timeformat.PrecisionMinutes
	}
	if c.TimePrecision == timeformat.PrecisionHours {
		return timeformat.PrecisionHours
	}
	return timeformat.PrecisionMinutes
}

// ValidateDisplay ensures at least one visibility toggle is enabled.
func ValidateDisplay(c *Config) error {
	if !ShowAvatar(c) && !ShowName(c) && !ShowTime(c) {
		return fmt.Errorf("at least one of show_avatar, show_name, or show_time must be true")
	}
	return nil
}

func ApplyDefaults(c *Config) {
	if c == nil {
		return
	}
	if c.ShowAvatar == nil {
		t := true
		c.ShowAvatar = &t
	}
	if c.ShowName == nil {
		f := false
		c.ShowName = &f
	}
	if c.ShowTime == nil {
		t := true
		c.ShowTime = &t
	}
	if c.TimePrecision == "" {
		c.TimePrecision = timeformat.PrecisionMinutes
	}
}

func Path() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, DirName, FileName), nil
}

func Load() (*Config, error) {
	fileMu.Lock()
	defer fileMu.Unlock()
	return loadUnlocked()
}

func loadUnlocked() (*Config, error) {
	path, err := Path()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			c := &Config{}
			ApplyDefaults(c)
			return c, nil
		}
		return nil, err
	}
	var c Config
	if err := yaml.Unmarshal(data, &c); err != nil {
		return nil, err
	}
	ApplyDefaults(&c)
	return &c, nil
}

func Save(c *Config) error {
	if err := ValidateDisplay(c); err != nil {
		return err
	}
	fileMu.Lock()
	defer fileMu.Unlock()

	path, err := Path()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
