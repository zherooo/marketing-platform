package model

import (
	"time"
)

// 任务状态常量
const (
	TaskStatusPending   = 0 // 待处理
	TaskStatusRunning   = 1 // 运行中
	TaskStatusCompleted = 2 // 已完成
	TaskStatusFailed    = 3 // 失败
	TaskStatusCancelled = 4 // 已取消
)

// 数据类型常量
const (
	DataTypeHourlyReport = "hourly_report" // 小时报表
	DataTypeDailyReport  = "daily_report"  // 日报表
	DataTypeCampaign     = "campaign"       // 广告系列
	DataTypeAdGroup      = "adgroup"       // 广告组
	DataTypeAd           = "ad"            // 广告
	DataTypeCreative     = "creative"      // 广告创意
	DataTypeMaterial     = "material"       // 广告素材
)

// CrawlTask 抓取任务表模型
type CrawlTask struct {
	ID           uint       `gorm:"primarykey" json:"id"`
	TaskID       string     `gorm:"uniqueIndex;size:64;not null;comment:任务唯一标识" json:"task_id"`
	AccountID    string     `gorm:"index;size:64;not null;comment:账号ID" json:"account_id"`
	DataType     string     `gorm:"size:50;not null;comment:数据类型" json:"data_type"`
	StartDate    time.Time  `gorm:"not null;comment:开始日期" json:"start_date"`
	EndDate      time.Time  `gorm:"not null;comment:结束日期" json:"end_date"`
	Status       int        `gorm:"default:0;comment:状态 0-待处理 1-运行中 2-已完成 3-失败 4-已取消" json:"status"`
	Progress     int        `gorm:"default:0;comment:进度百分比 0-100" json:"progress"`
	TotalCount   int        `gorm:"default:0;comment:总记录数" json:"total_count"`
	SuccessCount int        `gorm:"default:0;comment:成功数" json:"success_count"`
	FailCount    int        `gorm:"default:0;comment:失败数" json:"fail_count"`
	ErrorMsg     string     `gorm:"type:text;comment:错误信息" json:"error_msg"`
	RetryCount   int        `gorm:"default:0;comment:重试次数" json:"retry_count"`
	NextRetryAt  *time.Time `gorm:"comment:下次重试时间" json:"next_retry_at"`
	StartedAt    *time.Time `gorm:"comment:开始时间" json:"started_at"`
	CompletedAt  *time.Time `gorm:"comment:完成时间" json:"completed_at"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// TableName 指定表名
func (CrawlTask) TableName() string {
	return "crawl_task"
}

// CrawlRequest 抓取请求
type CrawlRequest struct {
	AccountIDs []string   `json:"account_ids"` // 账号ID列表
	DataTypes  []string   `json:"data_types"`  // 数据类型列表
	StartDate  time.Time  `json:"start_date"`  // 开始日期
	EndDate    time.Time  `json:"end_date"`    // 结束日期
	Manual     bool       `json:"manual"`      // 是否手动触发
}

// CrawlTaskItem 任务项（用于协程间传递）
type CrawlTaskItem struct {
	TaskID    string    `json:"task_id"`
	AccountID string    `json:"account_id"`
	DataType  string    `json:"data_type"`
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
}

// CrawlTaskLog 任务日志
type CrawlTaskLog struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	TaskID    string    `gorm:"index;size:64;not null;comment:任务ID" json:"task_id"`
	Level     string    `gorm:"size:10;not null;comment:日志级别 info/warn/error" json:"level"`
	Message   string    `gorm:"type:text;not null;comment:日志内容" json:"message"`
	CreatedAt time.Time `json:"created_at"`
}

// TableName 指定表名
func (CrawlTaskLog) TableName() string {
	return "crawl_task_log"
}

// CrawlStatistics 抓取统计
type CrawlStatistics struct {
	TotalTasks     int64 `json:"total_tasks"`     // 总任务数
	PendingTasks   int64 `json:"pending_tasks"`   // 待处理任务数
	RunningTasks   int64 `json:"running_tasks"`    // 运行中任务数
	CompletedTasks int64 `json:"completed_tasks"` // 已完成任务数
	FailedTasks    int64 `json:"failed_tasks"`    // 失败任务数
}
