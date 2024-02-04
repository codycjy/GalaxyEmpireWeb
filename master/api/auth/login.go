package auth

import (
	"GalaxyEmpireWeb/api"
	"GalaxyEmpireWeb/models"
	"GalaxyEmpireWeb/services/jwtservice"
	"GalaxyEmpireWeb/services/userservice"
	"GalaxyEmpireWeb/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type authResponse struct {
	Token string `json:"token"`
}

// @Summary User Login
// @Description Authenticate user and generate JWT token
// @Tags auth
// @Accept json
// @Produce json
// @Param user body models.User required "UserID and Password"
// @Success 200 {object} authResponse "Successful response with JWT token"
// @Failure 400 {object} api.ErrorResponse "Bad request with error message"
// @Failure 500 {object} api.ErrorResponse "Internal server error with error message"
// @Router /login [post]
func LoginHandler(c *gin.Context) {
	traceID := utils.TraceIDFromContext(c)
	userService, err := userservice.GetService(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.ErrorResponse{
			Succeed: false,
			Error:   err.Error(),
			Message: "Failed to get user service",
			TraceID: traceID,
		})
	}
	user := &models.User{}
	if err := c.ShouldBindJSON(user); err != nil {
		c.JSON(http.StatusUnauthorized, api.ErrorResponse{
			Succeed: false,
			Error:   err.Error(),
			Message: "Failed to bind json",
			TraceID: traceID,
		})
		return
	}
	err2 := userService.LoginUser(c, user)
	if err2 != nil {
		c.JSON(http.StatusUnauthorized, api.ErrorResponse{
			Succeed: false,
			Error:   err2.Error(),
			Message: "Wrong Username or Password",
			TraceID: traceID,
		})
		return
	}
	token, err3 := jwtservice.GenerateToken(user.ID)
	if err3 != nil {
		c.JSON(http.StatusInternalServerError, api.ErrorResponse{
			Succeed: false,
			Error:   err3.Error(),
			Message: "Failed to generate token",
			TraceID: traceID,
		})
		return
	}
	c.JSON(http.StatusOK, authResponse{
		Token: token,
	})
}
