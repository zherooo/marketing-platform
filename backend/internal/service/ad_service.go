package service

import (
	"marketing-platform/internal/database"
	"marketing-platform/internal/model"

	"gorm.io/gorm"
)

// AdService 广告服务
type AdService struct {
	db *gorm.DB
}

// NewAdService 创建广告服务
func NewAdService() *AdService {
	return &AdService{
		db: database.GetDB(),
	}
}

// GetAds 获取广告列表
func (s *AdService) GetAds(accountID, adGroupID string, page, pageSize int) ([]model.Ad, int64, error) {
	var ads []model.Ad
	var total int64

	query := s.db.Model(&model.Ad{})
	if accountID != "" {
		query = query.Where("account_id = ?", accountID)
	}
	if adGroupID != "" {
		query = query.Where("adgroup_id = ?", adGroupID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&ads).Error; err != nil {
		return nil, 0, err
	}

	return ads, total, nil
}

// GetAdByID 获取广告详情
func (s *AdService) GetAdByID(adID string) (*model.Ad, error) {
	var ad model.Ad
	if err := s.db.Where("ad_id = ?", adID).First(&ad).Error; err != nil {
		return nil, err
	}
	return &ad, nil
}

// GetAdGroupAds 获取广告组下的广告
func (s *AdService) GetAdGroupAds(adGroupID string, page, pageSize int) ([]model.Ad, int64, error) {
	var ads []model.Ad
	var total int64

	query := s.db.Model(&model.Ad{}).Where("adgroup_id = ?", adGroupID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&ads).Error; err != nil {
		return nil, 0, err
	}

	return ads, total, nil
}

// CreativeService 广告创意服务
type CreativeService struct {
	db *gorm.DB
}

// NewCreativeService 创建广告创意服务
func NewCreativeService() *CreativeService {
	return &CreativeService{
		db: database.GetDB(),
	}
}

// GetCreatives 获取广告创意列表
func (s *CreativeService) GetCreatives(accountID, adID string, page, pageSize int) ([]model.AdCreative, int64, error) {
	var creatives []model.AdCreative
	var total int64

	query := s.db.Model(&model.AdCreative{})
	if accountID != "" {
		query = query.Where("account_id = ?", accountID)
	}
	if adID != "" {
		query = query.Where("ad_id = ?", adID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&creatives).Error; err != nil {
		return nil, 0, err
	}

	return creatives, total, nil
}

// GetCreativeByID 获取广告创意详情
func (s *CreativeService) GetCreativeByID(creativeID string) (*model.AdCreative, error) {
	var creative model.AdCreative
	if err := s.db.Where("creative_id = ?", creativeID).First(&creative).Error; err != nil {
		return nil, err
	}
	return &creative, nil
}
