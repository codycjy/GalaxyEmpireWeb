package middleware

import (
	"GalaxyEmpireWeb/services/jwtservice"
	"GalaxyEmpireWeb/services/userservice"
	"net/http"

	"github.com/gin-gonic/gin"
)

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
		role := userService.GetUserRole(c, claims.UserID)
		// 设置上下文
		c.Set("claims", claims)
		c.Set("role", role)
		c.Set("userID", claims.UserID)
		c.Next()

	}
}
