package client

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"pixivhono/config"
)

var (
	cacheMu                    sync.RWMutex
	cachedAccessToken          string
	cachedAccessTokenExpiresAt time.Time
	tokenRefreshMu             sync.Mutex
	inflightTokenRefresh       *tokenRefreshPromise
)

type tokenRefreshPromise struct {
	err   error
	done  chan struct{}
	token string
}

type RefreshResponse struct {
	Response struct {
		AccessToken string `json:"access_token"`
	} `json:"response"`
}

func ClearTokenCache() {
	cacheMu.Lock()
	defer cacheMu.Unlock()
	cachedAccessToken = ""
	cachedAccessTokenExpiresAt = time.Time{}
}

func makeClientHash(clientTime string) string {
	data := []byte(clientTime + config.PIXIV_HASH_SECRET)
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

func RefreshPixivAccessToken(refreshToken string) (string, error) {
	clientTime := time.Now().UTC().Format("2006-01-02T15:04:05.000") + "+00:00"
	clientHash := makeClientHash(clientTime)

	form := url.Values{}
	form.Set("client_id", config.PIXIV_CLIENT_ID)
	form.Set("client_secret", config.PIXIV_CLIENT_SECRET)
	form.Set("get_secure_url", "1")
	form.Set("grant_type", "refresh_token")
	form.Set("refresh_token", refreshToken)

	req, err := http.NewRequest("POST", config.PIXIV_OAUTH_URL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}

	req.Header.Set("content-type", "application/x-www-form-urlencoded")
	req.Header.Set("x-client-time", clientTime)
	req.Header.Set("x-client-hash", clientHash)
	req.Header.Set("app-os", "ios")
	req.Header.Set("app-os-version", "16.4.1")
	req.Header.Set("user-agent", "PixivIOSApp/7.16.9 (iOS 16.4.1; iPad13,4)")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("Failed to refresh token (status %d)", resp.StatusCode)
	}

	var data RefreshResponse
	if err := json.Unmarshal(bodyBytes, &data); err != nil {
		return "", err
	}

	if data.Response.AccessToken == "" {
		return "", fmt.Errorf("Failed to refresh token: access_token not found in response")
	}

	return data.Response.AccessToken, nil
}

func GetPixivAccessToken(refreshToken string) (string, error) {
	now := time.Now()

	// 1. Read lock to check cache
	cacheMu.RLock()
	if cachedAccessToken != "" && now.Before(cachedAccessTokenExpiresAt) {
		token := cachedAccessToken
		cacheMu.RUnlock()
		return token, nil
	}
	cacheMu.RUnlock()

	// 2. Lock to handle in-flight refreshing (singleflight pattern)
	tokenRefreshMu.Lock()
	if inflightTokenRefresh != nil {
		promise := inflightTokenRefresh
		tokenRefreshMu.Unlock()
		<-promise.done
		return promise.token, promise.err
	}

	promise := &tokenRefreshPromise{
		done: make(chan struct{}),
	}
	inflightTokenRefresh = promise
	tokenRefreshMu.Unlock()

	// Execute refresh in current goroutine
	token, err := RefreshPixivAccessToken(refreshToken)
	if err == nil {
		cacheMu.Lock()
		cachedAccessToken = token
		ttl := config.GlobalConfig.PixivAccessTokenTTLMs
		if ttl <= 0 {
			ttl = 3000000
		}
		cachedAccessTokenExpiresAt = time.Now().Add(time.Duration(ttl) * time.Millisecond)
		cacheMu.Unlock()
	}

	promise.token = token
	promise.err = err
	close(promise.done)

	tokenRefreshMu.Lock()
	inflightTokenRefresh = nil
	tokenRefreshMu.Unlock()

	return token, err
}

func PixivGet(path string, params map[string]string, accessToken string) (interface{}, error) {
	u, err := url.Parse(config.PIXIV_APP_API_BASE)
	if err != nil {
		return nil, err
	}
	u.Path = path
	q := u.Query()
	for k, v := range params {
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Host", "app-api.pixiv.net")
	req.Header.Set("App-OS", "ios")
	req.Header.Set("App-OS-Version", "14.6")
	req.Header.Set("User-Agent", "PixivIOSApp/7.13.3 (iOS 14.6; iPhone13,2)")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("Pixiv API failed (status %d)", resp.StatusCode)
	}

	var data interface{}
	if err := json.Unmarshal(bodyBytes, &data); err != nil {
		return nil, err
	}

	return data, nil
}
