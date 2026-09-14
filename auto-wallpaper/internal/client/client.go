package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client talks to auto-wallpaper-api.
type Client struct {
	BaseURL string
	Token   string
	HTTP    *http.Client
}

func New(baseURL, token string) *Client {
	return &Client{
		BaseURL: baseURL,
		Token:   token,
		HTTP:    &http.Client{Timeout: 30 * time.Second},
	}
}

type Wallpaper struct {
	ID       int64  `json:"id"`
	Screen   string `json:"screen"`
	Filename string `json:"filename"`
	Enabled  bool   `json:"enabled"`
	FileURL  string `json:"fileUrl"`
}

type Schedule struct {
	Screen        string `json:"screen"`
	Mode          string `json:"mode"`
	IntervalHours int    `json:"intervalHours"`
	StartHour     int    `json:"startHour"`
	Enabled       bool   `json:"enabled"`
}

type ScreenState struct {
	Screen           string      `json:"screen"`
	Schedule         Schedule    `json:"schedule"`
	Wallpapers       []Wallpaper `json:"wallpapers"`
	CurrentWallpaper *Wallpaper  `json:"currentWallpaper"`
	SlotIndex        int64       `json:"slotIndex"`
}

type AgentState struct {
	Now     string        `json:"now"`
	Screens []ScreenState `json:"screens"`
}

// Login obtains a session token.
func (c *Client) Login(username, password string) (string, error) {
	body, _ := json.Marshal(map[string]string{"username": username, "password": password})
	req, err := http.NewRequest(http.MethodPost, c.BaseURL+"/api/auth/login", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := c.HTTP.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(res.Body)
		return "", fmt.Errorf("login HTTP %d: %s", res.StatusCode, string(b))
	}
	var out struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		return "", err
	}
	if out.Token == "" {
		return "", fmt.Errorf("empty token")
	}
	c.Token = out.Token
	return out.Token, nil
}

// EnsureToken logs in when Token empty.
func (c *Client) EnsureToken(username, password string) error {
	if c.Token != "" {
		return nil
	}
	_, err := c.Login(username, password)
	return err
}

// Health checks /health.
func (c *Client) Health() error {
	res, err := c.HTTP.Get(c.BaseURL + "/health")
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("health HTTP %d", res.StatusCode)
	}
	return nil
}

// AgentState fetches rotation state.
func (c *Client) AgentState() (*AgentState, error) {
	req, err := http.NewRequest(http.MethodGet, c.BaseURL+"/api/agent/state", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.Token)
	res, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("agent state HTTP %d: %s", res.StatusCode, string(b))
	}
	var out AgentState
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DownloadFile downloads wallpaper bytes.
func (c *Client) DownloadFile(id int64) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/wallpapers/%d/file", c.BaseURL, id), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.Token)
	res, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download HTTP %d", res.StatusCode)
	}
	return io.ReadAll(res.Body)
}
