package models

import (
	"log"
	"os"
)

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
func (planTask *PlanTask) SetAccountInfo(account AccountInfo) {
	// TODO: implement
	log.Fatal()
}
func (planTask *PlanTask) GetAccountID() uint {
	// TODO: implement
	log.Fatal()
	return 0
}
func (planTask *PlanTask) GetID() uint {
	// TODO: implement
	log.Fatal()
	return 0
}
