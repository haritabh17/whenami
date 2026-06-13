package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/haritabh17/theirtime/internal/openurl"
	"github.com/haritabh17/theirtime/internal/slack"
	"github.com/haritabh17/theirtime/internal/ui"
)

const (
	OAuthPort       = 8765
	RedirectPath    = "/callback"
	UserScope       = "users:read"
	callbackTimeout = 5 * time.Minute
)

func RedirectURI() string {
	return fmt.Sprintf("http://127.0.0.1:%d%s", OAuthPort, RedirectPath)
}

type Result struct {
	Token  string
	UserID string
	TeamID string
}

func Authenticate(clientID, clientSecret string) (*Result, error) {
	state, err := randomState()
	if err != nil {
		return nil, err
	}

	codeCh := make(chan string, 1)
	errCh := make(chan error, 1)

	mux := http.NewServeMux()
	server := &http.Server{
		Addr:    fmt.Sprintf("127.0.0.1:%d", OAuthPort),
		Handler: mux,
	}

	mux.HandleFunc(RedirectPath, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("state") != state {
			errCh <- fmt.Errorf("invalid OAuth state")
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("Invalid state. You can close this tab."))
			return
		}
		if e := r.URL.Query().Get("error"); e != "" {
			errCh <- fmt.Errorf("slack OAuth error: %s", e)
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("Authorization failed. You can close this tab."))
			return
		}
		code := r.URL.Query().Get("code")
		if code == "" {
			errCh <- fmt.Errorf("missing authorization code")
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>theirtime connected</title>
  <style>
    body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
           display: flex; align-items: center; justify-content: center; min-height: 100vh;
           margin: 0; background: #0f0f10; color: #f5f5f7; }
    .card { text-align: center; padding: 2.5rem 3rem; border-radius: 16px;
            background: #1c1c1e; border: 1px solid #3a3a3c; max-width: 420px; }
    h1 { font-size: 1.5rem; font-weight: 600; margin: 0 0 0.5rem; }
    p { color: #98989d; margin: 0; line-height: 1.5; }
    .mark { color: #30d158; font-size: 2rem; margin-bottom: 1rem; }
  </style>
</head>
<body>
  <div class="card">
    <div class="mark">✓</div>
    <h1>theirtime connected</h1>
    <p>Return to your terminal to finish setup.</p>
  </div>
</body>
</html>`))
		codeCh <- code
	})

	ln, err := net.Listen("tcp", server.Addr)
	if err != nil {
		return nil, fmt.Errorf("listen on %s: %w (is another theirtime auth running?)", server.Addr, err)
	}

	go func() {
		_ = server.Serve(ln)
	}()

	authURL := authorizeURL(clientID, state)
	if err := openurl.Open(authURL); err != nil {
		ui.URL("Could not open browser. Open this URL manually:", authURL)
	}

	var code string
	select {
	case code = <-codeCh:
	case err := <-errCh:
		_ = server.Shutdown(context.Background())
		return nil, err
	case <-time.After(callbackTimeout):
		_ = server.Shutdown(context.Background())
		return nil, fmt.Errorf("OAuth timed out after %s", callbackTimeout)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = server.Shutdown(ctx)

	token, userID, teamID, err := slack.ExchangeOAuthCode(code, clientID, clientSecret, RedirectURI())
	if err != nil {
		return nil, err
	}

	return &Result{Token: token, UserID: userID, TeamID: teamID}, nil
}

func authorizeURL(clientID, state string) string {
	params := url.Values{
		"client_id":    {clientID},
		"user_scope":   {UserScope},
		"redirect_uri": {RedirectURI()},
		"state":        {state},
	}
	return "https://slack.com/oauth/v2/authorize?" + params.Encode()
}

func randomState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
