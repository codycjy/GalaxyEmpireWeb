package taskservice

import (
	"GalaxyEmpireWeb/models"
	"GalaxyEmpireWeb/queue"
	"GalaxyEmpireWeb/utils"
	"context"
	"net/http"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type TaskGenerator struct {
	db          *gorm.DB
	taskChan    chan *models.TaskItem
	messageChan chan *queue.DelayedMessage
}

var taskGenerator *TaskGenerator

func newTaskGenerator(db *gorm.DB, taskChan chan *models.TaskItem, messageChan chan *queue.DelayedMessage) *TaskGenerator {
	return &TaskGenerator{
		db:          db,
		taskChan:    taskChan,
		messageChan: messageChan,
	}
}

func getTaskGenerator(db *gorm.DB, taskChan chan *models.TaskItem, messageChan chan *queue.DelayedMessage) *TaskGenerator {
	if taskGenerator == nil {
		taskGenerator = newTaskGenerator(db, taskChan, messageChan)
	}
	return taskGenerator
}

func setAccountInfo(ctx context.Context, task models.TaskModel, db *gorm.DB) *utils.ServiceError {
	traceID := utils.TraceIDFromContext(ctx)
	var account models.Account
	log.Info("[service]Get account",
		zap.Uint("accountID", task.GetAccountID()),
		zap.String("traceID", traceID),
	)

	result := db.Where("id = ?", task.GetAccountID()).First(&account)
	if err := result.Error; err != nil {
		log.Error("[service]Get account error",
			zap.String("traceID", traceID),
			zap.Error(err),
		)
		return utils.NewServiceError(http.StatusInternalServerError, "[service]Failed to get account", err)

	}
	task.SetAccountInfo(*account.ToInfo())
	return nil
}
func findTasks[T models.TaskModel](ctx context.Context, db *gorm.DB) (items []models.TaskItem) {
	traceID := utils.TraceIDFromContext(ctx)
	log.Info("[service]Find tasks",
		zap.String("traceID", traceID),
	)

	var tasks []T

	cur := db.Where("status = ? AND enabled = ?", models.TaskStatusStop, true).Find(&tasks)
	if cur.Error != nil {
		log.Error("[service]Find tasks error",
			zap.String("traceID", traceID),
			zap.Error(cur.Error),
		)
	}
	for _, task := range tasks {
		setAccountInfo(ctx, task, db)
		taskModel := models.NewTaskItem(task, 0)
		items = append(items, *taskModel)
		if ctx.Done() != nil {
			break
		}
	}
	return items
}

func findTasksByID[T models.TaskModel](ctx context.Context, id uint, db *gorm.DB) (model T, err *utils.ServiceError) {
	traceID := utils.TraceIDFromContext(ctx)
	log.Info("[service]Find tasks",
		zap.String("traceID", traceID),
	)

	cur := db.Where("status = ? AND enabled = ? AND id = ?", models.TaskStatusStop, true, id).Find(&model)
	if cur.Error != nil {
		log.Error("[service]Find tasks error",
			zap.String("traceID", traceID),
			zap.Error(cur.Error),
		)
		return *new(T), utils.NewServiceError(http.StatusInternalServerError, "[service]Failed to find task", cur.Error)
	}
	return model, nil
}
func (generator *TaskGenerator) GenerateTask(ctx context.Context) {
	routeTasks := findTasks[*models.RouteTask](ctx, generator.db)
	for _, task := range routeTasks {
		taskToMQ(context.Background(), task, task.GetDelay(), generator.messageChan)
	}
}
func (generator *TaskGenerator) Start() {
	go generator.manageTaskChan()
	for {
		time.Sleep(15 * time.Second)
		ctx := utils.NewContextWithTraceID()
		ctx, _ = context.WithTimeout(ctx, 5*time.Second)
		log.Info("[service]Generate task",
			zap.String("traceID", utils.TraceIDFromContext(ctx)),
		)
		generator.GenerateTask(ctx)
	}
}
func (generator *TaskGenerator) manageTaskChan() {
	for {
		ctx := utils.NewContextWithTraceID()
		task := <-generator.taskChan
		log.Info("[service]Manage Task Channel",
			zap.String("traceID", utils.TraceIDFromContext(ctx)),
		)

		taskToMQ(ctx, task, task.GetDelay(), generator.messageChan)
	}
}

func taskToMQ(ctx context.Context, task models.TaskModel, delay int64, messageChan chan *queue.DelayedMessage) utils.AppError {
	traceID := utils.TraceIDFromContext(ctx)
	log.Info("[service]Send task",
		zap.Uint("TaskID", task.GetID()),
		zap.String("traceID", traceID),
	)
	taskItem := models.NewTaskItem(task, delay)
	message, err := taskItem.PrepareMessage()
	if err != nil {
		log.Error("[service]Prepare message error",
			zap.String("traceID", traceID),
			zap.Error(err),
		)
		return utils.NewServiceError(http.StatusInternalServerError, "[service]Failed to prepare task message", err)
	}
	messageChan <- queue.NewDelayedMessage(task.RoutingKey(), message, delay)
	log.Info("[service]Send task to chan",
		zap.String("traceID", traceID),
		zap.Uint("TaskID", task.GetID()),
	)
	return nil
}
