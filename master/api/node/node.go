package node

import (
	"GalaxyEmpireWeb/api"
	"GalaxyEmpireWeb/logger"
	"GalaxyEmpireWeb/services/nodeservice"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var log = logger.GetLogger()

type nodePing struct {
	NodeName string `json:"node_name"`
}
type nodePong struct {
	Succeed bool
	Message string
	TraceID string
}

func NodePing(c *gin.Context) {
	traceID := c.GetString("traceID")

	var ping nodePing
	err := c.ShouldBindJSON(&ping)
	if err != nil {
		log.Error(
			"Bind json error",
			zap.String("traceID", traceID),
		)
		c.JSON(http.StatusBadRequest, api.ErrorResponse{
			Succeed: false,
			Message: "Bind json error",
			Error:   err.Error(),
			TraceID: traceID,
		})
	}

	nodeService := nodeservice.GetService()
	err = nodeService.RegisterNode(c, ping.NodeName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.ErrorResponse{
			Succeed: false,
			Message: "Register node error",
			Error:   err.Error(),
			TraceID: traceID,
		})
	}

	c.JSON(http.StatusOK, nodePong{
		Message: "pong",
		Succeed: true,
		TraceID: traceID,
	})

}
