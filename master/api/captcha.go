package api

import (
	"GalaxyEmpireWeb/services/captchaservice"
	"GalaxyEmpireWeb/utils"
	"net/http"

	"github.com/dchest/captcha"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type captchaResponse struct {
	Succeed   bool   `json:"succeed"`
	CaptchaID string `json:"captcha_id"`
	TraceID   string `json:"traceID"`
}

// @BasePath /api/v1

// GetCaptcha godoc
// @Summary Get captcha
// @Description Get captcha
// @Tags Captcha
// @Accept json
// @Produce json
// @Success 200 {object} captchaResponse
// @Router /captcha [get]
func GetCaptcha(c *gin.Context) {
	traceID := utils.TraceIDFromContext(c)
	log.Info("[api]GetCaptcha", zap.String("traceID", traceID))
	captchaService := captchaservice.GetCaptchaService()
	captchaID := captchaService.GenerateCaptcha()
	c.JSON(http.StatusOK, captchaResponse{
		Succeed:   true,
		CaptchaID: captchaID,
		TraceID:   traceID,
	})

}

// GeneratePicture godoc
// @Summary Generate captcha picture
// @Description Generate captcha picture
// @Tags Captcha
// @Accept json
// @Produce image/png
// @Produce json
// @Param captchaID path string true "captchaID"
// @Success 200 {file} file "A captcha image is returned on success"
// @Failure 500 {object} ErrorResponse "If an error occurs, a JSON with error details is returned"
// @/GeneratePicture/{captchaID} [get]
func GeneratePicture(c *gin.Context) {
	captchaID := c.Param("captchaID")
	traceID := utils.TraceIDFromContext(c)
	c.Header("Content-Type", "image/png")

	if err := captcha.WriteImage(c.Writer, captchaID, captcha.StdWidth, captcha.StdHeight); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Succeed: false,
			TraceID: traceID,
			Error:   err.Error(),
			Message: "failed to write captcha image",
		})
		log.Error("[api]GeneratePicture - Failed",
			zap.String("traceID", traceID),
			zap.String("captchaID", captchaID),
			zap.Error(err),
		)
		return
	}
	log.Info("[api]GeneratePicture - Succeed", zap.String("traceID", traceID))
}
