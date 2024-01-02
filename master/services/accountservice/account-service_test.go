package accountservice

import (
	"GalaxyEmpireWeb/models"
	"GalaxyEmpireWeb/repositories/redis"
	"GalaxyEmpireWeb/repositories/sqlite"
	"GalaxyEmpireWeb/services/userservice"
	"GalaxyEmpireWeb/utils"
	"context"
	"testing"

	"gorm.io/gorm"
)

func TestAccountService_GetAccountById(t *testing.T) {
	db := sqlite.GetTestDB()
	db.AutoMigrate(&models.User{}, &models.Account{})

	ctx := utils.NewContextWithTraceID()
	ctx = context.WithValue(ctx, "userID", uint(1)) // TODO: add to case
	ctx = context.WithValue(ctx, "role", 1)
	tests := []struct {
		name    string
		setup   func(*gorm.DB) uint
		wantID  uint
		wantErr bool
	}{
		{
			name: "Valid account ID",
			setup: func(tx *gorm.DB) uint {
				// Create test data within the transaction
				account := models.Account{Username: "testuser"}
				tx.Create(&account)
				return account.ID
			},
			wantID:  1,
			wantErr: false,
		},
		{
			name:    "Invalid account ID",
			setup:   func(tx *gorm.DB) uint { return 999 },
			wantID:  0,
			wantErr: true,
		},
		// 其他测试用例...
	}

	rdb := redis.GetRedisDB()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Begin a transaction
			tx := db.Begin()
			defer tx.Rollback()

			InitService(tx, rdb)
			userservice.InitService(tx, rdb)
			service, err := GetService(ctx)
			id := tt.setup(tx)

			got, err := service.GetById(ctx, id, []string{})
			if (err != nil) != tt.wantErr {
				t.Errorf("AccountService.GetAccountById() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got.ID != tt.wantID {
				t.Errorf("AccountService.GetAccountById() = %v, want %v", got, tt.wantID)
			}
		})
	}
}

func TestAccountService_GetByUserId(t *testing.T) {
	db := sqlite.GetTestDB()
	models.AutoMigrate(db)
	ctx := utils.NewContextWithTraceID()

	tests := []struct {
		name    string
		setup   func(*gorm.DB) uint
		wantErr bool
		want    []models.Account
	}{
		{
			name: "Test get accounts by user ID",
			setup: func(tx *gorm.DB) uint {
				// Create test data within the transaction
				account1 := models.Account{Username: "testaccount1",
					Password: "testpassword",
					Email:    "test1@example.com",
					Server:   "testserver",
				}
				account2 := models.Account{
					Username: "testaccount2",
					Password: "testpassword",
					Email:    "test2@example.com",
					Server:   "testserver",
				}
				user := models.User{Username: "testuser",
					Password: "testpassword",
					Accounts: []models.Account{account1, account2},
				}
				tx.Create(&user)
				return user.ID
			},
			wantErr: false,
			want: []models.Account{
				{
					Username: "testaccount1",
					Password: "testpassword",
					Email:    "test1@example.com",
					Server:   "testserver",
				},
				{
					Username: "testaccount2",
					Password: "testpassword",
					Email:    "test2@example.com",
					Server:   "testserver"},
			},
		},
		{
			name:    "Test get accounts by invalid user ID",
			setup:   func(tx *gorm.DB) uint { return 999 },
			wantErr: true,
			want:    nil,
		},
	}
	rdb := redis.GetRedisDB()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Begin a transaction
			tx := db.Begin()
			defer tx.Rollback()

			userId := tt.setup(tx)

			service := NewService(tx, rdb)
			got, err := service.GetByUserId(ctx, userId, []string{})
			if (err != nil) != tt.wantErr {
				t.Errorf("GetByUserId() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			for i, account := range *got {
				if account.Username != tt.want[i].Username {
					t.Errorf("GetByUserId() = %v, want %v", account, tt.want[i])
				}
			}
		})
	}
}
