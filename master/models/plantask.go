package models

import "os"

var PlanTaskName string = "PlanTask"

type PlanTask struct {
	BaseTask
	AccountID uint
}

func (t *PlanTask) QueueName() string {
	if os.Getenv("ENV") == "test" {
		return "Test_" + "HPQueue"
	}
	return "HPQueue"
}
func (t *PlanTask) TaskType() string {
	return "DailyTask"
}
