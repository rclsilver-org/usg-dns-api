package db

import (
	"fmt"

	"github.com/ovh/configstore"
)

const (
	keyPath = "DB_PATH"
)

var (
	defaultPath = "usg-dns-api.db"
)

type config struct {
	Path string
}

func loadConfig() (*config, error) {
	var cfg config

	path, err := configstore.GetItemValue(keyPath)
	if err != nil {
		if _, ok := err.(configstore.ErrItemNotFound); !ok {
			return nil, fmt.Errorf("unable to get the DB path: %w", err)
		}
		cfg.Path = defaultPath
	} else {
		cfg.Path = path
	}

	return &cfg, nil
}
