package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func UUIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		uuid := uuid.New().String()
		c.Set("UUID", uuid)
		c.Next()
	}

}
