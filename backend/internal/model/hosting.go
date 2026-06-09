package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

// ========================================
// 智能托管 - 数据模型
// ========================================

// ---------- 托管规则 ----------

// TriggerCondition 触发条件（JSON 存储）
type TriggerCondition struct {
	Type     string  `json:"type"`     // 条件类型: cost_control/budget_manage/effect_optimize/risk_alert
	Metric   string  `json:"metric"`   // 监控指标: conversion_cost/cpc/daily_spend/impressions/conversions/budget_ratio
	Operator string  `json:"operator"` // 比较运算符: gt/gte/lt/lte/eq
	Threshold float64 `json:"threshold"` // 阈值
	Duration int     `json:"duration"` // 持续时间(分钟)// 触发需持续多久
	// 特殊条件
	TimeRange   string `json:"time_range,omitempty"`    // 时间范围: 0-6, 表示凌晨0-6点
	AdMinDays   int    `json:"ad_min_days,omitempty"`   // 广告最少投放天数
	MaxConvCount int   `json:"max_conv_count,omitempty"` // 最大转化数
	// 状态变化条件
	StatusField string `json:"status_field,omitempty"` // 状态字段
	StatusValue string `json:"status_value,omitempty"` // 状态值
}

// Scan 实现 sql.Scanner 接口
func (c *TriggerCondition) Scan(value any) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, c)
}

// Value 实现 driver.Valuer 接口
func (c TriggerCondition) Value() (driver.Value, error) {
	return json.Marshal(c)
}

// ExecutionAction 执行动作（JSON 存储）
type ExecutionAction struct {
	Type       string  `json:"type"`                  // 动作类型: pause_ad/adjust_bid/raise_budget/notify/resume_ad/quick_start
	TargetID   string  `json:"target_id,omitempty"`   // 目标ID
	TargetType string  `json:"target_type,omitempty"` // 目标类型: ad/adgroup/campaign
	// 调整参数
	BidAdjustRatio    float64 `json:"bid_adjust_ratio,omitempty"`    // 出价调整比例(1=不变,0.9=降10%)
	BudgetRaiseRatio  float64 `json:"budget_raise_ratio,omitempty"`  // 预算提升比例
	BudgetRaiseAmount float64 `json:"budget_raise_amount,omitempty"` // 预算提升固定金额(分)
	// 通知参数
	NotifyChannels []string `json:"notify_channels,omitempty"` // 通知渠道: email/sms/dingtalk/feishu
	NotifyUsers    []string `json:"notify_users,omitempty"`    // 通知用户列表
	Message        string   `json:"message,omitempty"`         // 自定义通知内容
}

// Scan 实现 sql.Scanner 接口
func (a *ExecutionAction) Scan(value any) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, a)
}

// Value 实现 driver.Valuer 接口
func (a ExecutionAction) Value() (driver.Value, error) {
	return json.Marshal(a)
}

// HostingRule 智能托管规则表
type HostingRule struct {
	ID            uint             `gorm:"primarykey" json:"id"`
	RuleName      string           `gorm:"size:100;not null;comment:规则名称" json:"rule_name"`
	Category      string           `gorm:"size:50;not null;index;comment:规则分类(cost_control/budget_manage/effect_optimize/risk_alert)" json:"category"`
	Scene         string           `gorm:"size:100;not null;comment:托管场景" json:"scene"`
	Description   string           `gorm:"size:500;comment:规则描述" json:"description"`
	Status        int8             `gorm:"default:1;index;comment:状态(1=启用,0=禁用)" json:"status"`
	Priority      int8             `gorm:"default:5;comment:优先级(1-10,数字越大优先级越高)" json:"priority"`

	// 触发条件（JSON存储）
	TriggerCondition *TriggerCondition `gorm:"type:json;comment:触发条件" json:"trigger_condition"`

	// 执行动作（JSON存储）
	ExecutionAction *ExecutionAction `gorm:"type:json;comment:执行动作" json:"execution_action"`

	// 作用范围
	AccountIDs   string `gorm:"type:text;comment:作用账号ID列表(JSON数组)" json:"account_ids"`
	CampaignIDs  string `gorm:"type:text;comment:作用广告系列ID列表(JSON数组)" json:"campaign_ids"`
	AdGroupIDs   string `gorm:"type:text;comment:作用广告组ID列表(JSON数组)" json:"ad_group_ids"`
	AdIDs        string `gorm:"type:text;comment:作用广告ID列表(JSON数组)" json:"ad_ids"`

	// 执行限制
	CooldownMinutes      int `gorm:"default:60;comment:冷却时间(分钟)" json:"cooldown_minutes"`
	MaxExecutionsPerDay  int `gorm:"default:10;comment:每日最大执行次数" json:"max_executions_per_day"`

	// 统计
	TodayExecCount int `gorm:"default:0;comment:今日执行次数" json:"today_exec_count"`
	TotalExecCount int `gorm:"default:0;comment:累计执行次数" json:"total_exec_count"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (HostingRule) TableName() string {
	return "hosting_rules"
}

// ---------- 托管执行记录 ----------

// ExecutionStatus 执行状态
const (
	ExecStatusSuccess = 1 // 执行成功
	ExecStatusFailed  = 2 // 执行失败
	ExecStatusPending = 3 // 待执行
	ExecStatusRolled  = 4 // 已回滚
)

// HostingExecution 托管执行记录表
type HostingExecution struct {
	ID            uint      `gorm:"primarykey" json:"id"`
	RuleID        uint      `gorm:"not null;index;comment:规则ID" json:"rule_id"`
	RuleName      string    `gorm:"size:100;comment:规则名称" json:"rule_name"`
	AccountID     string    `gorm:"size:64;index;comment:账号ID" json:"account_id"`
	TargetID      string    `gorm:"size:64;comment:目标ID" json:"target_id"`
	TargetType    string    `gorm:"size:20;comment:目标类型(ad/adgroup/campaign)" json:"target_type"`
	ActionType    string    `gorm:"size:50;comment:动作类型" json:"action_type"`

	// 触发时数据快照
	TriggerSnapshot string `gorm:"type:text;comment:触发时的指标快照(JSON)" json:"trigger_snapshot"`

	// 执行参数
	ActionParams string `gorm:"type:text;comment:执行参数(JSON)" json:"action_params"`

	// 回滚支持
	BeforeValue string `gorm:"type:text;comment:执行前的值(用于回滚,JSON)" json:"before_value"`
	AfterValue  string `gorm:"type:text;comment:执行后的值(JSON)" json:"after_value"`

	Status      int8      `gorm:"default:3;index;comment:执行状态(1=成功,2=失败,3=待执行,4=已回滚)" json:"status"`
	ErrorMsg    string    `gorm:"type:text;comment:错误信息" json:"error_msg"`
	APIRawResp  string    `gorm:"type:text;comment:API原始响应" json:"api_raw_resp"`
	RollbackAt  *time.Time `gorm:"comment:回滚时间" json:"rollback_at"`
	ExecutedAt  *time.Time `gorm:"comment:执行时间" json:"executed_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func (HostingExecution) TableName() string {
	return "hosting_executions"
}

// ---------- 托管告警通知 ----------

// AlertSeverity 告警级别
const (
	AlertSeverityLow      = 1 // 低
	AlertSeverityMedium   = 2 // 中
	AlertSeverityHigh     = 3 // 高
	AlertSeverityCritical = 4 // 紧急
)

// AlertStatus 告警状态
const (
	AlertStatusUnread   = 1 // 未读
	AlertStatusRead     = 2 // 已读
	AlertStatusHandled  = 3 // 已处理
	AlertStatusIgnored  = 4 // 已忽略
)

// AlertType 告警类型
const (
	AlertTypeAdRejected   = "ad_rejected"    // 广告拒审
	AlertTypeAdAbnormal   = "ad_abnormal"    // 广告异常
	AlertTypeCompensation = "compensation"   // 赔付状态变动
	AlertTypeCostSpike    = "cost_spike"     // 成本飙升
	AlertTypeBudgetLimit  = "budget_limit"   // 预算触顶
)

// HostingAlert 托管告警通知表
type HostingAlert struct {
	ID           uint       `gorm:"primarykey" json:"id"`
	RuleID       uint       `gorm:"index;comment:关联规则ID" json:"rule_id"`
	ExecutionID  uint       `gorm:"index;comment:关联执行记录ID" json:"execution_id"`
	AccountID    string     `gorm:"size:64;index;comment:账号ID" json:"account_id"`
	AlertType    string     `gorm:"size:50;not null;index;comment:告警类型" json:"alert_type"`
	AlertTitle   string     `gorm:"size:200;not null;comment:告警标题" json:"alert_title"`
	AlertContent string     `gorm:"type:text;comment:告警内容" json:"alert_content"`
	Severity     int8       `gorm:"default:2;comment:严重级别(1=低,2=中,3=高,4=紧急)" json:"severity"`
	Status       int8       `gorm:"default:1;index;comment:状态(1=未读,2=已读,3=已处理,4=已忽略)" json:"status"`

	// 通知渠道与用户
	NotifyChannel string `gorm:"size:50;comment:通知渠道" json:"notify_channel"`
	NotifyUser    string `gorm:"size:100;comment:通知用户" json:"notify_user"`

	// 处理信息
	Handler       string     `gorm:"size:100;comment:处理人" json:"handler"`
	HandleResult  string     `gorm:"type:text;comment:处理结果" json:"handle_result"`
	HandledAt     *time.Time `gorm:"comment:处理时间" json:"handled_at"`
	ReadAt        *time.Time `gorm:"comment:阅读时间" json:"read_at"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

func (HostingAlert) TableName() string {
	return "hosting_alerts"
}

// ---------- 广告性能快照 ----------

// AdPerformanceSnapshot 广告性能快照表
// 用于规则引擎评估时获取最新的广告性能数据
type AdPerformanceSnapshot struct {
	ID           uint      `gorm:"primarykey" json:"id"`
	SnapshotTime time.Time `gorm:"not null;index;comment:快照时间" json:"snapshot_time"`

	// 广告标识
	AccountID  string `gorm:"size:64;not null;index:idx_ad_key;comment:账号ID" json:"account_id"`
	AdID       string `gorm:"size:64;not null;index:idx_ad_key;comment:广告ID" json:"ad_id"`
	AdName     string `gorm:"size:200;comment:广告名称" json:"ad_name"`

	// 广告组/系列信息
	AdGroupID   string `gorm:"size:64;index;comment:广告组ID" json:"adgroup_id"`
	AdGroupName string `gorm:"size:200;comment:广告组名称" json:"adgroup_name"`
	CampaignID  string `gorm:"size:64;index;comment:广告系列ID" json:"campaign_id"`

	// 投放状态
	AdStatus       int8 `gorm:"comment:广告状态" json:"ad_status"`
	DiagnosisStatus int8 `gorm:"comment:诊断状态(0=正常,1=拒审,2=异常)" json:"diagnosis_status"`

	// 核心指标
	Impressions      int64   `gorm:"default:0;comment:曝光量" json:"impressions"`
	Clicks           int64   `gorm:"default:0;comment:点击量" json:"clicks"`
	Conversions      int64   `gorm:"default:0;comment:转化量" json:"conversions"`
	Spend            float64 `gorm:"type:decimal(15,2);default:0;comment:消耗(分)" json:"spend"`
	CTR              float64 `gorm:"type:decimal(10,6);default:0;comment:点击率" json:"ctr"`
	CVR              float64 `gorm:"type:decimal(10,6);default:0;comment:转化率" json:"cvr"`
	CPC              float64 `gorm:"type:decimal(15,6);default:0;comment:单次点击成本(分)" json:"cpc"`
	CPM              float64 `gorm:"type:decimal(15,6);default:0;comment:千次曝光成本(分)" json:"cpm"`
	CostPerConversion float64 `gorm:"type:decimal(15,6);default:0;comment:单次转化成本(分)" json:"cost_per_conversion"`

	// 预算信息
	DailyBudget     float64 `gorm:"type:decimal(15,2);default:0;comment:日预算(分)" json:"daily_budget"`
	DailyBudgetUsed float64 `gorm:"type:decimal(15,2);default:0;comment:日消耗(分)" json:"daily_budget_used"`
	BudgetRatio     float64 `gorm:"type:decimal(10,4);default:0;comment:预算消耗比例" json:"budget_ratio"`

	// 出价信息
	BidAmount float64 `gorm:"type:decimal(15,2);default:0;comment:出价(分)" json:"bid_amount"`
	BidMode   string  `gorm:"size:50;comment:出价模式" json:"bid_mode"`

	// 投放时长(小时)
	DeliveryHours int `gorm:"default:0;comment:已投放时长(小时)" json:"delivery_hours"`
	// 在线天数
	OnlineDays int `gorm:"default:0;comment:在线天数" json:"online_days"`

	// 学习期状态
	IsLearningPhase bool `gorm:"default:false;comment:是否处于学习期" json:"is_learning_phase"`

	// A/B测试
	ABTestGroup string `gorm:"size:50;comment:A/B测试分组" json:"ab_test_group"`

	CreatedAt time.Time `json:"created_at"`
}

func (AdPerformanceSnapshot) TableName() string {
	return "ad_performance_snapshots"
}
