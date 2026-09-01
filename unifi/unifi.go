package unifi

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"

	"github.com/sirupsen/logrus"
)

type Client struct {
	cfg *config
	clt *http.Client

	// unifiOS reports whether the controller is hosted by UniFi OS, which
	// serves the network application behind a /proxy/network prefix.
	unifiOS bool
}

func NewClient(ctx context.Context) (*Client, error) {
	// load the configuration
	cfg, err := loadConfig()
	if err != nil {
		return nil, fmt.Errorf("unable to load the configuration: %w", err)
	}
	logrus.WithContext(ctx).Debug("loaded the unifi configuration")

	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("unable to create a cookies jar: %w", err)
	}

	transport := http.DefaultTransport.(*http.Transport).Clone()
	if cfg.Insecure {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		logrus.WithContext(ctx).Warn("TLS certificate verification is disabled")
	}

	// build the client
	client := &Client{
		cfg: cfg,
		clt: &http.Client{
			Jar:       jar,
			Transport: transport,
		},
		// An API key only exists on UniFi OS, so the prefix is known upfront.
		unifiOS: cfg.ApiKey != "",
	}

	return client, nil
}

// apiPath prefixes a network application path with the right API root.
func (c *Client) apiPath(uri string) string {
	if c.unifiOS {
		return "/proxy/network/api" + uri
	}
	return "/api" + uri
}

func (c *Client) do(ctx context.Context, method, uri string, headers http.Header, queryArgs map[string]string, body any) (*http.Response, error) {
	parsedURL, err := url.Parse(c.cfg.Url + uri)
	if err != nil {
		return nil, fmt.Errorf("unable to parse the URL: %w", err)
	}

	q := parsedURL.Query()
	for key, value := range queryArgs {
		q.Add(key, value)
	}
	parsedURL.RawQuery = q.Encode()

	var bodyReader io.Reader
	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("unable to marshal the body: %w", err)
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	req, err := http.NewRequestWithContext(ctx, method, parsedURL.String(), bodyReader)
	if err != nil {
		return nil, fmt.Errorf("unable to build the request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.cfg.ApiKey != "" {
		req.Header.Set("X-API-KEY", c.cfg.ApiKey)
	}

	for k, values := range headers {
		for _, v := range values {
			req.Header.Add(k, v)
		}
	}

	res, err := c.clt.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error while executing the request: %w", err)
	}

	if c.clt.Jar != nil {
		cookies := res.Cookies()
		parsedUrlWithoutQuery := &url.URL{Scheme: parsedURL.Scheme, Host: parsedURL.Host, Path: parsedURL.Path}
		c.clt.Jar.SetCookies(parsedUrlWithoutQuery, cookies)
	}

	return res, err
}

func (c *Client) Login(ctx context.Context) error {
	// An API key authenticates every single request: no session is needed.
	if c.cfg.ApiKey != "" {
		logrus.WithContext(ctx).Debug("using the API key authentication, skipping the login")
		return nil
	}

	credentials := map[string]string{
		"username": c.cfg.Username,
		"password": c.cfg.Password,
	}

	// UniFi OS consoles authenticate on /api/auth/login, whereas the standalone
	// network application uses /api/login. Try the former, then fall back.
	res, err := c.do(ctx, http.MethodPost, "/api/auth/login", nil, nil, credentials)
	if err != nil {
		return fmt.Errorf("unable to execute the query: %w", err)
	}
	drain(res)

	if res.StatusCode == http.StatusOK {
		c.unifiOS = true
		logrus.WithContext(ctx).Debug("logged in to a unifi-os controller")
		return nil
	}

	res, err = c.do(ctx, http.MethodPost, "/api/login", nil, nil, credentials)
	if err != nil {
		return fmt.Errorf("unable to execute the query: %w", err)
	}
	drain(res)

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}
	c.unifiOS = false
	logrus.WithContext(ctx).Debug("logged in to a standalone controller")

	return nil
}

func (c *Client) GetNetworks(ctx context.Context) ([]NetworkConf, error) {
	res, err := c.do(ctx, http.MethodGet, c.apiPath("/s/"+c.cfg.Site+"/rest/networkconf"), nil, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("unable to execute the query: %w", err)
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}

	var result result[[]NetworkConf]
	if err := unmarshal(res, &result); err != nil {
		return nil, err
	}
	return result.Data, nil
}

func (c *Client) GetUsers(ctx context.Context) ([]User, error) {
	res, err := c.do(ctx, http.MethodGet, c.apiPath("/s/"+c.cfg.Site+"/list/user"), nil, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("unable to execute the query: %w", err)
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}

	var result result[[]User]
	if err := unmarshal(res, &result); err != nil {
		return nil, err
	}
	return result.Data, nil
}

// drain consumes and closes a response body so the connection can be reused.
func drain(res *http.Response) {
	if res == nil || res.Body == nil {
		return
	}
	io.Copy(io.Discard, res.Body)
	res.Body.Close()
}

func unmarshal(res *http.Response, ret any) error {
	dataBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("unable to read the data: %w", err)
	}

	if err := json.Unmarshal(dataBytes, ret); err != nil {
		return fmt.Errorf("unable to unmarshal the data: %w", err)
	}

	return nil
}
