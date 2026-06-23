package slack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const apiBase = "https://slack.com/api"

type Client struct {
	Token      string
	HTTPClient *http.Client
}

func NewClient(token string) *Client {
	return &Client{
		Token: token,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

type APIError struct {
	Code   string
	Needed string
}

func (e *APIError) Error() string {
	if e.Needed != "" {
		return fmt.Sprintf("%s (need scope: %s)", e.Code, e.Needed)
	}
	return e.Code
}

func IsMissingScope(err error) bool {
	if err == nil {
		return false
	}
	if ae, ok := err.(*APIError); ok {
		return ae.Code == "missing_scope"
	}
	return false
}

func IsRevoked(err error) bool {
	if err == nil {
		return false
	}
	if ae, ok := err.(*APIError); ok {
		return ae.Code == "token_revoked" || ae.Code == "account_inactive"
	}
	return false
}

func slackErr(code, needed string) *APIError {
	return &APIError{Code: code, Needed: needed}
}

// Presence is a Slack user presence value from users.getPresence.
type Presence string

const (
	PresenceActive Presence = "active"
	PresenceAway   Presence = "away"
)

// UserInfo is a subset of users.info used for team watch.
type UserInfo struct {
	ID          string
	DisplayName string
	RealName    string
	TZ          string
	AvatarURL   string
}

func (c *Client) GetUserInfo(userID string) (UserInfo, error) {
	var resp struct {
		OK   bool `json:"ok"`
		User struct {
			ID      string `json:"id"`
			TZ      string `json:"tz"`
			Deleted bool   `json:"deleted"`
			Profile struct {
				DisplayName string `json:"display_name"`
				RealName    string `json:"real_name"`
				Image48     string `json:"image_48"`
			} `json:"profile"`
		} `json:"user"`
		Error  string `json:"error"`
		Needed string `json:"needed"`
	}
	if err := c.get("users.info", url.Values{"user": {userID}}, &resp); err != nil {
		return UserInfo{}, err
	}
	if !resp.OK {
		return UserInfo{}, slackErr(resp.Error, resp.Needed)
	}
	if resp.User.Deleted {
		return UserInfo{}, fmt.Errorf("user %s is deleted", userID)
	}
	name := resp.User.Profile.DisplayName
	if name == "" {
		name = resp.User.Profile.RealName
	}
	return UserInfo{
		ID:          resp.User.ID,
		DisplayName: name,
		RealName:    resp.User.Profile.RealName,
		TZ:          resp.User.TZ,
		AvatarURL:   resp.User.Profile.Image48,
	}, nil
}

func (c *Client) GetUserPresence(userID string) (Presence, error) {
	var resp struct {
		OK       bool   `json:"ok"`
		Presence string `json:"presence"`
		Error    string `json:"error"`
		Needed   string `json:"needed"`
	}
	if err := c.get("users.getPresence", url.Values{"user": {userID}}, &resp); err != nil {
		return "", err
	}
	if !resp.OK {
		return "", slackErr(resp.Error, resp.Needed)
	}
	switch Presence(resp.Presence) {
	case PresenceActive, PresenceAway:
		return Presence(resp.Presence), nil
	default:
		return "", fmt.Errorf("unknown presence %q", resp.Presence)
	}
}

func (c *Client) AuthTest() (userID, teamID string, err error) {
	var resp struct {
		OK     bool   `json:"ok"`
		UserID string `json:"user_id"`
		TeamID string `json:"team_id"`
		Error  string `json:"error"`
	}
	if err := c.get("auth.test", url.Values{}, &resp); err != nil {
		return "", "", err
	}
	if !resp.OK {
		return "", "", &APIError{Code: resp.Error}
	}
	return resp.UserID, resp.TeamID, nil
}

func (c *Client) get(method string, params url.Values, out interface{}) error {
	u, _ := url.Parse(apiBase + "/" + method)
	u.RawQuery = params.Encode()
	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.Token)
	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	return decodeJSON(res.Body, out)
}

func (c *Client) post(method string, payload interface{}, out interface{}) error {
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, apiBase+"/"+method, bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	return decodeJSON(res.Body, out)
}

func decodeJSON(r io.Reader, out interface{}) error {
	dec := json.NewDecoder(r)
	return dec.Decode(out)
}

// ExchangeOAuthCode trades an authorization code for a user token.
func ExchangeOAuthCode(code, clientID, clientSecret, redirectURI string) (token, userID, teamID string, err error) {
	form := url.Values{
		"client_id":     {clientID},
		"client_secret": {clientSecret},
		"code":          {code},
		"redirect_uri":  {redirectURI},
	}
	req, err := http.NewRequest(http.MethodPost, apiBase+"/oauth.v2.access", bytes.NewReader([]byte(form.Encode())))
	if err != nil {
		return "", "", "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", "", "", err
	}
	defer res.Body.Close()
	var resp struct {
		OK         bool   `json:"ok"`
		Error      string `json:"error"`
		AuthedUser struct {
			ID          string `json:"id"`
			AccessToken string `json:"access_token"`
		} `json:"authed_user"`
		Team struct {
			ID string `json:"id"`
		} `json:"team"`
	}
	if err := decodeJSON(res.Body, &resp); err != nil {
		return "", "", "", err
	}
	if !resp.OK {
		return "", "", "", &APIError{Code: resp.Error}
	}
	if resp.AuthedUser.AccessToken == "" {
		return "", "", "", fmt.Errorf("missing user access token in OAuth response")
	}
	return resp.AuthedUser.AccessToken, resp.AuthedUser.ID, resp.Team.ID, nil
}
