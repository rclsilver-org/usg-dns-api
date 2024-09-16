package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/juju/errors"
	"github.com/loopfz/gadgeto/tonic"
)

type APIError struct {
	Message string `json:"message"`
}

func parseError(err error) (int, string) {
	if _, ok := err.(tonic.BindError); ok {
		return http.StatusBadRequest, err.Error()
	} else {
		switch {
		case errors.Is(err, errors.BadRequest), errors.Is(err, errors.NotValid), errors.Is(err, errors.NotSupported), errors.Is(err, errors.NotProvisioned):
			return http.StatusBadRequest, err.Error()

		case errors.Is(err, errors.Forbidden):
			return http.StatusForbidden, err.Error()

		case errors.Is(err, errors.MethodNotAllowed):
			return http.StatusMethodNotAllowed, err.Error()

		case errors.Is(err, errors.NotFound), errors.Is(err, errors.UserNotFound):
			return http.StatusNotFound, err.Error()

		case errors.Is(err, errors.Unauthorized):
			return http.StatusUnauthorized, err.Error()

		case errors.Is(err, errors.AlreadyExists):
			return http.StatusConflict, err.Error()

		case errors.Is(err, errors.NotImplemented):
			return http.StatusNotImplemented, err.Error()
		}

		return http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError)
	}
}

func errorHook(ctx *gin.Context, e error) (int, interface{}) {
	code, msg := parseError(e)

	err := APIError{
		Message: msg,
	}

	return code, err
}
