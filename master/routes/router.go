package routes

import (
	"GalaxyEmpireWeb/api"

	"github.com/gin-gonic/gin"
)

func init() {
	RegisterRoutes()
}

func RegisterRoutes() {
	r := gin.Default()
	v1 := r.Group("/api/v1")
	v1.GET("/ping", api.Ping)
}
