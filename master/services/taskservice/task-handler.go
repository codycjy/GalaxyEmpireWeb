package taskservice

import (
	"GalaxyEmpireWeb/consts"
	"GalaxyEmpireWeb/models"
	"GalaxyEmpireWeb/utils"
	"context"
	"encoding/json"
	"fmt"

	r "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// TaskHandler handle response mq
type TaskHandler struct {
	rdb         *r.Client
	db          *gorm.DB
	messageChan chan []byte
}

func NewTaskHandler(rdb *r.Client, db *gorm.DB, messageChan chan []byte) *TaskHandler // TODO:

func (taskHandler TaskHandler) HandleResponse() {
	for msg := range taskHandler.messageChan {
		var response models.TaskResponse
		err := json.Unmarshal(msg, &response)
		if err != nil {
			log.Error("[service] Task Handler - fail to unmarshal response",
				zap.Error(err),
			)
			continue
		}
		ctx := utils.NewContextWithTraceID()
		taskHandler.handle(ctx, response)

	}
}

func (taskHandler TaskHandler) handle(ctx context.Context, response models.TaskResponse) {
	if response.Success != true {
		log.Warn("[service]task failed",
			zap.Int("TaskID", response.TaskID),
			zap.String("ID", response.ID.String()),
		)
	}

	taskCountKey:=fmt.Sprintf("%s%d",consts.TaskCountPrefix,response.TaskID)
	taskHandler.rdb.Decr(ctx,taskCountKey)
	// TODO: send new task and delay to generate chan 

}
