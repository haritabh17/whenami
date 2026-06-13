package slackcfg

import (
	"github.com/haritabh17/theirtime/internal/envload"
	"github.com/haritabh17/theirtime/internal/keychain"
)

func Credentials() (clientID, clientSecret string, err error) {
	if id, secret, err := keychain.GetAppCredentials(); err == nil {
		return id, secret, nil
	}

	envload.Load()
	clientID = getEnv("THEIRTIME_SLACK_CLIENT_ID")
	clientSecret = getEnv("THEIRTIME_SLACK_CLIENT_SECRET")
	if clientID != "" && clientSecret != "" {
		return clientID, clientSecret, nil
	}

	return "", "", ErrNotConfigured
}

var ErrNotConfigured = errNotConfigured{}

type errNotConfigured struct{}

func (errNotConfigured) Error() string {
	return "Slack app not configured — run theirtime onboard"
}

func IsNotConfigured(err error) bool {
	_, ok := err.(errNotConfigured)
	return ok
}
