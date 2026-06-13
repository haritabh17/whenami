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
	"golang.org/x/term"
)

func EnsureAppCredentials() (clientID, clientSecret string, err error) {
	if id, secret, err := slackcfg.Credentials(); err == nil {
		return id, secret, nil
	}

	fmt.Println()
	fmt.Println("Step 1 of 2 — Create your personal Slack app")
	fmt.Println("────────────────────────────────────────────")
	fmt.Println("theirtime uses a Slack app that you create and control.")
	fmt.Println("Opening Slack with a pre-filled app manifest…")
	fmt.Println()

	createURL := manifest.CreateAppURL()
	if err := openurl.Open(createURL); err != nil {
		fmt.Printf("Could not open browser. Open this URL manually:\n%s\n\n", createURL)
	}

	fmt.Println("In the browser:")
	fmt.Println("  1. Sign in to Slack (if asked)")
	fmt.Println("  2. Pick a workspace and click Next")
	fmt.Println("  3. Review the manifest and click Create")
	fmt.Println("  4.", manifest.BasicInfoHint)
	fmt.Println("  5. Copy Client ID and Client Secret")
	fmt.Println()

	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Paste Client ID: ")
	clientID, err = readLine(reader)
	if err != nil {
		return "", "", err
	}
	clientID = strings.TrimSpace(clientID)
	if clientID == "" {
		return "", "", fmt.Errorf("Client ID is required")
	}

	fmt.Print("Paste Client Secret (input hidden): ")
	secretBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()
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

	fmt.Println("App credentials saved to Keychain.")
	fmt.Println()
	return clientID, clientSecret, nil
}

func readLine(r *bufio.Reader) (string, error) {
	line, err := r.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(line), nil
}
