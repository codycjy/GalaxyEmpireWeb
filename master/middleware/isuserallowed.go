package middleware

import (
	"GalaxyEmpireWeb/api"
	"GalaxyEmpireWeb/services/userservice"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func AllowedMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		traceID, _ := c.Get("traceID")
		userIDStr := c.Param("id")
		userID, _ := strconv.Atoi(userIDStr)
		userservice, err1 := userservice.GetService(c)
		if err1 != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, api.ErrorResponse{
				Succeed: false,
				Error:   err1.Error(),
				Message: "Get User Service Error",
				TraceID: traceID.(string),
			})
			return
		}
		result, _ := userservice.IsUserAllowed(c, uint(userID))
		if !result {
			c.AbortWithStatusJSON(http.StatusUnauthorized, api.ErrorResponse{
				Succeed: false,
				Message: "User Not allowed",
				TraceID: traceID.(string),
			})
			return
		}
		c.Next()
	}
}
