package routes

import (
	"GalaxyEmpireWeb/api"
	"GalaxyEmpireWeb/api/account"
	"GalaxyEmpireWeb/api/auth"
	"GalaxyEmpireWeb/api/user"
	"GalaxyEmpireWeb/docs"
	"GalaxyEmpireWeb/middleware"
	"os"

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
	v1.GET("/captcha", api.GetCaptcha)
	v1.GET("/captcha/:captchaID", api.GeneratePicture)
	if os.Getenv("ENV") == "test" || os.Getenv("ESCAPE_CAPTCHA") != "" {
		v1.POST("/login", auth.LoginHandler)
		v1.POST("/register", user.CreateUser)
	} else {
		v1.POST("/login", middleware.CpatchaMiddleware(), auth.LoginHandler)
		v1.POST("/register", middleware.CpatchaMiddleware(), user.CreateUser)
	}
	v1.Use(middleware.JWTAuthMiddleware())
	u := v1.Group("/user")
	{
		u.GET("/:id", user.GetUser)
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
		a.POST("", account.CreateAccount)
		a.DELETE("", account.DeleteAccount)
	}

	return r
}
