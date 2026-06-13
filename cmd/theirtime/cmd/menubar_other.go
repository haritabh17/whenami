//go:build !darwin

package cmd

import "fmt"

func runMenubar(demo bool) error {
	_ = demo
	return fmt.Errorf("theirtime menubar is macOS only")
}
