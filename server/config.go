package server

import (
	"fmt"

	"github.com/ovh/configstore"
)

const (
	keyListenHost = "HTTP_LISTEN_HOST"
	keyListenPort = "HTTP_LISTEN_PORT"
	keyHostsFile  = "HOSTS_FILE"

	defaultListenHost = "localhost"
	defaultListenPort = 8080
	defaultHostsFile  = "hosts"
)

type config struct {
	ListenHost string
	ListenPort int

	HostsFile string

	Title   string
	Version string

	Verbose bool
}

func loadConfig() (*config, error) {
	var cfg config

	host, err := configstore.GetItemValue(keyListenHost)
	if err != nil {
		if _, ok := err.(configstore.ErrItemNotFound); !ok {
			return nil, fmt.Errorf("unable to get the HTTP listen host: %w", err)
		}
		cfg.ListenHost = defaultListenHost
	} else {
		cfg.ListenHost = host
	}

	port, err := configstore.GetItemValueInt(keyListenPort)
	if err != nil {
		if _, ok := err.(configstore.ErrItemNotFound); !ok {
			return nil, fmt.Errorf("unable to get the HTTP listen port: %w", err)
		}
		cfg.ListenPort = defaultListenPort
	} else {
		cfg.ListenPort = int(port)
	}

	hostsFile, err := configstore.GetItemValue(keyHostsFile)
	if err != nil {
		if _, ok := err.(configstore.ErrItemNotFound); !ok {
			return nil, fmt.Errorf("unable to get the hosts file path: %w", err)
		}
		cfg.HostsFile = defaultHostsFile
	} else {
		cfg.HostsFile = hostsFile
	}

	return &cfg, nil
}
