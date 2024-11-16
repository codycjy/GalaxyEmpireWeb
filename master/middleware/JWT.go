package middleware

import (
	"GalaxyEmpireWeb/logger"
	"GalaxyEmpireWeb/services/jwtservice"
	"GalaxyEmpireWeb/services/userservice"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)
var log = logger.GetLogger()

func JWTAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is missing"})
			return
		}
		claims, err := jwtservice.ParseToken(authHeader)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		userService, err := userservice.GetService(c)
		log.Info("[middleware]JWTAuthMiddleware", zap.Uint("UserID", claims.UserID))

		role := userService.GetUserRole(c, claims.UserID)
		// 设置上下文
		c.Set("claims", claims)
		c.Set("role", role)
		c.Set("userID", claims.UserID)
		log.Info("[middleware]JWTAuthMiddleware", zap.String("traceID", c.GetString("traceID")), zap.Int("role", role))
		c.Next()

	}
}
