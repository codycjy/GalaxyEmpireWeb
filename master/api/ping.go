package api

import "github.com/gin-gonic/gin"

type responseMessage struct {
	Message string `json:"message"`
}

// @BasePath /api/v1

// Ping godoc
// @Summary ping example
// @Schemes
// @Description do ping
// @Tags example
// @Accept json
// @Produce json
// @Success 200 {object} responseMessage
// @Router /ping [get]
func Ping(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "pong",
	})
}
