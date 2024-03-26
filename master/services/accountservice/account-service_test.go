package accountservice

import (
	"GalaxyEmpireWeb/models"
	"GalaxyEmpireWeb/repositories/redis"
	"GalaxyEmpireWeb/repositories/sqlite"
	"GalaxyEmpireWeb/utils"
	"context"
	"testing"

	r "github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

func Test_accountService_GetById(t *testing.T) {
	ctx := utils.NewContextWithTraceID()
	ctx = context.WithValue(ctx, "userID", uint(1))
	type fields struct {
		DB  *gorm.DB
		RDB *r.Client
	}
	type args struct {
		ctx    context.Context
		id     uint
		fields []string
	}
	// 初始化数据库并进行迁移
	db := sqlite.GetTestDB()
	db.AutoMigrate(&models.Account{})

	tests := []struct {
		name    string
		fields  fields
		args    args
		setup   func(*gorm.DB) *models.Account // 设置函数来插入测试数据
		wantErr bool
	}{
		{name: "Test Get Account By Id",
			fields: fields{
				DB: db,
			},
			setup: func(db *gorm.DB) *models.Account {
				// 插入测试账户
				testAccount := &models.Account{
					Username: "Test Account",
					Password: "testpassword",
					Email:    "test@example.com",
					Server:   "testserver",
					UserID:   1,
				}
				db.Create(testAccount)
				return testAccount
			},
			wantErr: false,
		},
		{
			name: "Test Get Account By Nonexistent Id",
			fields: fields{
				DB: db,
			},
			setup: func(db *gorm.DB) *models.Account {
				return nil
			},
			wantErr: true,
		},
	}
	rdb := redis.GetRedisDB()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 启动事务
			tx := tt.fields.DB.Begin()
			defer tx.Rollback()
			// 插入测试数据
			service := NewService(tx, rdb)
			testAccount := tt.setup(tx)
			var id uint
			if testAccount != nil {
				id = testAccount.ID
			}
			got, err := service.GetById(ctx, id, tt.args.fields)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetById() got = %v, wantErr %v", got, tt.wantErr)
			}
			if got == nil {

			} else {
				if !tt.wantErr && got.Username != testAccount.Username || got.UserID != testAccount.UserID {
					t.Errorf("AccountService.GetById() = %v, want %v", got, testAccount)
				}
			}
		})
	}
}

func TestAccountService_GetByUserId(t *testing.T) {
	db := sqlite.GetTestDB()
	db.AutoMigrate(&models.User{}, &models.Account{})
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Begin a transaction
			tx := db.Begin()
			defer tx.Rollback()

			userId := tt.setup(tx)
			service := &accountService{
				DB: tx,
			}
			got, err := service.GetByUserId(ctx, userId, []string{})

			if (err != nil) != tt.wantErr {
				if err != nil {
					t.Errorf("GetByUserId() error = %v, wantErr %v", err.Error(), tt.wantErr)
				} else {
					t.Errorf("GetByUserId() error is nil,wantErr: %v got: %v", tt.wantErr, got)

				}
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

func Test_accountService_Create(t *testing.T) {
	ctx := utils.NewContextWithTraceID()
	ctx = context.WithValue(ctx, "userID", uint(1))
	ctx = context.WithValue(ctx, "role", uint(1))
	type fields struct {
		DB *gorm.DB
	}
	type args struct {
		ctx     context.Context
		account *models.Account
	}

	// 初始化数据库并进行迁移
	db := sqlite.GetTestDB()
	db.AutoMigrate(&models.Account{})

	setup := func(db *gorm.DB) {
		// 插入测试账户
		testAccount := &models.Account{
			Username: "TestUser",
			Password: "testpassword",
			Email:    "test@test.com",
			Server:   "testserver",
			UserID:   1,
		}
		db.Create(testAccount)
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Test Create Account",
			fields: fields{
				DB: db,
			},
			args: args{
				ctx: ctx,
				account: &models.Account{
					Username: "Create Account",
					Password: "testpassword",
					Email:    "new@test.com",
					Server:   "test1",
					UserID:   1,
				},
			},
			wantErr: false,
		},
		{
			name: "Test Create Account with Duplicate Username",
			fields: fields{
				DB: db,
			},
			args: args{
				ctx: ctx,
				account: &models.Account{
					Username: "TestUser",
					Password: "testpassword",
					Email:    "duplicate@test.com",
					Server:   "testserver",
					UserID:   1,
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := tt.fields.DB.Begin()
			defer tx.Rollback()
			setup(tx)
			service := &accountService{
				DB: tx,
			}
			if err := service.Create(ctx, tt.args.account); (err != nil) != tt.wantErr {
				t.Errorf("accountService.Create() = %v, want %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				// 尝试数据库中寻找创建的账户
				var createdAccount models.Account
				err := tx.Where("username = ?", tt.args.account.Username).First(&createdAccount).Error
				if err != nil {
					t.Errorf("AccountService.Create() account was not created in the database, want account to be created")
				}

			}
		})
	}
}

func Test_accountService_Delete(t *testing.T) {
	ctx := utils.NewContextWithTraceID()
	ctx = context.WithValue(ctx, "userID", uint(1))
	db := sqlite.GetTestDB()
	db.AutoMigrate(&models.Account{})

	type fields struct {
		DB  *gorm.DB
		RDB *r.Client
	}
	type args struct {
		ID uint
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		setup   func(*gorm.DB) uint // 设置函数来插入测试数据
		wantErr bool
	}{
		{
			name: "Test Delete Existing Account",
			fields: fields{
				DB: db,
			},
			setup: func(db *gorm.DB) uint {
				// 插入测试账户
				testAccount := &models.Account{
					Username: "Test Account",
					Password: "testpassword",
					Email:    "test@test.com",
					Server:   "testserver",
					UserID:   1,
				}
				db.Create(testAccount)
				return testAccount.ID
			},
			wantErr: false,
		},
		{
			name: "Test Delete Nonexistent Account",
			fields: fields{
				DB: db,
			},
			args: args{
				ID: 999, // 不存在的账户ID
			},
			wantErr: true,
		},
	}
	rdb := redis.GetRedisDB()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := tt.fields.DB.Begin()
			defer tx.Rollback()

			if tt.setup != nil {
				tt.args.ID = tt.setup(tx)
			}

			service := NewService(tx, rdb)
			if err := service.Delete(ctx, tt.args.ID); (err != nil) != tt.wantErr {
				t.Errorf("accountService.Delete() = %v, want %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				var deletedAccount models.Account
				err := tx.Where("id = ?", tt.args.ID).First(&deletedAccount).Error
				if err == nil {
					t.Errorf("AccountService.Delete()= %v,account still exists in the database", deletedAccount)
				}
			}
		})
	}
}

func Test_accountService_Update(t *testing.T) {
	ctx := utils.NewContextWithTraceID()
	ctx = context.WithValue(ctx, "userID", uint(1))
	db := sqlite.GetTestDB()
	db.AutoMigrate(&models.Account{})
	type fields struct {
		DB  *gorm.DB
		RDB *r.Client
	}
	type args struct {
		account *models.Account
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		setup   func(*gorm.DB) uint // 设置函数来插入测试数据
		wantErr bool
	}{
		{
			name: "Test Update Account",
			fields: fields{
				DB: db,
			},
			args: args{
				account: &models.Account{
					Username: "Updated Account",
					Password: "updatepassword",
					Email:    "update@test.com",
					Server:   "updateserver",
					UserID:   1,
				},
			},
			setup: func(db *gorm.DB) uint {
				// Insert a test account into the database
				testAccount := &models.Account{
					Username: "Updated Account",
					Password: "testpassword",
					Email:    "test@test.com",
					Server:   "updateserver",
					UserID:   1,
				}
				db.Create(testAccount)
				return testAccount.ID
			},
			wantErr: false,
		},
	}
	rdb := redis.GetRedisDB()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 启动事务
			tx := tt.fields.DB.Begin()
			// 回滚事务
			defer tx.Rollback()
			Updateid := tt.setup(tx)
			service := NewService(tx, rdb)
			tt.args.account.ID = Updateid
			if err := service.Update(ctx, tt.args.account); (err != nil) != tt.wantErr {
				t.Errorf("accountService.Update() error= %v, want %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				// 尝试数据库中寻找更新的账户
				var updatedAccount models.Account
				tx.First(&updatedAccount, Updateid)
				if updatedAccount.Username != tt.args.account.Username || updatedAccount.Email != tt.args.account.Email ||
					updatedAccount.Server != tt.args.account.Server || updatedAccount.Password != tt.args.account.Password {
					t.Errorf("AccountService.Update() =%v, want %v", updatedAccount, tt.args.account)
				}
			}
		})
	}
}
