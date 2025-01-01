package middleware

import (
	"GalaxyEmpireWeb/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

func AdminAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		role := utils.GetRoleFromContext(c)
		if role != 1 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"succeed": false,
				"message": "Admin access required",
				"traceID": c.GetString("traceID"),
			})
			return
		}
		c.Next()
	}
}
