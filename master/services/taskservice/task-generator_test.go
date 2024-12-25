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
		NextStart: time.Now().Unix(),
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
		NextStart: time.Now().Unix(),
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
		NextStart: time.Now().Add(2 * time.Hour).Unix(),
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
		NextStart: time.Now().Unix(),
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
		NextStart: time.Now().Unix(),
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

	testAccount := &models.Account{
		Model: gorm.Model{ID: 1},
		// Add minimal account details needed for testing
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		account *models.Account
		want    *models.SingleTaskRequest
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
			account: testAccount,
			want: &models.SingleTaskRequest{
				TaskID:    NormalTask.ID,
				Name:      NormalTask.Name,
				NextStart: NormalTask.NextStart,
				Enabled:   NormalTask.Enabled,
				Account:   *testAccount.ToInfo(),
				TaskType:  NormalTask.TaskType,
				Target:    NormalTask.Targets[NormalTask.NextIndex],
				Repeat:    NormalTask.Repeat,
				Fleet:     models.Fleet{}.ToDTO(),
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
			account: testAccount,
			want:    nil,
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
			account: testAccount,
			want:    nil,
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
			account: testAccount,
			want:    nil,
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
			account: testAccount,
			want:    nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := &taskService{
				DB:       tt.fields.DB,
				MQ:       tt.fields.MQ,
				Enforcer: tt.fields.Enforcer,
			}
			got := ts.GenerateSingleTask(tt.args.task, tt.account)

			// Handle nil case
			if got == nil && tt.want == nil {
				return
			}
			if (got == nil) != (tt.want == nil) {
				t.Errorf("taskService.GenerateSingleTask() = %v, want %v", got, tt.want)
				return
			}

			// For non-nil results, compare fields except UUID
			if got != nil {
				// Temporarily store the generated UUID
				gotUUID := got.UUID
				// Set UUID to empty for comparison
				got.UUID = ""
				if tt.want != nil {
					tt.want.UUID = ""
				}

				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("taskService.GenerateSingleTask() = %v, want %v", got, tt.want)
				}
				// Restore the UUID if needed
				got.UUID = gotUUID
			}
		})
	}
}
