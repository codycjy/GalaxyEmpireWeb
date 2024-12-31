package routes

import (
	"GalaxyEmpireWeb/api"
	"GalaxyEmpireWeb/api/account"
	"GalaxyEmpireWeb/api/auth"
	"GalaxyEmpireWeb/api/payment"
	"GalaxyEmpireWeb/api/task"
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

func RegisterRoutes() *gin.Engine {
	r := gin.Default()
	r.Use(middleware.TraceIDMiddleware())
	docs.SwaggerInfo.BasePath = "/api/v1"
	v1 := r.Group("/api/v1")
	unprotectedRoutes := v1.Group("/")
	{
		unprotectedRoutes.POST("/webhook", payment.HandleWebhook)
		unprotectedRoutes.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
		unprotectedRoutes.GET("/ping", api.Ping)
		unprotectedRoutes.GET("/captcha", api.GetCaptcha)
		unprotectedRoutes.GET("/captcha/:captchaID", api.GeneratePicture)
		if os.Getenv("ENV") == "test" {
			v1.POST("/login", auth.LoginHandler)
			v1.POST("/register", user.CreateUser)
		} else {
			v1.POST("/login", middleware.CpatchaMiddleware(), auth.LoginHandler)
			v1.POST("/register", middleware.CpatchaMiddleware(), user.CreateUser)
		}
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
		balance.PUT("", user.UpdateBalance) // TODO: use other interface not http
	}

	a := v1.Group("/account")
	{
		a.GET("/:id", account.GetAccountByID)
		a.GET("/user/:userid", account.GetAccountByUserID)
		a.POST("", account.CreateAccount)
		a.DELETE("", account.DeleteAccount)
		a.POST("/check", account.CheckAccountAvailable)
		a.GET("/check/:uuid", account.CheckAccountByUUID)
		a.POST("/extend", account.ExtendAccount)
	}
	t := v1.Group("/task")
	{
		t.GET("/:id", task.GetTaskByID)
		t.GET("/account/:id", task.GetTaskByAccountID)
		t.POST("", task.AddTask)
		t.DELETE("", task.DeleteTask)
		t.PUT("", task.UpdateTask)
	}
	task.RegisterPlanetRoutes(t)

	p := v1.Group("/payment")
	{
		r.POST("/webhook", payment.HandleWebhook) // Doesn't need JWT
		p.POST("/create-checkout", payment.CreateCheckoutSession)
		p.POST("/deposit", payment.CreateDepositSession)
		p.GET("/history", payment.GetPaymentHistory)
		p.GET("/:payment_id", payment.GetPaymentStatus)
	}

	return r
}
