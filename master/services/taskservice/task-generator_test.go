package taskservice

import (
	"GalaxyEmpireWeb/models"
	"GalaxyEmpireWeb/queue"
	"GalaxyEmpireWeb/services/casbinservice"
	"reflect"
	"testing"
	"time"

	"gorm.io/gorm"
)

func Test_taskService_GenerateSingleTask(t *testing.T) {
	type fields struct {
		DB       *gorm.DB
		MQ       *queue.RabbitMQConnection
		Enforcer casbinservice.Enforcer
	}
	type args struct {
		task *models.Task
	}
	NormalTask := models.Task{
		Name:      "test",
		NextStart: time.Now(),
		Enabled:   true,
		AccountID: 1,
		TaskType:  1,
		Status:    "ready",
		Targets: []models.Target{
			{
				Galaxy: 1,
				System: 1,
				Planet: 1,
			},
			{
				Galaxy: 1,
				System: 1,
				Planet: 2,
			},
		},
		Repeat:    1,
		NextIndex: 1,
		TargetNum: 2,
	}
	DisableTask := models.Task{
		Name:      "test",
		NextStart: time.Now(),
		Enabled:   false,
		AccountID: 1,
		TaskType:  1,
		Status:    "ready",
		Targets: []models.Target{
			{
				Galaxy: 1,
				System: 1,
				Planet: 1,
			},
		},
		Repeat:    1,
		NextIndex: 1,
		TargetNum: 1,
	}
	DisableTask2 := models.Task{

		Name:      "test",
		NextStart: time.Now().Add(2 * time.Hour),
		Enabled:   true,
		AccountID: 1,
		TaskType:  1,
		Status:    "ready",
		Targets: []models.Target{
			{
				Galaxy: 1,
				System: 1,
				Planet: 1,
			}},
		Repeat:    1,
		NextIndex: 1,
		TargetNum: 1,
	}
	DisableTask3 := models.Task{
		Name:      "test",
		NextStart: time.Now(),
		Enabled:   true,
		AccountID: 1,
		TaskType:  1,
		Status:    models.TaskStatusMap[models.TASK_STATUS_RUNNING],
		Targets: []models.Target{
			{
				Galaxy: 1,
				System: 1,
				Planet: 1,
			},
		},
		Repeat:    1,
		NextIndex: 1,
		TargetNum: 1,
	}

	ErrorIndexTask := models.Task{
		Name:      "test",
		NextStart: time.Now(),
		Enabled:   true,
		AccountID: 1,
		TaskType:  1,
		Status:    "ready",
		Targets: []models.Target{
			{
				Galaxy: 1,
				System: 1,
				Planet: 1,
			}},
		Repeat:    1,
		NextIndex: 2,
		TargetNum: 1,
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		want   *models.SingleTaskRequest
	}{
		{
			name: "Normal Task",
			fields: fields{
				DB:       nil,
				MQ:       nil,
				Enforcer: nil,
			},
			args: args{
				task: &NormalTask,
			},
			want: &models.SingleTaskRequest{
				TaskID:    NormalTask.ID,
				Name:      NormalTask.Name,
				NextStart: NormalTask.NextStart,
				Enabled:   NormalTask.Enabled,
				Account:   models.AccountInfo{},
				TaskType:  NormalTask.TaskType,
				Target:    NormalTask.Targets[NormalTask.NextIndex],
				Repeat:    NormalTask.Repeat,
				Fleet:     models.Fleet{},
			},
		},
		{
			name: "Disable Task, Enabled = false",
			fields: fields{
				DB:       nil,
				MQ:       nil,
				Enforcer: nil,
			},
			args: args{
				task: &DisableTask,
			},
			want: nil,
		},
		{
			name: "Disable Task, NextStart > time.Now().Add(time.Hour)",
			fields: fields{
				DB:       nil,
				MQ:       nil,
				Enforcer: nil,
			},
			args: args{
				task: &DisableTask2,
			},
			want: nil,
		},
		{
			name: "Disable Task, Status = Running",
			fields: fields{
				DB:       nil,
				MQ:       nil,
				Enforcer: nil,
			},
			args: args{
				task: &DisableTask3,
			},
			want: nil,
		},
		{
			name: "Error Task, NextIndex out of range",
			fields: fields{
				DB:       nil,
				MQ:       nil,
				Enforcer: nil,
			},
			args: args{
				task: &ErrorIndexTask,
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := &taskService{
				DB:       tt.fields.DB,
				MQ:       tt.fields.MQ,
				Enforcer: tt.fields.Enforcer,
			}
			if got := ts.GenerateSingleTask(tt.args.task); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("taskService.GenerateSingleTask() = %v, want %v", got, tt.want)
			}
		})
	}
}
