package service

import (
	"marketing-platform/internal/database"
	"marketing-platform/internal/model"

	"gorm.io/gorm"
)

// CampaignService 广告系列服务
type CampaignService struct {
	db *gorm.DB
}

// NewCampaignService 创建广告系列服务
func NewCampaignService() *CampaignService {
	return &CampaignService{
		db: database.GetDB(),
	}
}

// GetCampaigns 获取广告系列列表
func (s *CampaignService) GetCampaigns(accountID string, page, pageSize int) ([]model.Campaign, int64, error) {
	var campaigns []model.Campaign
	var total int64

	query := s.db.Model(&model.Campaign{})
	if accountID != "" {
		query = query.Where("account_id = ?", accountID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&campaigns).Error; err != nil {
		return nil, 0, err
	}

	return campaigns, total, nil
}

// GetCampaignByID 获取广告系列详情
func (s *CampaignService) GetCampaignByID(campaignID string) (*model.Campaign, error) {
	var campaign model.Campaign
	if err := s.db.Where("campaign_id = ?", campaignID).First(&campaign).Error; err != nil {
		return nil, err
	}
	return &campaign, nil
}

// GetCampaignsByProject 获取项目下的广告系列
func (s *CampaignService) GetCampaignsByProject(accountID, projectID string, page, pageSize int) ([]model.Campaign, int64, error) {
	var campaigns []model.Campaign
	var total int64

	query := s.db.Model(&model.Campaign{}).Where("account_id = ?", accountID)
	if projectID != "" {
		query = query.Where("project_id = ?", projectID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&campaigns).Error; err != nil {
		return nil, 0, err
	}

	return campaigns, total, nil
}

// AdGroupService 广告组服务
type AdGroupService struct {
	db *gorm.DB
}

// NewAdGroupService 创建广告组服务
func NewAdGroupService() *AdGroupService {
	return &AdGroupService{
		db: database.GetDB(),
	}
}

// GetAdGroups 获取广告组列表
func (s *AdGroupService) GetAdGroups(accountID, campaignID string, page, pageSize int) ([]model.AdGroup, int64, error) {
	var adGroups []model.AdGroup
	var total int64

	query := s.db.Model(&model.AdGroup{})
	if accountID != "" {
		query = query.Where("account_id = ?", accountID)
	}
	if campaignID != "" {
		query = query.Where("campaign_id = ?", campaignID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&adGroups).Error; err != nil {
		return nil, 0, err
	}

	return adGroups, total, nil
}

// GetAdGroupByID 获取广告组详情
func (s *AdGroupService) GetAdGroupByID(adGroupID string) (*model.AdGroup, error) {
	var adGroup model.AdGroup
	if err := s.db.Where("adgroup_id = ?", adGroupID).First(&adGroup).Error; err != nil {
		return nil, err
	}
	return &adGroup, nil
}

// GetCampaignAdGroups 获取广告系列下的广告组
func (s *AdGroupService) GetCampaignAdGroups(campaignID string, page, pageSize int) ([]model.AdGroup, int64, error) {
	var adGroups []model.AdGroup
	var total int64

	query := s.db.Model(&model.AdGroup{}).Where("campaign_id = ?", campaignID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&adGroups).Error; err != nil {
		return nil, 0, err
	}

	return adGroups, total, nil
}
