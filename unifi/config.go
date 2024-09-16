package unifi

import (
	"fmt"
	"strings"

	"github.com/ovh/configstore"
)

const (
	keyUrl      = "UNIFI_URL"
	keySite     = "UNIFI_SITE"
	keyUsername = "UNIFI_USERNAME"
	keyPassword = "UNIFI_PASSWORD"
)

type config struct {
	Url      string
	Site     string
	Username string
	Password string
}

func loadConfig() (*config, error) {
	var cfg config

	url, err := configstore.GetItemValue(keyUrl)
	if err != nil {
		if _, ok := err.(configstore.ErrItemNotFound); ok {
			err = fmt.Errorf("not found")
		}
		return nil, fmt.Errorf("unable to get the unifi URL: %w", err)
	} else {
		cfg.Url = strings.TrimSuffix(url, "/") + "/api"
	}

	site, err := configstore.GetItemValue(keySite)
	if err != nil {
		if _, ok := err.(configstore.ErrItemNotFound); ok {
			err = fmt.Errorf("not found")
		}
		return nil, fmt.Errorf("unable to get the unifi site: %w", err)
	} else {
		cfg.Site = site
	}

	username, err := configstore.GetItemValue(keyUsername)
	if err != nil {
		if _, ok := err.(configstore.ErrItemNotFound); ok {
			err = fmt.Errorf("not found")
		}
		return nil, fmt.Errorf("unable to get the unifi username: %w", err)
	} else {
		cfg.Username = username
	}

	password, err := configstore.GetItemValue(keyPassword)
	if err != nil {
		if _, ok := err.(configstore.ErrItemNotFound); ok {
			err = fmt.Errorf("not found")
		}
		return nil, fmt.Errorf("unable to get the unifi password: %w", err)
	} else {
		cfg.Password = password
	}

	return &cfg, nil
}
