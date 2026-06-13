package manifest

import (
	_ "embed"
	"net/url"
)

//go:embed theirtime.manifest.yaml
var YAML string

// CreateAppURL opens Slack's "create app from manifest" flow with theirtime pre-filled.
func CreateAppURL() string {
	return "https://api.slack.com/apps?new_app=1&manifest_yaml=" + url.QueryEscape(YAML)
}

// BasicInfoHint is shown after the user creates their app.
const BasicInfoHint = "In Slack: open your new app → Settings → Basic Information → App Credentials"
