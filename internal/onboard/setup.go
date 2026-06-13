package onboard

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/haritabh17/theirtime/internal/keychain"
	"github.com/haritabh17/theirtime/internal/manifest"
	"github.com/haritabh17/theirtime/internal/openurl"
	"github.com/haritabh17/theirtime/internal/slackcfg"
	"github.com/haritabh17/theirtime/internal/ui"
	"golang.org/x/term"
)

func EnsureAppCredentials() (clientID, clientSecret string, err error) {
	if id, secret, err := slackcfg.Credentials(); err == nil {
		return id, secret, nil
	}

	ui.Step(1, 2, "Create your Slack app")
	ui.Muted("You create and own the app. Credentials stay in your Keychain.")
	ui.Blank()
	ui.Action("Opening Slack with a pre-filled manifest…")
	ui.Blank()

	createURL := manifest.CreateAppURL()
	if err := openurl.Open(createURL); err != nil {
		ui.URL("Could not open browser. Open this URL manually:", createURL)
		ui.Blank()
	}

	ui.BrowserSteps(manifest.BasicInfoHint)

	reader := bufio.NewReader(os.Stdin)

	ui.Prompt("Client ID")
	clientID, err = readLine(reader)
	if err != nil {
		return "", "", err
	}
	clientID = strings.TrimSpace(clientID)
	if clientID == "" {
		return "", "", fmt.Errorf("Client ID is required")
	}

	ui.Prompt("Client Secret")
	secretBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
	ui.Blank()
	if err != nil {
		return "", "", fmt.Errorf("read Client Secret: %w", err)
	}
	clientSecret = strings.TrimSpace(string(secretBytes))
	if clientSecret == "" {
		return "", "", fmt.Errorf("Client Secret is required")
	}

	if err := keychain.SetAppCredentials(clientID, clientSecret); err != nil {
		return "", "", fmt.Errorf("save app credentials to Keychain: %w", err)
	}

	ui.Success("App credentials saved to Keychain")
	ui.Blank()
	return clientID, clientSecret, nil
}

func readLine(r *bufio.Reader) (string, error) {
	line, err := r.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(line), nil
}
