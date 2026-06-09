package service

import (
	"time"

	"gorm.io/gorm"

	"marketing-platform/internal/model"
)

// CrawlService 抓取服务 - 查询模块与拉取模块分离
// 职责：只负责抓取任务相关的数据库操作
type CrawlService struct {
	db *gorm.DB
}

// NewCrawlService 创建抓取服务
func NewCrawlService() *CrawlService {
	return &CrawlService{}
}

// SetDB 设置数据库实例
func (s *CrawlService) SetDB(db *gorm.DB) {
	s.db = db
}

// SaveTask 保存抓取任务
func (s *CrawlService) SaveTask(task *model.CrawlTask) error {
	return s.db.Create(task).Error
}

// GetTaskByTaskID 根据任务ID获取任务
func (s *CrawlService) GetTaskByTaskID(taskID string) (*model.CrawlTask, error) {
	var task model.CrawlTask
	if err := s.db.Where("task_id = ?", taskID).First(&task).Error; err != nil {
		return nil, err
	}
	return &task, nil
}

// UpdateTask 更新任务
func (s *CrawlService) UpdateTask(taskID string, updates map[string]interface{}) error {
	return s.db.Model(&model.CrawlTask{}).Where("task_id = ?", taskID).Updates(updates).Error
}

// UpdateTaskStatus 更新任务状态
func (s *CrawlService) UpdateTaskStatus(taskID string, status int) error {
	updates := map[string]interface{}{
		"status": status,
	}

	now := time.Now()
	switch status {
	case model.TaskStatusRunning:
		updates["started_at"] = now
	case model.TaskStatusCompleted:
		updates["completed_at"] = now
	}

	return s.db.Model(&model.CrawlTask{}).Where("task_id = ?", taskID).Updates(updates).Error
}

// UpdateTaskProgress 更新任务进度
func (s *CrawlService) UpdateTaskProgress(taskID string, progress, totalCount, successCount, failCount int) error {
	return s.db.Model(&model.CrawlTask{}).Where("task_id = ?", taskID).Updates(map[string]interface{}{
		"progress":      progress,
		"total_count":   totalCount,
		"success_count": successCount,
		"fail_count":    failCount,
	}).Error
}

// HandleTaskError 处理任务错误
func (s *CrawlService) HandleTaskError(taskID string, errorMsg string) error {
	var task model.CrawlTask
	if err := s.db.Where("task_id = ?", taskID).First(&task).Error; err != nil {
		return err
	}

	updates := map[string]interface{}{
		"error_msg":    errorMsg,
		"retry_count":  task.RetryCount + 1,
	}

	if task.RetryCount < 3 {
		updates["status"] = model.TaskStatusPending
		nextRetry := time.Now().Add(5 * time.Second)
		updates["next_retry_at"] = nextRetry
	} else {
		updates["status"] = model.TaskStatusFailed
	}

	return s.db.Model(&model.CrawlTask{}).Where("task_id = ?", taskID).Updates(updates).Error
}

// ListTasks 查询任务列表
func (s *CrawlService) ListTasks(accountID string, status int, page, pageSize int) ([]model.CrawlTask, int64, error) {
	var tasks []model.CrawlTask
	var total int64

	query := s.db.Model(&model.CrawlTask{})

	if accountID != "" {
		query = query.Where("account_id = ?", accountID)
	}
	if status >= 0 {
		query = query.Where("status = ?", status)
	}

	// 统计总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&tasks).Error; err != nil {
		return nil, 0, err
	}

	return tasks, total, nil
}

// GetStatistics 获取抓取统计
func (s *CrawlService) GetStatistics() (*model.CrawlStatistics, error) {
	stats := &model.CrawlStatistics{}

	var total int64
	s.db.Model(&model.CrawlTask{}).Count(&total)
	stats.TotalTasks = total

	s.db.Model(&model.CrawlTask{}).Where("status = ?", model.TaskStatusPending).Count(&stats.PendingTasks)
	s.db.Model(&model.CrawlTask{}).Where("status = ?", model.TaskStatusRunning).Count(&stats.RunningTasks)
	s.db.Model(&model.CrawlTask{}).Where("status = ?", model.TaskStatusCompleted).Count(&stats.CompletedTasks)
	s.db.Model(&model.CrawlTask{}).Where("status = ?", model.TaskStatusFailed).Count(&stats.FailedTasks)

	return stats, nil
}

// GetRunningTasks 获取正在运行的任务
func (s *CrawlService) GetRunningTasks() ([]model.CrawlTask, error) {
	var tasks []model.CrawlTask
	if err := s.db.Where("status = ?", model.TaskStatusRunning).Order("started_at ASC").Find(&tasks).Error; err != nil {
		return nil, err
	}
	return tasks, nil
}

// GetPendingTasks 获取待处理的任务
func (s *CrawlService) GetPendingTasks(limit int) ([]model.CrawlTask, error) {
	var tasks []model.CrawlTask
	query := s.db.Where("status = ?", model.TaskStatusPending).
		Where("retry_count < ?", 3).
		Order("created_at ASC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Find(&tasks).Error; err != nil {
		return nil, err
	}
	return tasks, nil
}

// SaveTaskLog 保存任务日志
func (s *CrawlService) SaveTaskLog(taskID, level, message string) error {
	log := &model.CrawlTaskLog{
		TaskID:  taskID,
		Level:   level,
		Message: message,
	}
	return s.db.Create(log).Error
}

// GetTaskLogs 获取任务日志
func (s *CrawlService) GetTaskLogs(taskID string, limit int) ([]model.CrawlTaskLog, error) {
	var logs []model.CrawlTaskLog
	query := s.db.Where("task_id = ?", taskID).Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Find(&logs).Error; err != nil {
		return nil, err
	}
	return logs, nil
}

// CancelTask 取消任务
func (s *CrawlService) CancelTask(taskID string) error {
	return s.db.Model(&model.CrawlTask{}).Where("task_id = ?", taskID).Update("status", model.TaskStatusCancelled).Error
}

// RetryTask 重试任务
func (s *CrawlService) RetryTask(taskID string) error {
	return s.db.Model(&model.CrawlTask{}).Where("task_id = ?", taskID).Updates(map[string]interface{}{
		"status":       model.TaskStatusPending,
		"retry_count":  0,
		"error_msg":    "",
		"next_retry_at": nil,
	}).Error
}

// CleanOldTasks 清理旧任务
func (s *CrawlService) CleanOldTasks(days int) (int64, error) {
	before := time.Now().AddDate(0, 0, -days)
	result := s.db.Where("created_at < ? AND status IN (?, ?, ?)",
		before, model.TaskStatusCompleted, model.TaskStatusFailed, model.TaskStatusCancelled).
		Delete(&model.CrawlTask{})
	return result.RowsAffected, result.Error
}
