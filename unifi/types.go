package unifi

import (
	"encoding/json"
	"fmt"
	"net"
)

type result[T any] struct {
	Meta struct {
		Result string `json:"rc"`
	} `json:"meta"`

	Data T `json:"data"`
}

type NetworkConf struct {
	Name       string     `json:"name"`
	Enabled    bool       `json:"enabled"`
	IpSubnet   *net.IPNet `json:"ip_subnet"`
	DomainName string     `json:"domain_name"`
}

func (n *NetworkConf) UnmarshalJSON(data []byte) error {
	type alias NetworkConf

	temp := &struct {
		IpSubnet string `json:"ip_subnet"`
		*alias
	}{
		alias: (*alias)(n),
	}

	if err := json.Unmarshal(data, temp); err != nil {
		return err
	}

	if temp.IpSubnet != "" {
		_, ipnet, err := net.ParseCIDR(temp.IpSubnet)
		if err != nil {
			return fmt.Errorf("invalid CIDR block: %w", err)
		}
		n.IpSubnet = ipnet
	}

	return nil
}

type User struct {
	Name       string           `json:"name"`
	HostName   string           `json:"hostname"`
	UseFixedIP bool             `json:"use_fixedip"`
	FixedIP    net.IP           `json:"fixed_ip"`
	LastIP     net.IP           `json:"last_ip"`
	HwAddress  net.HardwareAddr `json:"mac"`
}

func (u *User) UnmarshalJSON(data []byte) error {
	type alias User

	temp := &struct {
		FixedIP   string `json:"fixed_ip"`
		LastIP    string `json:"last_ip"`
		HwAddress string `json:"mac"`
		*alias
	}{
		alias: (*alias)(u),
	}

	if err := json.Unmarshal(data, temp); err != nil {
		return err
	}

	if temp.FixedIP != "" {
		fixedIP := net.ParseIP(temp.FixedIP)
		if fixedIP == nil {
			return fmt.Errorf("invalid FixedIP address: %s", temp.FixedIP)
		}
		u.FixedIP = fixedIP
	}

	if temp.LastIP != "" {
		lastIP := net.ParseIP(temp.LastIP)
		if lastIP == nil {
			return fmt.Errorf("invalid LastIP address: %s", temp.LastIP)
		}
		u.LastIP = lastIP
	}

	if string(temp.HwAddress) != "" {
		hwAddr, err := net.ParseMAC(temp.HwAddress)
		if err != nil {
			return fmt.Errorf("invalid HwAddress: %w", err)
		}
		u.HwAddress = hwAddr
	}

	return nil
}
