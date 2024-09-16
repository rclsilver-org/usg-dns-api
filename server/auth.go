package server

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/rclsilver-org/usg-dns-api/pkg/utils"
)

func (s *Server) AuthMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token := ctx.GetHeader("Authorization")

		if token == "" {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Token required"})
			ctx.Abort()
			return
		}

		if utils.StringHash(token) != s.db.GetMasterToken() {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			ctx.Abort()
			return
		}

		ctx.Next()
	}
}
