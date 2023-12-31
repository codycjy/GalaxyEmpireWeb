//go:build !test

package user

import (
	"GalaxyEmpireWeb/api"
	"GalaxyEmpireWeb/logger"
	"GalaxyEmpireWeb/models"
	"GalaxyEmpireWeb/services/userservice"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type userResponse struct {
	Succeed bool            `json:"succeed"`
	Data    *models.UserDTO `json:"data"`
}
type usersResponse struct {
	Succeed bool             `json:"succeed"`
	Data    []models.UserDTO `json:"data"`
}

var log = logger.GetLogger()

// GetUser godoc
// @Summary Get user by ID
// @Description Get User by ID
// @Tags user
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} userResponse "Successful response with user data"
// @Failure 400 {object} api.ErrorResponse "Bad Request with error message"
// @Failure 500 {object} api.ErrorResponse "Internal Server Error with error message"
// @Router /user/{id} [get]
func GetUser(c *gin.Context) {
	traceID := c.GetString("traceID")
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.ErrorResponse{
			Succeed: false,
			Error:   err.Error(),
			Message: "Wrong User ID",
			TraceID: traceID,
		})
		return
	}
	userService, err := userservice.GetService(c) //TODO: remove error
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.ErrorResponse{
			Succeed: false,
			Error:   err.Error(),
			Message: "User service not initialized",
			TraceID: traceID,
		})
	}
	user, err := userService.GetById(c, uint(id), []string{})
	if err != nil {
		c.JSON(http.StatusNotFound, api.ErrorResponse{
			Succeed: false,
			Error:   err.Error(),
			Message: "User not found",
			TraceID: traceID,
		})
		return
	}
	c.JSON(http.StatusOK, userResponse{
		Succeed: true,
		Data:    user.ToDTO(),
	})
}

// GetUsers godoc
// @Summary Get all users
// @Description Get all Users
// @Tags user
// @Accept json
// @Produce json
// @Success 200 {object} usersResponse "Successful response with user data"
// @Failure 400 {object} api.ErrorResponse "Bad Request with error message"
// @Failure 500 {object} api.ErrorResponse "Internal Server Error with error message"
// @Router /users [get]
func GetUsers(c *gin.Context) {
	traceID := c.GetString("traceID")
	userService, err := userservice.GetService(c)
	if err != nil {
		log.Error("[api]User service not initialized",
			zap.String("traceID", traceID),
		)
		c.JSON(http.StatusInternalServerError, api.ErrorResponse{
			Succeed: false,
			Error:   err.Error(),
			Message: "User service not initialized",
			TraceID: traceID,
		})
	}
	users, err := userService.GetAllUsers(c)
	usersDTO := make([]models.UserDTO, len(users))
	for _, user := range users {
		usersDTO = append(usersDTO, *user.ToDTO())
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, api.ErrorResponse{
			Succeed: false,
			Error:   err.Error(),
			Message: "Failed to get users",
			TraceID: traceID,
		})
		return
	}
	c.JSON(http.StatusOK, usersResponse{
		Succeed: true,
		Data:    usersDTO,
	})
}

// CreateUser godoc
// @Summary Crea user
// @Description Create User
// @Tags user
// @Accept json
// @Produce json
// @Param user body models.User required "User ID or Username"
// @Success 200 {object} userResponse "Successful response with user data"
// @Failure 400 {object} api.ErrorResponse "Bad Request with error message"
// @Failure 500 {object} api.ErrorResponse "Internal Server Error with error message"
// @Router /user [post]
func CreateUser(c *gin.Context) {
	traceID := c.GetString("traceID")
	var user *models.User
	err := c.BindJSON(user)
	if err != nil {

		log.Error("[api]Failed to bind to user",
			zap.String("traceID", traceID),
		)

		c.JSON(http.StatusBadRequest, api.ErrorResponse{
			Succeed: false,
			Error:   err.Error(),
			Message: "Failed to bind to user",
			TraceID: traceID,
		})
		return
	}
	userService, err := userservice.GetService(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.ErrorResponse{
			Succeed: false,
			Error:   err.Error(),
			Message: "User service not initialized",
			TraceID: traceID,
		})
	}
	err = userService.Create(c, user)

	if err != nil {
		c.JSON(http.StatusInternalServerError, api.ErrorResponse{
			Succeed: false,
			Error:   err.Error(),
			Message: "Create user failed",
			TraceID: traceID,
		})
		return
	}
	c.JSON(http.StatusOK, userResponse{
		Succeed: true,
		Data:    user.ToDTO(),
	})

}

// UpdateUser godoc
// @Summary Update user
// @Description Update User
// @Tags user
// @Accept json
// @Produce json
// @Param user body models.User required "User ID or Username"
// @Success 200 {object} userResponse "Successful response with user data"
// @Failure 400 {object} api.ErrorResponse "Bad Request with error message"
// @Failure 500 {object} api.ErrorResponse "Internal Server Error with error message"
// @Router /user [put]
func UpdateUser(c *gin.Context) {
	traceID := c.GetString("traceID")

	var user *models.User
	err := c.BindJSON(&user)
	if err != nil {
		log.Error("[api]Failed to bind to user",
			zap.String("traceID", traceID),
		)
		c.JSON(http.StatusBadRequest, api.ErrorResponse{
			Succeed: false,
			Error:   err.Error(),
			Message: "Failed to bind to user",
			TraceID: traceID,
		})
		return
	}
	userService, err := userservice.GetService(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.ErrorResponse{
			Succeed: false,
			Error:   err.Error(),
			Message: "User service not initialized",
			TraceID: traceID,
		})
	}
	err = userService.Update(c, user)

	if err != nil {
		c.JSON(http.StatusInternalServerError, api.ErrorResponse{
			Succeed: false,
			Error:   err.Error(),
			Message: "Update user failed",
			TraceID: traceID,
		})
		return
	}
	c.JSON(http.StatusOK, userResponse{
		Succeed: true,
		Data:    user.ToDTO(),
	})

}

// DeleteUser godoc
// @Summary Delete user
// @Description Delete User
// @Tags user
// @Accept json
// @Produce json
// @Param user body models.User required "User ID or Username"
// @Success 200 {object} userResponse "Successful response with user data"
// @Failure 400 {object} api.ErrorResponse "Bad Request with error message"
// @Failure 500 {object} api.ErrorResponse "Internal Server Error with error message"
// @Router /user [delete]
func DeleteUser(c *gin.Context) {
	traceID := c.GetString("traceID")

	var user *models.User
	err := c.BindJSON(&user)

	if err != nil {
		log.Error("[api]Failed to bind to user",
			zap.String("traceID", traceID),
		)
		c.JSON(http.StatusBadRequest, api.ErrorResponse{
			Succeed: false,
			Error:   err.Error(),
			Message: "Failed to bind to user",
			TraceID: traceID,
		})
		return
	}

	userService, err := userservice.GetService(c)
	err = userService.Delete(c, user.ID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, api.ErrorResponse{
			Succeed: false,
			Error:   err.Error(),
			Message: "Delete user failed",
			TraceID: traceID,
		})
		return
	}

	c.JSON(http.StatusOK, userResponse{
		Succeed: true,
	})
}
