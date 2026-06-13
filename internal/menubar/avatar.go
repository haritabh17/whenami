package menubar

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

const avatarHTTPTimeout = 10 * time.Second

func fetchAvatar(url string) ([]byte, error) {
	if url == "" {
		return nil, fmt.Errorf("empty avatar url")
	}
	client := &http.Client{Timeout: avatarHTTPTimeout}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("avatar http %d", resp.StatusCode)
	}
	const maxAvatarBytes = 512 << 10
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxAvatarBytes))
	if err != nil {
		return nil, err
	}
	if len(body) == 0 {
		return nil, fmt.Errorf("empty avatar body")
	}
	return body, nil
}
