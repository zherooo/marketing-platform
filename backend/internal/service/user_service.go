package service

import (
	"errors"

	"marketing-platform/internal/database"
	"marketing-platform/internal/model"

	"gorm.io/gorm"
)

// UserService 用户服务
type UserService struct{}

// NewUserService 创建用户服务实例
func NewUserService() *UserService {
	return &UserService{}
}

// FindByID 根据ID查找用户
func (s *UserService) FindByID(id uint) (*model.User, error) {
	var user model.User
	if err := database.GetDB().First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// FindByUsername 根据用户名查找用户
func (s *UserService) FindByUsername(username string) (*model.User, error) {
	var user model.User
	if err := database.GetDB().Where("username = ?", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// Create 创建用户
func (s *UserService) Create(user *model.User) error {
	return database.GetDB().Create(user).Error
}

// Update 更新用户
func (s *UserService) Update(user *model.User) error {
	return database.GetDB().Save(user).Error
}
