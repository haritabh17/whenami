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
		_, _ = w.Write([]byte(`<!DOCTYPE html><html><body><p>theirtime connected. You can close this tab and return to the terminal.</p></body></html>`))
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
		fmt.Printf("Open this URL in your browser:\n%s\n", authURL)
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
