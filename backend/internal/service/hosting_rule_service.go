package service

import (
	"marketing-platform/internal/database"
	"marketing-platform/internal/model"

	"gorm.io/gorm"
)

// HostingRuleService 托管规则服务
type HostingRuleService struct {
	db *gorm.DB
}

// NewHostingRuleService 创建托管规则服务
func NewHostingRuleService() *HostingRuleService {
	return &HostingRuleService{
		db: database.GetDB(),
	}
}

// CreateRule 创建规则
func (s *HostingRuleService) CreateRule(rule *model.HostingRule) error {
	return s.db.Create(rule).Error
}

// UpdateRule 更新规则
func (s *HostingRuleService) UpdateRule(rule *model.HostingRule) error {
	return s.db.Save(rule).Error
}

// DeleteRule 删除规则
func (s *HostingRuleService) DeleteRule(id uint) error {
	return s.db.Delete(&model.HostingRule{}, id).Error
}

// GetRuleByID 获取规则详情
func (s *HostingRuleService) GetRuleByID(id uint) (*model.HostingRule, error) {
	var rule model.HostingRule
	if err := s.db.First(&rule, id).Error; err != nil {
		return nil, err
	}
	return &rule, nil
}

// ListRules 获取规则列表
func (s *HostingRuleService) ListRules(category, status string, page, pageSize int) ([]model.HostingRule, int64, error) {
	var rules []model.HostingRule
	var total int64

	query := s.db.Model(&model.HostingRule{})
	if category != "" {
		query = query.Where("category = ?", category)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("priority DESC, created_at DESC").Offset(offset).Limit(pageSize).Find(&rules).Error; err != nil {
		return nil, 0, err
	}

	return rules, total, nil
}

// ToggleRuleStatus 切换规则启用/禁用
func (s *HostingRuleService) ToggleRuleStatus(id uint) (*model.HostingRule, error) {
	var rule model.HostingRule
	if err := s.db.First(&rule, id).Error; err != nil {
		return nil, err
	}
	newStatus := int8(1)
	if rule.Status == 1 {
		newStatus = 0
	}
	if err := s.db.Model(&rule).Update("status", newStatus).Error; err != nil {
		return nil, err
	}
	rule.Status = newStatus
	return &rule, nil
}

// ResetDailyCount 重置每日计数（每日零点调用）
func (s *HostingRuleService) ResetDailyCount() error {
	return s.db.Model(&model.HostingRule{}).Where("today_exec_count > 0").
		Update("today_exec_count", 0).Error
}

// GetActiveRules 获取所有启用状态的规则
func (s *HostingRuleService) GetActiveRules() ([]model.HostingRule, error) {
	var rules []model.HostingRule
	if err := s.db.Where("status = ?", 1).Order("priority DESC").Find(&rules).Error; err != nil {
		return nil, err
	}
	return rules, nil
}
