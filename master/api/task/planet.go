package task

import (
	"GalaxyEmpireWeb/api"
	"GalaxyEmpireWeb/models"
	"GalaxyEmpireWeb/services/taskservice"
	"GalaxyEmpireWeb/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// QueryPlanetIDRequest represents the request body for querying a planet ID
type QueryPlanetIDRequest struct {
	Target  models.Target  `json:"target"`
	Account models.Account `json:"account"`
}

// QueryPlanetIDResponse represents the response for the planet ID query
type QueryPlanetIDResponse struct {
	UUID    string `json:"uuid"`
	TraceID string `json:"trace_id"`
}

// GetPlanetIDResponse represents the response containing the planet ID
type GetPlanetIDResponse struct {
	PlanetID int    `json:"planet_id"`
	TraceID  string `json:"trace_id"`
	Succeed  bool   `json:"succeed"`
}

// QueryPlanetID initiates a planet ID query task
// @Summary Query planet ID
// @Description Start a task to query a planet's ID
// @Tags Task
// @Accept json
// @Produce json
// @Param request body QueryPlanetIDRequest true "Query parameters"
// @Success 200 {object} QueryPlanetIDResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/task/planet/query [post]
func QueryPlanetID(c *gin.Context) {
	var req QueryPlanetIDRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Error("[API::QueryPlanetID] invalid request", zap.Error(err))
		c.JSON(http.StatusBadRequest, api.ErrorResponse{
			Succeed: false,
			Error:   err.Error(),
			Message: "Invalid Request",
			TraceID: utils.TraceIDFromContext(c),
		})
		return
	}

	target := &req.Target
	account := &req.Account
	// TODO: validate target and account
	uuid, err := taskservice.GetService().QueryPlanetID(c, target, account)
	if err != nil {
		log.Error("[API::QueryPlanetID] failed to query planet ID", zap.Error(err))
		c.JSON(err.StatusCode(), api.ErrorResponse{
			Succeed: false,
			Error:   err.Error(),
			Message: err.Msg(),
			TraceID: utils.TraceIDFromContext(c),
		})
		return
	}

	c.JSON(http.StatusOK, QueryPlanetIDResponse{UUID: uuid, TraceID: utils.TraceIDFromContext(c)})
}

// GetPlanetID retrieves the planet ID from a previous query
// @Summary Get planet ID
// @Description Get the planet ID from a previous query task
// @Tags Task
// @Accept json
// @Produce json
// @Param uuid path string true "Task UUID"
// @Success 200 {object} GetPlanetIDResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/task/planet/{uuid} [get]
func GetPlanetID(c *gin.Context) {
	uuid := c.Param("uuid")
	if uuid == "" {
		c.JSON(http.StatusBadRequest, api.ErrorResponse{
			Succeed: false,
			Error:   "Missing UUID",
			Message: "Missing UUID",
			TraceID: utils.TraceIDFromContext(c),
		})
		return
	}

	planetID, err := taskservice.GetService().GetPlanetID(c.Request.Context(), uuid)
	if err != nil {
		log.Error("[API::GetPlanetID] failed to get planet ID", zap.Error(err))
		c.JSON(err.StatusCode(), api.ErrorResponse{
			Succeed: false,
			Error:   err.Error(),
			Message: err.Msg(),
			TraceID: utils.TraceIDFromContext(c),
		})
		return
	}

	c.JSON(http.StatusOK, GetPlanetIDResponse{PlanetID: planetID,
		TraceID: utils.TraceIDFromContext(c), Succeed: true})
}

// RegisterPlanetRoutes registers the planet-related routes
func RegisterPlanetRoutes(r *gin.RouterGroup) {
	planet := r.Group("/planet")
	{
		planet.POST("/query", QueryPlanetID)
		planet.GET("/:uuid", GetPlanetID)
	}
}
