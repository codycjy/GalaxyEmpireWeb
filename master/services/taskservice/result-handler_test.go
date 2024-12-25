package taskservice

import (
	"GalaxyEmpireWeb/config"
	"GalaxyEmpireWeb/models"
	"GalaxyEmpireWeb/queue"
	"GalaxyEmpireWeb/services/casbinservice"
	"reflect"
	"testing"
	"time"

	"gorm.io/gorm"
)

func Test_taskService_HandleSingleResult(t *testing.T) {
	type fields struct {
		DB       *gorm.DB
		MQ       *queue.RabbitMQConnection
		Enforcer casbinservice.Enforcer
	}
	type args struct {
		response *models.SingleTaskResponse
	}
	NormalResponse := &models.SingleTaskResponse{
		TaskID:        1,
		Status:        0,
		TaskType:      1,
		BackTimestamp: 1,
	}
	LoginResponse := &models.SingleTaskResponse{
		TaskID:   1,
		Status:   0,
		TaskType: 99,

		BackTimestamp: 0,
	}
	FailResponse := &models.SingleTaskResponse{
		TaskID:        1,
		TaskType:      15,
		BackTimestamp: 0,
		Status:        -1,
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *models.Task
		wantErr bool
	}{
		{
			name: "NormalResponse",
			fields: fields{
				DB:       nil,
				MQ:       nil,
				Enforcer: nil,
			},
			args:    args{response: NormalResponse},
			want:    &models.Task{Model: gorm.Model{ID: 1}, Status: models.TaskStatusMap[models.TASK_STATUS_READY], NextStart: time.Unix(1, 0).Add(time.Duration(config.TASK_DELAY) * time.Second).Unix()},
			wantErr: false,
		},
		{
			name: "LoginResponse",
			fields: fields{
				DB:       nil,
				MQ:       nil,
				Enforcer: nil,
			},
			args:    args{response: LoginResponse},
			want:    &models.Task{Model: gorm.Model{ID: 1}},
			wantErr: false,
		},
		{
			name: "FailResponse",
			fields: fields{
				DB:       nil,
				MQ:       nil,
				Enforcer: nil,
			},
			args:    args{response: FailResponse},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := &taskService{
				DB:       tt.fields.DB,
				MQ:       tt.fields.MQ,
				Enforcer: tt.fields.Enforcer,
			}
			got, err := ts.HandleSingleResult(tt.args.response)
			if (err != nil) != tt.wantErr {
				t.Errorf("taskService.HandleSingleResult() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("taskService.HandleSingleResult() = %v, want %v", got, tt.want)
			}
		})
	}
}
