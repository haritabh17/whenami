package openurl

import (
	"fmt"
	"os/exec"
	"runtime"
)

func Open(u string) error {
	switch runtime.GOOS {
	case "darwin":
		return exec.Command("open", u).Start()
	default:
		return fmt.Errorf("open URL not supported on %s", runtime.GOOS)
	}
}
