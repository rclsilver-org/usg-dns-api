package unifi

import (
	"bytes"
	"context"
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

	// build the client
	client := &Client{
		cfg: cfg,
		clt: &http.Client{
			Jar: jar,
		},
	}

	return client, nil
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
	res, err := c.do(ctx, http.MethodPost, "/login", nil, nil, map[string]string{
		"username": c.cfg.Username,
		"password": c.cfg.Password,
	})
	if err != nil {
		return fmt.Errorf("unable to execute the query: %w", err)
	}
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}

	return nil
}

func (c *Client) GetNetworks(ctx context.Context) ([]NetworkConf, error) {
	res, err := c.do(ctx, http.MethodGet, "/s/"+c.cfg.Site+"/rest/networkconf", nil, nil, nil)
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
	res, err := c.do(ctx, http.MethodGet, "/s/"+c.cfg.Site+"/list/user", nil, nil, nil)
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
