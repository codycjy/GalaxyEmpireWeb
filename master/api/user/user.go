package user

import (
	"GalaxyEmpireWeb/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type UserResponse struct {
	Succeed bool        `json:"succeed"`
	Data    models.User `json:"data"`
}
type UsersResponse struct {
	Succeed bool          `json:"succeed"`
	Data    []models.User `json:"data"`
}

// GetUser godoc
// @Summary Get user by ID
// @Description Get User by ID
// @Tags user
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} UserResponse "Successful response with user data"
// @Failure 400 {object} map[string]interface{} "Bad Request with error message"
// @Failure 500 {object} map[string]interface{} "Internal Server Error with error message"
// @Router /user/{id} [get]
func GetUser(c *gin.Context) {
	var user models.User
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"succeed": false,
			"error":   err.Error(),
		})
		return
	}
	err = user.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, UserResponse{
			Succeed: false,
		})
		return
	}
	c.JSON(http.StatusOK, UserResponse{
		Succeed: true,
		Data:    user,
	})
}

// GetUser godoc
// @Summary Get all users
// @Description Get all Users
// @Tags user
// @Accept json
// @Produce json
// @Success 200 {object} UsersResponse "Successful response with user data"
// @Failure 400 {object} map[string]interface{} "Bad Request with error message"
// @Failure 500 {object} map[string]interface{} "Internal Server Error with error message"
// @Router /users [get]
func GetUsers(c *gin.Context) {
	var users []models.User
	err := models.GetAllUsers(&users)

	if err != nil {
		c.JSON(http.StatusNotFound, UsersResponse{
			Succeed: false,
		})
		return
	}
	c.JSON(http.StatusOK, UsersResponse{
		Succeed: true,
		Data:    users,
	})
}

// GetUser godoc
// @Summary Crea user
// @Description Create User
// @Tags user
// @Accept json
// @Produce json
// @Param user body models.User required "User ID or Username"
// @Success 200 {object} UserResponse "Successful response with user data"
// @Failure 400 {object} map[string]interface{} "Bad Request with error message"
// @Failure 500 {object} map[string]interface{} "Internal Server Error with error message"
// @Router /user [post]
func CreateUser(c *gin.Context) {
	var user models.User
	err := c.BindJSON(&user)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"succeed": false,
			"message": "Bad request",
			"error":   err,
		})
		return
	}
	err = user.Create()

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"succeed": false,
			"message": "Internal server error",
			"error":   err,
		})
		return
	}
	c.JSON(http.StatusOK, UserResponse{
		Succeed: true,
		Data:    user,
	})

}

// UpdateUser godoc
// @Summary Update user
// @Description Update User
// @Tags user
// @Accept json
// @Produce json
// @Param user body models.User required "User ID or Username"
// @Success 200 {object} UserResponse "Successful response with user data"
// @Failure 400 {object} map[string]interface{} "Bad Request with error message"
// @Failure 500 {object} map[string]interface{} "Internal Server Error with error message"
// @Router /user [put]
func UpdateUser(c *gin.Context) {
	var user models.User
	err := c.BindJSON(&user)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"succeed": false,
			"message": "Bad request",
			"error":   err,
		})
		return
	}
	err = user.Update()

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"succeed": false,
			"message": "Internal server error",
			"error":   err,
		})
		return
	}
	c.JSON(http.StatusOK, UserResponse{
		Succeed: true,
		Data:    user,
	})

}

// DeleteUser godoc
// @Summary Delete user
// @Description Delete User
// @Tags user
// @Accept json
// @Produce json
// @Param user body models.User required "User ID or Username"
// @Success 200 {object} UserResponse "Successful response with user data"
// @Failure 400 {object} map[string]interface{} "Bad Request with error message"
// @Failure 500 {object} map[string]interface{} "Internal Server Error with error message"
// @Router /user [delete]
func DeleteUser(c *gin.Context) {
	var user models.User
	err := c.BindJSON(&user)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"succeed": false,
			"message": "Bad request",
			"error":   err,
		})
		return
	}
	err = user.Delete()

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"succeed": false,
			"message": "Internal server error",
			"error":   err,
		})
		return
	}
	c.JSON(http.StatusOK, UserResponse{
		Succeed: true,
		Data:    user,
	})
}
