package service

import (
	"marketing-platform/internal/database"
	"marketing-platform/internal/model"

	"gorm.io/gorm"
)

// MaterialService 广告素材服务
type MaterialService struct {
	db *gorm.DB
}

// NewMaterialService 创建广告素材服务
func NewMaterialService() *MaterialService {
	return &MaterialService{
		db: database.GetDB(),
	}
}

// GetMaterials 获取广告素材列表
func (s *MaterialService) GetMaterials(accountID string, materialType string, page, pageSize int) ([]model.AdMaterial, int64, error) {
	var materials []model.AdMaterial
	var total int64

	query := s.db.Model(&model.AdMaterial{})
	if accountID != "" {
		query = query.Where("account_id = ?", accountID)
	}
	if materialType != "" {
		query = query.Where("material_type = ?", materialType)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&materials).Error; err != nil {
		return nil, 0, err
	}

	return materials, total, nil
}

// GetMaterialByID 获取广告素材详情
func (s *MaterialService) GetMaterialByID(materialID string) (*model.AdMaterial, error) {
	var material model.AdMaterial
	if err := s.db.Where("material_id = ?", materialID).First(&material).Error; err != nil {
		return nil, err
	}
	return &material, nil
}

// GetMaterialStats 获取素材统计
func (s *MaterialService) GetMaterialStats(accountID string) (map[string]interface{}, error) {
	var stats struct {
		Total     int64 `gorm:"column:total"`
		Image     int64 `gorm:"column:image"`
		Video     int64 `gorm:"column:video"`
		Text      int64 `gorm:"column:text"`
		Card      int64 `gorm:"column:card"`
		MiniApp   int64 `gorm:"column:mini_app"`
	}

	// 统计总数
	s.db.Model(&model.AdMaterial{}).Where("account_id = ?", accountID).Count(&stats.Total)

	// 统计各类型数量
	s.db.Model(&model.AdMaterial{}).Where("account_id = ? AND material_type = ?", accountID, 1).Count(&stats.Image)
	s.db.Model(&model.AdMaterial{}).Where("account_id = ? AND material_type = ?", accountID, 2).Count(&stats.Video)
	s.db.Model(&model.AdMaterial{}).Where("account_id = ? AND material_type = ?", accountID, 3).Count(&stats.Text)
	s.db.Model(&model.AdMaterial{}).Where("account_id = ? AND material_type = ?", accountID, 4).Count(&stats.Card)
	s.db.Model(&model.AdMaterial{}).Where("account_id = ? AND material_type = ?", accountID, 5).Count(&stats.MiniApp)

	return map[string]interface{}{
		"total":    stats.Total,
		"image":    stats.Image,
		"video":    stats.Video,
		"text":     stats.Text,
		"card":     stats.Card,
		"mini_app": stats.MiniApp,
	}, nil
}
