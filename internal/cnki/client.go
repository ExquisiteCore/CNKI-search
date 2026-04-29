package cnki

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"

	"github.com/ExquisiteCore/cnki-search/internal/model"
)

// ClientOptions configures the HTTP CNKI client.
type ClientOptions struct {
	BaseURL    string
	HTTPClient *http.Client
	UserAgent  string
}

// Client talks to CNKI over HTTP and keeps cookies in its HTTP client.
type Client struct {
	baseURL    *url.URL
	httpClient *http.Client
	userAgent  string
}

// NewClient builds an HTTP CNKI client. BaseURL is mainly for tests; production
// callers should leave it empty to use https://kns.cnki.net.
func NewClient(opts ClientOptions) *Client {
	base := strings.TrimRight(opts.BaseURL, "/")
	if base == "" {
		base = URLKNSBase
	}
	parsed, err := url.Parse(base)
	if err != nil {
		parsed, _ = url.Parse(URLKNSBase)
	}

	hc := opts.HTTPClient
	if hc == nil {
		jar, _ := cookiejar.New(nil)
		hc = &http.Client{Jar: jar}
	} else if hc.Jar == nil {
		jar, _ := cookiejar.New(nil)
		clone := *hc
		clone.Jar = jar
		hc = &clone
	}

	ua := opts.UserAgent
	if ua == "" {
		ua = defaultUserAgent
	}

	return &Client{
		baseURL:    parsed,
		httpClient: hc,
		userAgent:  ua,
	}
}

func (c *Client) resolve(path string) string {
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return path
	}
	ref, err := url.Parse(path)
	if err != nil {
		return c.baseURL.String()
	}
	return c.baseURL.ResolveReference(ref).String()
}

func (c *Client) newRequest(ctx context.Context, method, rawURL string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, rawURL, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	return req, nil
}

func (c *Client) ensureClientID(ctx context.Context) error {
	if c.baseURL.Host != "kns.cnki.net" || c.httpClient.Jar == nil {
		return nil
	}

	knsURL, err := url.Parse(URLKNSBase)
	if err != nil {
		return err
	}
	for _, cookie := range c.httpClient.Jar.Cookies(knsURL) {
		if cookie.Name == "Ecp_ClientId" && cookie.Value != "" {
			return nil
		}
	}

	req, err := c.newRequest(ctx, http.MethodGet, URLClientID, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Origin", URLHome)
	req.Header.Set("Referer", URLHome+"/")

	body, err := c.doText(req)
	if err != nil {
		return err
	}
	var payload struct {
		Success bool   `json:"Success"`
		Data    string `json:"Data"`
	}
	if err := json.Unmarshal([]byte(body), &payload); err != nil {
		return fmt.Errorf("parse client id response: %w", err)
	}
	if payload.Data == "" {
		return fmt.Errorf("client id response did not include Data")
	}
	c.httpClient.Jar.SetCookies(knsURL, []*http.Cookie{{
		Name:  "Ecp_ClientId",
		Value: payload.Data,
		Path:  "/",
	}})
	return nil
}

func (c *Client) doText(req *http.Request) (string, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	body := string(data)
	if resp.StatusCode == http.StatusForbidden || strings.Contains(body, `"code":-403`) || looksLikeCaptcha(body) {
		return "", ErrCaptcha
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("cnki http %d: %s", resp.StatusCode, http.StatusText(resp.StatusCode))
	}
	return body, nil
}

func looksLikeCaptcha(body string) bool {
	lower := strings.ToLower(body)
	return strings.Contains(lower, "captcha") ||
		strings.Contains(body, "验证码") ||
		strings.Contains(body, "滑块验证") ||
		strings.Contains(body, "安全验证") ||
		strings.Contains(body, "拖动下方拼图") ||
		strings.Contains(lower, "/verify/home")
}

// Search is a convenience wrapper using a default HTTP client.
func Search(ctx context.Context, q model.Query) (*model.SearchResult, error) {
	return NewClient(ClientOptions{}).Search(ctx, q)
}

// Detail is a convenience wrapper using a default HTTP client.
func Detail(ctx context.Context, rawURL string, withRefs bool) (*model.Detail, error) {
	return NewClient(ClientOptions{}).Detail(ctx, rawURL, withRefs)
}

// References is a convenience wrapper using a default HTTP client.
func References(ctx context.Context, rawURL string) ([]model.Reference, error) {
	return NewClient(ClientOptions{}).References(ctx, rawURL)
}

const defaultUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/136.0.0.0 Safari/537.36"
