package service

import (
	"errors"
	"time"

	"marketing-platform/internal/database"
	"marketing-platform/internal/model"

	"gorm.io/gorm"
)

// OAuthService OAuth服务
type OAuthService struct{}

// NewOAuthService 创建OAuth服务实例
func NewOAuthService() *OAuthService {
	return &OAuthService{}
}

// FindTokenByAccountID 根据账号ID查找Token
func (s *OAuthService) FindTokenByAccountID(accountID string) (*model.OAuthToken, error) {
	var token model.OAuthToken
	if err := database.GetDB().Where("account_id = ?", accountID).First(&token).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &token, nil
}

// SaveOrUpdateToken 保存或更新Token
func (s *OAuthService) SaveOrUpdateToken(accountID, accessToken, refreshToken string, expiresIn int) error {
	token, err := s.FindTokenByAccountID(accountID)
	if err != nil {
		return err
	}

	expiresAt := time.Now().Add(time.Duration(expiresIn) * time.Second)

	if token == nil {
		// 新建
		token = &model.OAuthToken{
			AccountID:    accountID,
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			ExpiresAt:    expiresAt,
		}
		return database.GetDB().Create(token).Error
	}

	// 更新
	token.AccessToken = accessToken
	token.RefreshToken = refreshToken
	token.ExpiresAt = expiresAt
	return database.GetDB().Save(token).Error
}

// DeleteToken 删除Token
func (s *OAuthService) DeleteToken(accountID string) error {
	return database.GetDB().Where("account_id = ?", accountID).Delete(&model.OAuthToken{}).Error
}

// GetAllTokens 获取所有Token
func (s *OAuthService) GetAllTokens() ([]model.OAuthToken, error) {
	var tokens []model.OAuthToken
	err := database.GetDB().Find(&tokens).Error
	return tokens, err
}
