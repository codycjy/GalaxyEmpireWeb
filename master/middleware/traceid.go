package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// TraceIDMiddleware godoc
// Create middleware to add traceID
func TraceIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		uuid := uuid.New().String()
		c.Set("traceID", uuid)
		c.Next()
	}

}
