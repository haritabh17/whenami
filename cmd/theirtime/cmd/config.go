package cmd

import (
	"fmt"
	"os"
	"strconv"

	"github.com/haritabh17/theirtime/internal/config"
	"github.com/haritabh17/theirtime/internal/launchagent"
	"github.com/haritabh17/theirtime/internal/timeformat"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "View or change preferences",
}

var configGetCmd = &cobra.Command{
	Use:   "get [key]",
	Short: "Print a config value or all values",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		if len(args) == 0 {
			printAll(cfg)
			return nil
		}
		val, ok := getField(cfg, args[0])
		if !ok {
			return fmt.Errorf("unknown key: %s", args[0])
		}
		fmt.Println(val)
		return nil
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a config value",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		key := args[0]
		if err := setField(cfg, key, args[1]); err != nil {
			return err
		}
		if err := config.Save(cfg); err != nil {
			return err
		}
		if isDisplayKey(key) && launchagent.IsMenubarInstalled() {
			fmt.Fprintf(os.Stderr, "Restart the menu bar agent to apply display changes (theirtime install-agents).\n")
		}
		return nil
	},
}

func init() {
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configSetCmd)
}

func isDisplayKey(key string) bool {
	switch key {
	case "show_avatar", "show_name", "show_time", "format_24h", "time_precision":
		return true
	case "icon_size":
		return true
	default:
		return false
	}
}

func printAll(cfg *config.Config) {
	for _, k := range []string{"show_avatar", "show_name", "show_time", "format_24h", "time_precision", "icon_size"} {
		v, _ := getField(cfg, k)
		fmt.Printf("%s: %s\n", k, v)
	}
}

func getField(cfg *config.Config, key string) (string, bool) {
	switch key {
	case "show_avatar":
		return fmt.Sprintf("%v", config.ShowAvatar(cfg)), true
	case "show_name":
		return fmt.Sprintf("%v", config.ShowName(cfg)), true
	case "show_time":
		return fmt.Sprintf("%v", config.ShowTime(cfg)), true
	case "format_24h":
		return fmt.Sprintf("%v", cfg.Format24h), true
	case "time_precision":
		return config.TimePrecision(cfg), true
	case "icon_size":
		return fmt.Sprintf("%d", config.IconSize(cfg)), true
	default:
		return "", false
	}
}

func setField(cfg *config.Config, key, value string) error {
	switch key {
	case "show_avatar":
		b, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		cfg.ShowAvatar = &b
	case "show_name":
		b, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		cfg.ShowName = &b
	case "show_time":
		b, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		cfg.ShowTime = &b
	case "format_24h":
		b, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		cfg.Format24h = b
	case "time_precision":
		if value != timeformat.PrecisionHours && value != timeformat.PrecisionMinutes {
			return fmt.Errorf("time_precision must be %q or %q", timeformat.PrecisionHours, timeformat.PrecisionMinutes)
		}
		cfg.TimePrecision = value
	case "icon_size":
		size, err := strconv.Atoi(value)
		if err != nil {
			return err
		}
		size = config.NormalizeIconSize(size)
		cfg.IconSize = &size
	default:
		return fmt.Errorf("unknown key: %s", key)
	}
	return config.ValidateDisplay(cfg)
}
