package db

import (
	"net/netip"
	"regexp"

	"github.com/google/uuid"
	"github.com/juju/errors"
)

func validateID(id string) error {
	_, err := uuid.Parse(id)
	if err != nil {
		return errors.NewBadRequest(err, "invalid ID")
	}
	return nil
}

var (
	validateNameRegexp = regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?$`)
)

func validateName(name string) error {
	if !validateNameRegexp.Match([]byte(name)) {
		return errors.NewBadRequest(nil, "invalid name")
	}
	return nil
}

func validateTarget(target string) error {
	_, err := netip.ParseAddr(target)
	if err != nil {
		return errors.NewBadRequest(err, "invalid target")
	}
	return nil
}
