//go:build darwin

package cmd

import "github.com/haritabh17/theirtime/internal/menubar"

func runMenubar(demo bool) error {
	if demo {
		return menubar.RunDemo()
	}
	return menubar.Run()
}
