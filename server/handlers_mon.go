package server

import (
	"github.com/gin-gonic/gin"
)

type pingOutStatus string

const (
	pingOutStatusOK pingOutStatus = "OK"
)

type pingOut struct {
	Status pingOutStatus `json:"status"`
}

func (s *Server) monPing(c *gin.Context) (*pingOut, error) {
	return &pingOut{
		Status: pingOutStatusOK,
	}, nil
}
