package userservice

import (
	"GalaxyEmpireWeb/models"
	"GalaxyEmpireWeb/repositories/sqlite"
	"testing"

	"gorm.io/gorm"
)

func TestUserService_GetById(t *testing.T) {
	type fields struct {
		DB *gorm.DB
	}
	type args struct {
		id     uint
		fields []string
	}

	// 初始化数据库并进行迁移
	db := sqlite.GetTestDB()
	db.AutoMigrate(&models.User{})

	tests := []struct {
		name    string
		fields  fields
		args    args
		setup   func(*gorm.DB) *models.User // 设置函数来插入测试数据
		wantErr bool
	}{
		{
			name: "Test Get User By Id",
			fields: fields{
				DB: db,
			},
			setup: func(db *gorm.DB) *models.User {
				// 在事务中插入测试用户
				testUser := &models.User{
					Username: "Test User",
					Password: "testpassword",
					Balance:  100,
				}
				db.Create(testUser)
				return testUser
			},
			wantErr: false,
		},
		{
			name: "Test Get User By Nonexistent Id",
			fields: fields{
				DB: db,
			},
			setup: func(db *gorm.DB) *models.User {
				return nil
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 启动事务
			tx := tt.fields.DB.Begin()
			defer func() {
				// 测试结束时回滚事务
				tx.Rollback()
			}()

			InitService(tx)
			service,err:=GetService()
			// 设置测试数据
			testUser := tt.setup(tx)

			var id uint
			if testUser != nil {
				id = testUser.ID
			}

			got, err := service.GetById(id, tt.args.fields)
			if (err != nil) != tt.wantErr {
				t.Errorf("UserService.GetById() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && (got.Username != testUser.Username || got.Password != testUser.Password || got.Balance != testUser.Balance) {
				t.Errorf("UserService.GetById() = %v, want %v", got, testUser)
			}
		})
	}

}

func TestUserService_Create(t *testing.T) {
	type fields struct {
		DB *gorm.DB
	}
	type args struct {
		user *models.User
	}

	// 初始化数据库并进行迁移
	db := sqlite.GetTestDB()
	db.AutoMigrate(&models.User{})

	setup := func(db *gorm.DB) {
		// 插入重复用户以用于测试
		duplicateUser := &models.User{
			Username: "Duplicate User",
			Password: "testpassword",
			Balance:  100,
		}
		db.Create(duplicateUser)
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Test Create User",
			fields: fields{
				DB: db,
			},
			args: args{
				user: &models.User{
					Username: "Create User",
					Password: "testpassword",
					Balance:  100,
				},
			},
			wantErr: false,
		},
		{
			name: "Test Create User with Duplicate Username",
			fields: fields{
				DB: db,
			},
			args: args{
				user: &models.User{
					Username: "Duplicate User",
					Password: "testpassword",
					Balance:  100,
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 启动事务
			tx := tt.fields.DB.Begin()
			defer func() {
				// 测试结束时回滚事务
				tx.Rollback()
			}()

			// 设置测试环境
			setup(tx)

			service := &UserService{
				DB: tx,
			}
			if err := service.Create(tt.args.user); (err != nil) != tt.wantErr {
				t.Errorf("UserService.Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// 尝试从数据库中获取创建的用户
				var createdUser models.User
				err := tx.Where("username = ?", tt.args.user.Username).First(&createdUser).Error

				// 如果用户被找到，则创建成功
				if err != nil {
					t.Errorf("UserService.Create() user was not created in the database, want user to be created")
				}
			}
		})
	}
}

func TestUserService_Update(t *testing.T) {
	db := sqlite.GetTestDB()
	db.AutoMigrate(&models.User{}) // Create User table

	// Insert a test user into the database
	testUser := &models.User{
		Username: "Update User",
		Password: "testpassword",
		Balance:  100,
	}

	type fields struct {
		DB *gorm.DB
	}
	type args struct {
		user *models.User
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Test Update User",
			fields: fields{
				DB: db,
			},
			args: args{
				user: &models.User{
					Model: gorm.Model{
						ID: testUser.ID,
					},
					Username: "Updated User",
					Password: "updatedpassword",
					Balance:  200,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Start a new transaction
			tx := tt.fields.DB.Begin()
			tx.Create(testUser)
			defer func() {
				// Rollback the transaction after test
				tx.Rollback()
			}()

			service := &UserService{
				DB: tx, // Use the transaction as the DB
			}
			if err := service.Update(tt.args.user); (err != nil) != tt.wantErr {
				t.Errorf("UserService.Update() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				// Fetch the updated user from the database
				var updatedUser models.User
				tx.First(&updatedUser, tt.args.user.ID)
				// Check if the user is updated correctly
				if updatedUser.Username != tt.args.user.Username || updatedUser.Password != tt.args.user.Password || updatedUser.Balance != tt.args.user.Balance {
					t.Errorf("UserService.Update() = %v, want %v", updatedUser, tt.args.user)
				}
			}
		})
	}
}
func TestUserService_Delete(t *testing.T) {
	db := sqlite.GetTestDB()
	db.AutoMigrate(&models.User{}) // Create User table

	type fields struct {
		DB *gorm.DB
	}
	type args struct {
		id uint
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Test Delete User",
			fields: fields{
				DB: db,
			},
			args: args{
				id: 0, // We will set this later
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Start a new transaction
			tx := tt.fields.DB.Begin()
			defer func() {
				// Rollback the transaction after test
				tx.Rollback()
			}()

			// Insert a test user into the database
			testUser := &models.User{
				Username: "Delete User",
				Password: "testpassword",
				Balance:  100,
			}
			tx.Create(testUser)
			tt.args.id = testUser.ID

			service := &UserService{
				DB: tx, // Use the transaction as the DB
			}
			if err := service.Delete(tt.args.id); (err != nil) != tt.wantErr {
				t.Errorf("UserService.Delete() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				// Fetch the deleted user from the database
				var deletedUser models.User
				tx.First(&deletedUser, tt.args.id)
				// Check if the user is deleted correctly
				if deletedUser.ID == tt.args.id {
					t.Errorf("UserService.Delete() = %v, user still exists", deletedUser)
				}
			}
		})
	}
}
func TestUserService_GetAllUsers(t *testing.T) {
	type fields struct {
		DB *gorm.DB
	}
	db := sqlite.GetTestDB()
	db.AutoMigrate(&models.User{}) // Create User table

	tests := []struct {
		name      string
		fields    fields
		wantUsers []models.User
		wantErr   bool
	}{
		{
			name: "Test Get All Users",
			fields: fields{
				DB: db,
			},
			wantUsers: []models.User{
				{
					Username: "Test User 1",
					Password: "testpassword",
					Balance:  100,
				},
				{
					Username: "Test User 2",
					Password: "testpassword",
					Balance:  200,
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Start a new transaction
			tx := tt.fields.DB.Begin()
			defer func() {
				// Rollback the transaction after test
				tx.Rollback()
			}()

			// Insert test users into the database
			for _, user := range tt.wantUsers {
				tx.Create(&user)
			}

			service := &UserService{
				DB: tx, // Use the transaction as the DB
			}
			gotUsers, err := service.GetAllUsers()
			if (err != nil) != tt.wantErr {
				t.Errorf("UserService.GetAllUsers() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(gotUsers) != len(tt.wantUsers) {
				t.Errorf("UserService.GetAllUsers() = %v, want %v", len(gotUsers), len(tt.wantUsers))
			}
		})
	}
}
