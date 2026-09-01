package unifi

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/ovh/configstore"
)

const (
	keyUrl      = "UNIFI_URL"
	keySite     = "UNIFI_SITE"
	keyUsername = "UNIFI_USERNAME"
	keyPassword = "UNIFI_PASSWORD"
	keyApiKey   = "UNIFI_API_KEY"
	keyInsecure = "UNIFI_INSECURE"
)

type config struct {
	// Url is the controller base URL, without any trailing slash or API prefix.
	Url string

	Site string

	// Username and Password are only required when ApiKey is empty.
	Username string
	Password string

	// ApiKey enables the X-API-KEY authentication of UniFi OS. When set, no
	// login round-trip is performed.
	ApiKey string

	// Insecure skips the TLS certificate verification, which is needed for the
	// self-signed certificate a controller ships with.
	Insecure bool
}

// optionalValue returns an empty string when the item is not configured.
func optionalValue(key string) (string, error) {
	value, err := configstore.GetItemValue(key)
	if err != nil {
		if _, ok := err.(configstore.ErrItemNotFound); ok {
			return "", nil
		}
		return "", err
	}
	return value, nil
}

// requiredValue returns an explicit "not found" error when the item is missing.
func requiredValue(key string) (string, error) {
	value, err := configstore.GetItemValue(key)
	if err != nil {
		if _, ok := err.(configstore.ErrItemNotFound); ok {
			err = fmt.Errorf("not found")
		}
		return "", err
	}
	return value, nil
}

func loadConfig() (*config, error) {
	var cfg config

	url, err := requiredValue(keyUrl)
	if err != nil {
		return nil, fmt.Errorf("unable to get the unifi URL: %w", err)
	}
	cfg.Url = strings.TrimSuffix(url, "/")

	site, err := requiredValue(keySite)
	if err != nil {
		return nil, fmt.Errorf("unable to get the unifi site: %w", err)
	}
	cfg.Site = site

	apiKey, err := optionalValue(keyApiKey)
	if err != nil {
		return nil, fmt.Errorf("unable to get the unifi API key: %w", err)
	}
	cfg.ApiKey = apiKey

	// The username and the password are only needed when no API key is set.
	if cfg.ApiKey == "" {
		username, err := requiredValue(keyUsername)
		if err != nil {
			return nil, fmt.Errorf("unable to get the unifi username: %w", err)
		}
		cfg.Username = username

		password, err := requiredValue(keyPassword)
		if err != nil {
			return nil, fmt.Errorf("unable to get the unifi password: %w", err)
		}
		cfg.Password = password
	}

	insecure, err := optionalValue(keyInsecure)
	if err != nil {
		return nil, fmt.Errorf("unable to get the unifi insecure flag: %w", err)
	}
	if insecure != "" {
		value, err := strconv.ParseBool(insecure)
		if err != nil {
			return nil, fmt.Errorf("unable to parse the unifi insecure flag: %w", err)
		}
		cfg.Insecure = value
	}

	return &cfg, nil
}
