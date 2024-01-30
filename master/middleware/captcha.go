package middleware

import (
	"GalaxyEmpireWeb/services/captchaservice"
	"net/http"

	"github.com/gin-gonic/gin"
)

func CpatchaMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		captchaId := c.GetHeader("captchaId")
		userInput := c.GetHeader("userInput")
		if captchaId == "" || userInput == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "captchaId or userInput is empty"})
			c.Abort()
			return
		}
		capchaService := captchaservice.GetCaptchaService()
		if !capchaService.VerifyCaptcha(c, captchaId, userInput) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":     "captcha is error",
				"captchaId": captchaId,
				"userInput": userInput,
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
