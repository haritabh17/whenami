package keychain

import (
	"fmt"

	"github.com/99designs/keyring"
)

const (
	service             = "theirtime"
	accountUserToken    = "slack-user-token"
	accountClientID     = "slack-client-id"
	accountClientSecret = "slack-client-secret"

	labelUserToken    = "theirtime user token"
	labelClientID     = "theirtime client ID"
	labelClientSecret = "theirtime client secret"
)

func openKeyring() (keyring.Keyring, error) {
	return keyring.Open(keyring.Config{
		ServiceName:                    service,
		KeychainTrustApplication:       true,
		KeychainAccessibleWhenUnlocked: true,
	})
}

func item(key, label string, data []byte) keyring.Item {
	return keyring.Item{
		Key:         key,
		Label:       label,
		Description: label,
		Data:        data,
	}
}

func storedKeys() (map[string]bool, error) {
	kr, err := openKeyring()
	if err != nil {
		return nil, err
	}
	keys, err := kr.Keys()
	if err != nil {
		return nil, err
	}
	set := make(map[string]bool, len(keys))
	for _, k := range keys {
		set[k] = true
	}
	return set, nil
}

func SetToken(token string) error {
	kr, err := openKeyring()
	if err != nil {
		return err
	}
	return kr.Set(item(accountUserToken, labelUserToken, []byte(token)))
}

func GetToken() (string, error) {
	kr, err := openKeyring()
	if err != nil {
		return "", err
	}
	got, err := kr.Get(accountUserToken)
	if err != nil {
		return "", fmt.Errorf("no token in Keychain — run theirtime onboard")
	}
	return string(got.Data), nil
}

func DeleteToken() error {
	kr, err := openKeyring()
	if err != nil {
		return err
	}
	return kr.Remove(accountUserToken)
}

func HasToken() bool {
	set, err := storedKeys()
	if err != nil {
		return false
	}
	return set[accountUserToken]
}

func Presence() (hasToken, hasAppCredentials bool) {
	set, err := storedKeys()
	if err != nil {
		return false, false
	}
	return set[accountUserToken], set[accountClientID] && set[accountClientSecret]
}

func SetAppCredentials(clientID, clientSecret string) error {
	kr, err := openKeyring()
	if err != nil {
		return err
	}
	if err := kr.Set(item(accountClientID, labelClientID, []byte(clientID))); err != nil {
		return err
	}
	return kr.Set(item(accountClientSecret, labelClientSecret, []byte(clientSecret)))
}

func GetAppCredentials() (clientID, clientSecret string, err error) {
	kr, err := openKeyring()
	if err != nil {
		return "", "", err
	}
	idItem, err := kr.Get(accountClientID)
	if err != nil {
		return "", "", fmt.Errorf("no Slack app credentials — run theirtime onboard")
	}
	secretItem, err := kr.Get(accountClientSecret)
	if err != nil {
		return "", "", fmt.Errorf("no Slack app credentials — run theirtime onboard")
	}
	return string(idItem.Data), string(secretItem.Data), nil
}

func DeleteAppCredentials() error {
	kr, err := openKeyring()
	if err != nil {
		return err
	}
	_ = kr.Remove(accountClientID)
	return kr.Remove(accountClientSecret)
}

func HasAppCredentials() bool {
	set, err := storedKeys()
	if err != nil {
		return false
	}
	return set[accountClientID] && set[accountClientSecret]
}
