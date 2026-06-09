package model

import (
	"time"
)

// ReportType 报表类型
type ReportType int

const (
	ReportTypeDaily  ReportType = 1 // 日报表
	ReportTypeHourly ReportType = 2 // 小时报表
	ReportTypeTarget ReportType = 3 // 定向标签报表
)

// ReportLevel 报表级别 (对应腾讯广告 API level 参数)
type ReportLevel string

const (
	ReportLevelAdvertiser ReportLevel = "REPORT_LEVEL_ADVERTISER" // 广告主级别
	ReportLevelCampaign   ReportLevel = "REPORT_LEVEL_CAMPAIGN"   // 推广计划级别
	ReportLevelProject   ReportLevel = "REPORT_LEVEL_PROJECT"    // 项目级别
	ReportLevelAdgroup   ReportLevel = "REPORT_LEVEL_ADGROUP"    // 广告组级别
	ReportLevelAd        ReportLevel = "REPORT_LEVEL_AD"         // 广告级别
	ReportLevelKeyword   ReportLevel = "REPORT_LEVEL_BIDWORD"    // 关键词级别
)

// ReportStatus 报表任务状态
type ReportStatus int

const (
	ReportStatusPending   ReportStatus = 0 // 待处理
	ReportStatusRunning   ReportStatus = 1 // 处理中
	ReportStatusCompleted ReportStatus = 2 // 已完成
	ReportStatusFailed    ReportStatus = 3 // 失败
)

// DailyReport 日报表 - 对应腾讯广告 daily_reports/get 接口
// 参考: https://developers.e.qq.com/docs/api/insights/ad_insights/daily_reports_get
type DailyReport struct {
	ID     uint   `gorm:"primarykey" json:"id"`
	ReportID int64 `gorm:"uniqueIndex:idx_account_level_date;not null;comment:报表记录ID" json:"report_id"`

	// 账号与时间
	AccountID string    `gorm:"uniqueIndex:idx_account_level_date;index;size:64;not null;comment:广告主账号ID" json:"account_id"`
	Level     string    `gorm:"uniqueIndex:idx_account_level_date;size:50;not null;comment:报表级别" json:"level"`
	Date      time.Time `gorm:"uniqueIndex:idx_account_level_date;index;not null;comment:日期" json:"date"`

	// 结构字段
	CampaignID     string `gorm:"index;size:64;comment:推广计划ID" json:"campaign_id"`
	CampaignName   string `gorm:"size:500;comment:推广计划名称" json:"campaign_name"`
	ProjectID      string `gorm:"index;size:64;comment:项目ID" json:"project_id"`
	ProjectName    string `gorm:"size:500;comment:项目名称" json:"project_name"`
	AdgroupID      string `gorm:"index;size:64;comment:广告组ID" json:"adgroup_id"`
	AdgroupName    string `gorm:"size:500;comment:广告组名称" json:"adgroup_name"`
	AdID           string `gorm:"index;size:64;comment:广告ID" json:"ad_id"`
	AdName         string `gorm:"size:500;comment:广告名称" json:"ad_name"`
	PromotedObject string `gorm:"size:255;comment:推广目标" json:"promoted_object"`

	// 曝光与点击
	ViewCount  int64   `gorm:"default:0;comment:曝光量" json:"view_count"`
	ClickCount int64   `gorm:"default:0;comment:点击量" json:"click_count"`
	ClickRate  float64 `gorm:"type:decimal(10,4);comment:点击率" json:"click_rate"`

	// 花费
	Spend       float64 `gorm:"type:decimal(15,2);default:0;comment:花费(元)" json:"spend"`
	AvgCost     float64 `gorm:"type:decimal(10,2);comment:平均成本" json:"avg_cost"`
	Cpc         float64 `gorm:"type:decimal(10,2);comment:点击单价" json:"cpc"`
	Cpm         float64 `gorm:"type:decimal(10,2);comment:千次曝光成本" json:"cpm"`

	// 转化
	ConvertCount     int64   `gorm:"default:0;comment:转化数" json:"convert_count"`
	ConvertRate      float64 `gorm:"type:decimal(10,4);comment:转化率" json:"convert_rate"`
	CostPerConvert   float64 `gorm:"type:decimal(10,2);comment:单个转化成本" json:"cost_per_convert"`

	// 深度转化
	DeepConvertCount int64   `gorm:"default:0;comment:深度转化数" json:"deep_convert_count"`
	DeepConvertRate float64 `gorm:"type:decimal(10,4);comment:深度转化率" json:"deep_convert_rate"`

	// 互动指标
	LikeCount    int64 `gorm:"default:0;comment:点赞数" json:"like_count"`
	CommentCount int64 `gorm:"default:0;comment:评论数" json:"comment_count"`
	ShareCount   int64 `gorm:"default:0;comment:分享数" json:"share_count"`
	FollowCount  int64 `gorm:"default:0;comment:关注数" json:"follow_count"`

	// 抖音互动
	PlayCount         int64   `gorm:"default:0;comment:播放量" json:"play_count"`
	PlayDuration      int64   `gorm:"default:0;comment:播放时长(秒)" json:"play_duration"`
	StayRate          float64 `gorm:"type:decimal(10,4);comment:停留率" json:"stay_rate"`
	EffectivePlayCount int64  `gorm:"default:0;comment:有效播放量" json:"effective_play_count"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName 指定表名
func (DailyReport) TableName() string {
	return "daily_reports"
}

// HourlyReport 小时报表 - 对应腾讯广告 hourly_reports/get 接口
// 参考: https://developers.e.qq.com/docs/api/insights/ad_insights/hourly_reports_get
type HourlyReport struct {
	ID     uint   `gorm:"primarykey" json:"id"`
	ReportID int64 `gorm:"uniqueIndex:idx_account_level_datetime;not null;comment:报表记录ID" json:"report_id"`

	// 账号与时间
	AccountID string    `gorm:"uniqueIndex:idx_account_level_datetime;index;size:64;not null;comment:广告主账号ID" json:"account_id"`
	Level    string    `gorm:"uniqueIndex:idx_account_level_datetime;size:50;not null;comment:报表级别" json:"level"`
	Date     time.Time `gorm:"uniqueIndex:idx_account_level_datetime;index;not null;comment:日期" json:"date"`
	Hour     int       `gorm:"uniqueIndex:idx_account_level_datetime;not null;comment:小时(0-23)" json:"hour"`
	Datetime time.Time `gorm:"uniqueIndex:idx_account_level_datetime;not null;comment:完整时间" json:"datetime"`

	// 结构字段
	CampaignID   string `gorm:"index;size:64;comment:推广计划ID" json:"campaign_id"`
	CampaignName string `gorm:"size:500;comment:推广计划名称" json:"campaign_name"`
	AdgroupID    string `gorm:"index;size:64;comment:广告组ID" json:"adgroup_id"`
	AdgroupName  string `gorm:"size:500;comment:广告组名称" json:"adgroup_name"`
	AdID         string `gorm:"index;size:64;comment:广告ID" json:"ad_id"`
	AdName       string `gorm:"size:500;comment:广告名称" json:"ad_name"`

	// 曝光与点击
	ViewCount  int64   `gorm:"default:0;comment:曝光量" json:"view_count"`
	ClickCount int64   `gorm:"default:0;comment:点击量" json:"click_count"`
	ClickRate  float64 `gorm:"type:decimal(10,4);comment:点击率" json:"click_rate"`

	// 花费
	Spend float64 `gorm:"type:decimal(15,2);default:0;comment:花费(元)" json:"spend"`
	Cpc   float64 `gorm:"type:decimal(10,2);comment:点击单价" json:"cpc"`
	Cpm   float64 `gorm:"type:decimal(10,2);comment:千次曝光成本" json:"cpm"`

	// 转化
	ConvertCount   int64   `gorm:"default:0;comment:转化数" json:"convert_count"`
	ConvertRate    float64 `gorm:"type:decimal(10,4);comment:转化率" json:"convert_rate"`
	CostPerConvert float64 `gorm:"type:decimal(10,2);comment:单个转化成本" json:"cost_per_convert"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName 指定表名
func (HourlyReport) TableName() string {
	return "hourly_reports"
}

// TargetReport 定向标签报表 - 对应腾讯广告 targeting_tag_reports/get 接口
type TargetReport struct {
	ID     uint   `gorm:"primarykey" json:"id"`
	ReportID int64 `gorm:"uniqueIndex:idx_account_date_target;not null;comment:报表记录ID" json:"report_id"`

	// 账号与时间
	AccountID string    `gorm:"uniqueIndex:idx_account_date_target;index;size:64;not null;comment:广告主账号ID" json:"account_id"`
	Date      time.Time `gorm:"uniqueIndex:idx_account_date_target;index;not null;comment:日期" json:"date"`

	// 定向维度
	TargetID   string `gorm:"uniqueIndex:idx_account_date_target;index;size:64;comment:定向ID" json:"target_id"`
	TargetName string `gorm:"size:255;comment:定向名称" json:"target_name"`
	TargetType string `gorm:"uniqueIndex:idx_account_date_target;size:50;comment:定向类型(GENDER/AGE/RESIDENCE/INTEREST)" json:"target_type"`

	// 结构字段
	AdgroupID   string `gorm:"index;size:64;comment:广告组ID" json:"adgroup_id"`
	AdgroupName string `gorm:"size:500;comment:广告组名称" json:"adgroup_name"`

	// 效果指标
	ViewCount      int64   `gorm:"default:0;comment:曝光量" json:"view_count"`
	ClickCount     int64   `gorm:"default:0;comment:点击量" json:"click_count"`
	ClickRate      float64 `gorm:"type:decimal(10,4);comment:点击率" json:"click_rate"`
	Spend          float64 `gorm:"type:decimal(15,2);default:0;comment:花费(元)" json:"spend"`
	ConvertCount   int64   `gorm:"default:0;comment:转化数" json:"convert_count"`
	ConvertRate    float64 `gorm:"type:decimal(10,4);comment:转化率" json:"convert_rate"`
	CostPerConvert float64 `gorm:"type:decimal(10,2);comment:单个转化成本" json:"cost_per_convert"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName 指定表名
func (TargetReport) TableName() string {
	return "target_reports"
}

// ReportTask 异步报表任务表
type ReportTask struct {
	ID          uint         `gorm:"primarykey" json:"id"`
	TaskID      string       `gorm:"uniqueIndex;size:64;not null;comment:任务ID" json:"task_id"`
	AccountID   string       `gorm:"index;size:64;not null;comment:广告主账号ID" json:"account_id"`
	ReportType  ReportType   `gorm:"not null;comment:报表类型(1日报,2小时报,3定向)" json:"report_type"`
	Level       string       `gorm:"size:50;comment:报表级别" json:"level"`
	Status      ReportStatus `gorm:"default:0;comment:状态(0待处理,1处理中,2已完成,3失败)" json:"status"`
	FilePath    string       `gorm:"type:text;comment:导出文件路径" json:"file_path"`
	RecordCount int64        `gorm:"default:0;comment:记录数" json:"record_count"`
	ErrorMsg    string       `gorm:"type:text;comment:错误信息" json:"error_msg"`
	StartDate   time.Time    `gorm:"comment:开始日期" json:"start_date"`
	EndDate     time.Time    `gorm:"comment:结束日期" json:"end_date"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
	CompletedAt *time.Time   `gorm:"comment:完成时间" json:"completed_at"`
}

// TableName 指定表名
func (ReportTask) TableName() string {
	return "report_tasks"
}
