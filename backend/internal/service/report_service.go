package service

import (
	"time"

	"marketing-platform/internal/database"
	"marketing-platform/internal/model"
)

// ReportService 报表服务
type ReportService struct{}

// NewReportService 创建报表服务实例
func NewReportService() *ReportService {
	return &ReportService{}
}

// GetDailyReports 获取日报表
func (s *ReportService) GetDailyReports(accountID string, startDate, endDate time.Time, page, pageSize int) ([]model.DailyReport, int64, error) {
	var reports []model.DailyReport
	var total int64

	query := database.GetDB().Model(&model.DailyReport{}).
		Where("account_id = ?", accountID).
		Where("date >= ? AND date <= ?", startDate, endDate)

	// 统计总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Order("date DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&reports).Error; err != nil {
		return nil, 0, err
	}

	return reports, total, nil
}

// GetHourlyReports 获取小时报表
func (s *ReportService) GetHourlyReports(accountID string, date time.Time, page, pageSize int) ([]model.HourlyReport, int64, error) {
	var reports []model.HourlyReport
	var total int64

	query := database.GetDB().Model(&model.HourlyReport{}).
		Where("account_id = ?", accountID).
		Where("date = ?", date)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("hour DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&reports).Error; err != nil {
		return nil, 0, err
	}

	return reports, total, nil
}

// GetTargetReports 获取定向报表
func (s *ReportService) GetTargetReports(accountID string, startDate, endDate time.Time, page, pageSize int) ([]model.TargetReport, int64, error) {
	var reports []model.TargetReport
	var total int64

	query := database.GetDB().Model(&model.TargetReport{}).
		Where("account_id = ?", accountID).
		Where("date >= ? AND date <= ?", startDate, endDate)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("date DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&reports).Error; err != nil {
		return nil, 0, err
	}

	return reports, total, nil
}

// GetDailyTrend 获取日报趋势
func (s *ReportService) GetDailyTrend(accountID string, startDate, endDate time.Time) ([]map[string]interface{}, error) {
	var results []map[string]interface{}

	err := database.GetDB().Model(&model.DailyReport{}).
		Select("date, SUM(view_count) as view_count, SUM(click_count) as click_count, SUM(spend) as spend").
		Where("account_id = ?", accountID).
		Where("date >= ? AND date <= ?", startDate, endDate).
		Group("date").
		Order("date ASC").
		Scan(&results).Error

	return results, err
}

// SaveDailyReport 保存日报
func (s *ReportService) SaveDailyReport(report *model.DailyReport) error {
	return database.GetDB().Save(report).Error
}

// BatchSaveDailyReports 批量保存日报
func (s *ReportService) BatchSaveDailyReports(reports []model.DailyReport) error {
	if len(reports) == 0 {
		return nil
	}
	return database.GetDB().CreateInBatches(reports, 1000).Error
}
