//go:build !test

package account

import (
	"GalaxyEmpireWeb/api"
	"GalaxyEmpireWeb/logger"
	"GalaxyEmpireWeb/models"
	"GalaxyEmpireWeb/services/accountservice"
	"GalaxyEmpireWeb/utils"
	"errors"
	"net/http"
	"net/mail"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type accountResponse struct {
	Succeed bool               `json:"succeed"`
	Data    *models.AccountDTO `json:"data"`
	TraceID string             `json:"traceID"`
}
type userAccountResponse struct {
	Succeed bool            `json:"succeed"`
	Data    *models.UserDTO `json:"data"`
	TraceID string          `json:"traceID"`
}

var log = logger.GetLogger()

// GetAccountByID godoc
// @Summary Get account by ID
// @Description Get Account by ID
// @Tags account
// @Accept json
// @Produce json
// @Param id path int true "Account ID"
// @Success 200 {object} accountResponse "Successful response with account data"
// @Failure 400 {object} api.ErrorResponse "Bad Request with error message"
// @Failure 404 {object} api.ErrorResponse "Not Found with error message"
// @Failure 500 {object} api.ErrorResponse "Internal Server Error with error message"
// @Router /account/{id} [get]
func GetAccountByID(c *gin.Context) {
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
	accountService, err := accountservice.GetService(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.ErrorResponse{
			Succeed: false,
			Error:   err.Error(),
			Message: "Account service not initialized",
			TraceID: traceID,
		})
	}

	account, serviceErr := accountService.GetById(c, uint(id), []string{})
	if serviceErr != nil {
		c.JSON(serviceErr.StatusCode(), api.ErrorResponse{
			Succeed: false,
			Error:   serviceErr.Error(),
			Message: serviceErr.Msg(),
			TraceID: traceID,
		})
		return
	}
	c.JSON(http.StatusOK, accountResponse{
		Succeed: true,
		Data:    account.ToDTO(),
		TraceID: traceID,
	})

}

// GetAccountByUserID godoc
// @Summary Get account by User ID
// @Description Get Account by User ID
// @Tags account
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} userAccountResponse "Successful response with account data"
// @Failure 400 {object} api.ErrorResponse "Bad Request with error message"
// @Failure 404 {object} api.ErrorResponse "Not Found with error message"
// @Failure 500 {object} api.ErrorResponse "Internal Server Error with error message"
// @Router /account/user/{id} [get]
func GetAccountByUserID(c *gin.Context) {
	traceID := c.GetString("traceID")
	idStr := c.Param("userid")
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

	role := c.Value("role").(int)
	ctxUserID := c.Value("userID").(uint)
	if role == 0 {
		if uint(id) != ctxUserID {
			c.JSON(http.StatusForbidden, api.ErrorResponse{
				Succeed: false,
				Error:   "Forbidden",
				Message: "You are not allowed to access this resource",
				TraceID: traceID,
			})
			return
		}
	}

	accountService, _ := accountservice.GetService(c)
	account, serviceErr := accountService.GetByUserId(c, uint(id), []string{})
	if serviceErr != nil {
		c.JSON(serviceErr.StatusCode(), api.ErrorResponse{
			Succeed: false,
			Error:   serviceErr.Error(),
			Message: serviceErr.Msg(),
			TraceID: traceID,
		})
		return
	}
	accountDTOs := make([]models.AccountDTO, len(*account))
	for i, acc := range *account {
		accountDTOs[i] = *acc.ToDTO()
	}
	user := &models.UserDTO{
		ID:       uint(id),
		Accounts: accountDTOs,
	}

	c.JSON(http.StatusOK, userAccountResponse{
		Succeed: true,
		Data:    user,
		TraceID: traceID,
	})
}

// CreateAccount godoc
// @Summary Create Account
// @Description Create Account
// @Tags account
// @Accept json
// @Produce json
// @Param account body models.Account true "Account"
// @Success 200 {object} accountResponse "Successful response with account data"
// @Failure 400 {object} api.ErrorResponse "Bad Request with error message"
// @Failure 500 {object} api.ErrorResponse "Internal Server Error"
// @Router /account [POST]
func CreateAccount(c *gin.Context) {
	traceID := c.GetString("traceID")
	var account models.Account
	err := c.ShouldBindJSON(&account)
	if err != nil {
		log.Error("[api]Create Account failed",
			zap.String("traceID", traceID),
			zap.Error(err),
		)
		c.JSON(http.StatusBadRequest, api.ErrorResponse{
			Succeed: false,
			Error:   err.Error(),
			Message: "Failed to bind json",
			TraceID: traceID,
		})
		return

	}
	err1 := verifyAccount(c, &account)
	if err1 != nil {
		log.Error("[api]Create Account failed",
			zap.String("traceID", traceID),
			zap.Error(err1),
		)
		c.JSON(err1.StatusCode(), api.ErrorResponse{
			Succeed: false,
			Error:   err1.Error(),
			Message: err1.Msg(),
			TraceID: traceID,
		})
		return
	}

	accountService, _ := accountservice.GetService(c)
	serviceErr := accountService.Create(c, &account)
	if serviceErr != nil {
		log.Error("[api]Create Account failed",
			zap.String("traceID", traceID),
			zap.Error(err),
		)
		c.JSON(serviceErr.StatusCode(), api.ErrorResponse{
			Succeed: false,
			Error:   serviceErr.Error(),
			Message: serviceErr.Msg(),
			TraceID: traceID,
		})
		return
	}

	accoutDTO := account.ToDTO()
	c.JSON(http.StatusOK, accountResponse{
		Succeed: true,
		Data:    accoutDTO,
		TraceID: traceID,
	})

}

// DeleteAccount godoc
// @Summary Delete Account
// @Description Delete Account
// @Tags account
// @Accept json
// @Produce json
// @Param account body models.Account true "Account"
// @Success 200 {object} accountResponse "Successful response with account data"
// @Failure 400 {object} api.ErrorResponse "Bad Request with error message"
// @Failure 500 {object} api.ErrorResponse "Internal Server Error"
// @Router /account [Delete]
func DeleteAccount(c *gin.Context) {
	traceID := c.GetString("traceID")
	var account models.Account
	err := c.ShouldBindJSON(&account)
	if err != nil {
		log.Error("[api]Delete Account failed",
			zap.String("traceID", traceID),
			zap.Error(err),
		)
		c.JSON(http.StatusBadRequest, api.ErrorResponse{
			Succeed: false,
			Error:   err.Error(),
			Message: "Bad Request",
			TraceID: traceID,
		})
	}
	accountService, _ := accountservice.GetService(c)
	serviceErr := accountService.Delete(c, account.ID)
	if serviceErr != nil {
		log.Error("[api]Delete Account failed",
			zap.String("traceID", traceID),
			zap.Uint("accountID", account.ID),
		)
		c.JSON(serviceErr.StatusCode(), api.ErrorResponse{
			Succeed: false,
			Error:   serviceErr.Error(),
			Message: serviceErr.Msg(),
			TraceID: traceID,
		})
		return
	}

	c.JSON(http.StatusOK, accountResponse{
		Succeed: true,
		TraceID: traceID,
	})

}

type accountCheckingResponse struct {
	Succeed bool   `json:"succeed"`
	TraceID string `json:"traceID"`
	UUID    string `json:"uuid"`
}

// CheckAccountAvailable godoc
// @Summary Check Account Available
// @Description Check Account Available
// @Tags account
// @Accept json
// @Produce json
// @Param account body models.Account true "Account"
// @Success 200 {object} accountCheckingResponse "Successful response with account data"
// @Failure 400 {object} api.ErrorResponse "Bad Request with error message"
// @Failure 500 {object} api.ErrorResponse "Internal Server Error"
// @Router /account/check [POST]

func CheckAccountAvailable(c *gin.Context) {
	traceID := utils.TraceIDFromContext(c)
	var account models.Account
	err := c.ShouldBindJSON(&account)
	if err != nil {
		log.Error("[api]Check Account Available failed",
			zap.String("traceID", traceID),
			zap.Error(err),
		)
		c.JSON(http.StatusBadRequest, api.ErrorResponse{
			Succeed: false,
			Error:   err.Error(),
			Message: "Bad Request",
			TraceID: traceID,
		})
		return
	}
	accountService, _ := accountservice.GetService(c)
	uuid, serviceErr := accountService.RequestCheckingAccountLogin(c, &account)
	if serviceErr != nil {
		log.Error("[api]Check Account Available failed",
			zap.String("traceID", traceID),
			zap.Error(serviceErr),
		)
		c.JSON(serviceErr.StatusCode(), api.ErrorResponse{
			Succeed: false,
			Error:   serviceErr.Error(),
			Message: serviceErr.Msg(),
			TraceID: traceID,
		})
		return
	}
	c.JSON(http.StatusOK, accountCheckingResponse{
		Succeed: false, // Keep false because the task is not done yet
		TraceID: traceID,
		UUID:    uuid,
	})
}

// CheckAccountByUUID godoc
// @Summary Check Account By UUID
// @Description Check Account By UUID
// @Tags Account
// @Param uuid path string true "UUID"
// @Produce JSON
// @Success 200 {object} accountCheckingResponse "Successful response with account data"
// @Failure 400 {object} api.ErrorResponse "Bad Request with error message"
// @Failure 500 {object} api.ErrorResponse "Internal Server Error"
// @Router /account/check/{uuid} [GET]

func CheckAccountByUUID(c *gin.Context) {
	traceID := utils.TraceIDFromContext(c)
	uuid := c.Param("uuid")
	accountService, _ := accountservice.GetService(c)
	result := accountService.GetLoginInfo(c, uuid)
	c.JSON(http.StatusOK, accountCheckingResponse{
		Succeed: result,
		TraceID: traceID,
		UUID:    uuid,
	})

}
func verifyAccount(c *gin.Context, account *models.Account) *utils.ApiError {
	traceID := c.GetString("traceID")
	log.Info("[api]Verify Account",
		zap.String("traceID", traceID),
		zap.String("username", account.Username),
		zap.String("email", account.Email),
	)
	if account.Username == "" {
		return utils.NewApiError(http.StatusBadRequest, "Username is required", errors.New("Username is required"))
	}
	if account.Password == "" {
		return utils.NewApiError(http.StatusBadRequest, "Password is required", errors.New("Password is required"))
	}
	if account.Email == "" {
		return utils.NewApiError(http.StatusBadRequest, "Email is required", errors.New("Email is required"))
	}

	_, err := mail.ParseAddress(account.Email)
	if err != nil {
		return utils.NewApiError(http.StatusBadRequest, "Invalid Email", err)
	}
	log.Info("[api]Verify Account - Succeed", zap.String("traceID", traceID))

	return nil
}
