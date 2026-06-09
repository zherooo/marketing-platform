package service

import (
	"time"

	"marketing-platform/internal/database"
	"marketing-platform/internal/model"

	"gorm.io/gorm"
)

// HostingAlertService 托管告警服务
type HostingAlertService struct {
	db *gorm.DB
}

// NewHostingAlertService 创建托管告警服务
func NewHostingAlertService() *HostingAlertService {
	return &HostingAlertService{
		db: database.GetDB(),
	}
}

// ListAlerts 获取告警列表
func (s *HostingAlertService) ListAlerts(accountID, alertType string, severity, status int, page, pageSize int) ([]model.HostingAlert, int64, error) {
	var alerts []model.HostingAlert
	var total int64

	query := s.db.Model(&model.HostingAlert{})
	if accountID != "" {
		query = query.Where("account_id = ?", accountID)
	}
	if alertType != "" {
		query = query.Where("alert_type = ?", alertType)
	}
	if severity > 0 {
		query = query.Where("severity = ?", severity)
	}
	if status > 0 {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&alerts).Error; err != nil {
		return nil, 0, err
	}

	return alerts, total, nil
}

// GetAlertByID 获取告警详情
func (s *HostingAlertService) GetAlertByID(id uint) (*model.HostingAlert, error) {
	var alert model.HostingAlert
	if err := s.db.First(&alert, id).Error; err != nil {
		return nil, err
	}
	return &alert, nil
}

// MarkAsRead 标记为已读
func (s *HostingAlertService) MarkAsRead(id uint) error {
	now := time.Now()
	return s.db.Model(&model.HostingAlert{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":  model.AlertStatusRead,
		"read_at": now,
	}).Error
}

// HandleAlert 处理告警
func (s *HostingAlertService) HandleAlert(id uint, handler, result string) error {
	now := time.Now()
	return s.db.Model(&model.HostingAlert{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":        model.AlertStatusHandled,
		"handler":       handler,
		"handle_result": result,
		"handled_at":    now,
	}).Error
}

// IgnoreAlert 忽略告警
func (s *HostingAlertService) IgnoreAlert(id uint) error {
	return s.db.Model(&model.HostingAlert{}).Where("id = ?", id).Update("status", model.AlertStatusIgnored).Error
}

// GetUnreadAlertCount 获取未读告警数量
func (s *HostingAlertService) GetUnreadAlertCount() (int64, error) {
	var count int64
	err := s.db.Model(&model.HostingAlert{}).Where("status = ?", model.AlertStatusUnread).Count(&count).Error
	return count, err
}
