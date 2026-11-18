package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

type Client struct {
	BaseURL string
	Token   string
	http    *http.Client
}

const maxBodyBytes = 32 << 20

func New(baseURL, token string) *Client {
	return &Client{BaseURL: baseURL, Token: token, http: &http.Client{Timeout: 30 * time.Second}}
}

func (c *Client) checkScheme() error {
	u, err := url.Parse(c.BaseURL)
	if err != nil {
		return fmt.Errorf("invalid base URL: %w", err)
	}
	if u.Scheme == "https" {
		return nil
	}
	host := u.Hostname()
	if host == "localhost" || host == "127.0.0.1" || host == "::1" || os.Getenv("UPTRIVA_INSECURE_HTTP") != "" {
		return nil
	}
	return fmt.Errorf("refusing insecure base URL %q — credentials would travel in cleartext (set UPTRIVA_INSECURE_HTTP=1 to override)", c.BaseURL)
}

func sanitizeError(s string) string {
	return strings.Map(func(r rune) rune {
		if r < 0x20 || r == 0x7f {
			return -1
		}
		return r
	}, s)
}

func (c *Client) do(method, path string, body any) ([]byte, error) {
	if err := c.checkScheme(); err != nil {
		return nil, err
	}
	var rdr io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		rdr = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, c.BaseURL+path, rdr)
	if err != nil {
		return nil, err
	}
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(io.LimitReader(resp.Body, maxBodyBytes))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		var e struct {
			Error string `json:"error"`
		}
		if json.Unmarshal(b, &e) == nil && e.Error != "" {
			return nil, fmt.Errorf("API %d: %s", resp.StatusCode, sanitizeError(e.Error))
		}
		return nil, fmt.Errorf("API %d", resp.StatusCode)
	}
	return b, nil
}

func (c *Client) PostPublic(path string, authHeader string, out any) error {
	if err := c.checkScheme(); err != nil {
		return err
	}
	rdr := bytes.NewReader([]byte("{}"))
	req, err := http.NewRequest("POST", c.BaseURL+path, rdr)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, maxBodyBytes))
	if resp.StatusCode >= 400 {
		var e struct {
			Error string `json:"error"`
		}
		if json.Unmarshal(respBody, &e) == nil && e.Error != "" {
			return fmt.Errorf("API %d: %s", resp.StatusCode, sanitizeError(e.Error))
		}
		return fmt.Errorf("API %d", resp.StatusCode)
	}
	if out != nil && len(respBody) > 0 {
		return json.Unmarshal(respBody, out)
	}
	return nil
}

func (c *Client) Get(path string, out any) error {
	b, err := c.do("GET", path, nil)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, out)
}

func (c *Client) Post(path string, body, out any) error {
	b, err := c.do("POST", path, body)
	if err != nil {
		return err
	}
	if out != nil {
		return json.Unmarshal(b, out)
	}
	return nil
}

func (c *Client) Put(path string, body, out any) error {
	b, err := c.do("PUT", path, body)
	if err != nil {
		return err
	}
	if out != nil {
		return json.Unmarshal(b, out)
	}
	return nil
}

func (c *Client) Delete(path string) error {
	_, err := c.do("DELETE", path, nil)
	return err
}
