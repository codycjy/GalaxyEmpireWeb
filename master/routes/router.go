package routes

import (
	"GalaxyEmpireWeb/api"
	"GalaxyEmpireWeb/api/account"
	"GalaxyEmpireWeb/api/user"
	"GalaxyEmpireWeb/docs"
	"GalaxyEmpireWeb/middleware"

	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/gin-gonic/gin"
)

func init() {
}

func RegisterRoutes(serviceMap map[string]interface{}) *gin.Engine {
	r := gin.Default()
	r.Use(middleware.TraceIDMiddleware())
	docs.SwaggerInfo.BasePath = "/api/v1"
	v1 := r.Group("/api/v1")
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	v1.GET("/ping", api.Ping)
	u := v1.Group("/user")
	{
		u.GET("/:id", user.GetUser)
		u.POST("", user.CreateUser) // TODO: Move to /register without auth
		u.DELETE("", user.DeleteUser)
		u.PUT("", user.UpdateUser)
	}
	balance := u.Group("/balance")
	{
		balance.PUT("", user.UpdateBalance)
	}

	a := v1.Group("/account")
	{
		a.GET("/:id", account.GetAccountByID)
		a.GET("/user/:userid", account.GetAccountByUserID)
	}

	return r
}
