package service

import (
	"encoding/json"
	"fmt"
	"time"

	"marketing-platform/internal/database"
	"marketing-platform/internal/engine"
	"marketing-platform/internal/logger"
	"marketing-platform/internal/model"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// HostingExecutorService 托管执行服务 —— 核心业务逻辑
type HostingExecutorService struct {
	db          *gorm.DB
	ruleEngine  *engine.RuleEngine
	executor    *engine.ActionExecutor
}

// NewHostingExecutorService 创建托管执行服务
func NewHostingExecutorService() *HostingExecutorService {
	return &HostingExecutorService{
		db:         database.GetDB(),
		ruleEngine: engine.NewRuleEngine(),
		executor:   engine.NewActionExecutor(),
	}
}

// ExecuteAllActiveRules 执行所有启用规则的评估和动作
// 由定时调度器调用
func (s *HostingExecutorService) ExecuteAllActiveRules() error {
	logger.Logger.Info("开始执行智能托管规则评估")

	// 1. 获取所有启用规则
	var activeRules []model.HostingRule
	if err := s.db.Where("status = ?", 1).Order("priority DESC").Find(&activeRules).Error; err != nil {
		return fmt.Errorf("获取启用规则失败: %w", err)
	}

	if len(activeRules) == 0 {
		logger.Logger.Info("没有启用的托管规则")
		return nil
	}

	totalExecuted := 0

	// 2. 逐条规则评估
	for i := range activeRules {
		rule := &activeRules[i]

		// 获取规则作用范围内的广告性能快照
		snapshots, err := s.getSnapshotsForRule(rule)
		if err != nil {
			logger.Logger.Error("获取性能快照失败",
				zap.Uint("rule_id", rule.ID),
				zap.Error(err))
			continue
		}

		if len(snapshots) == 0 {
			continue
		}

		// 规则引擎评估
		results, err := s.ruleEngine.EvaluateRule(rule, snapshots)
		if err != nil {
			logger.Logger.Error("规则评估失败",
				zap.Uint("rule_id", rule.ID),
				zap.Error(err))
			continue
		}

		// 执行匹配的动作
		for _, result := range results {
			if !result.Matched {
				continue
			}
			execution, err := s.executor.Execute(result)
			if err != nil {
				logger.Logger.Error("执行动作失败",
					zap.Uint("rule_id", rule.ID),
					zap.Error(err))
			}
			if execution != nil {
				totalExecuted++
			}
		}
	}

	logger.Logger.Info("智能托管规则评估完成",
		zap.Int("rules_evaluated", len(activeRules)),
		zap.Int("actions_executed", totalExecuted))

	return nil
}

// EvaluateSingleRule 评估单条规则（用于手动测试）
func (s *HostingExecutorService) EvaluateSingleRule(ruleID uint) ([]engine.TriggerResult, error) {
	var rule model.HostingRule
	if err := s.db.First(&rule, ruleID).Error; err != nil {
		return nil, fmt.Errorf("规则不存在: %w", err)
	}

	snapshots, err := s.getSnapshotsForRule(&rule)
	if err != nil {
		return nil, err
	}

	return s.ruleEngine.EvaluateRule(&rule, snapshots)
}

// getSnapshotsForRule 获取规则作用范围内的最新性能快照
func (s *HostingExecutorService) getSnapshotsForRule(rule *model.HostingRule) ([]model.AdPerformanceSnapshot, error) {
	query := s.db.Model(&model.AdPerformanceSnapshot{}).
		Where("snapshot_time >= ?", time.Now().Add(-1*time.Hour))

	// 按规则作用范围筛选
	if rule.AccountIDs != "" {
		var accounts []string
		if err := json.Unmarshal([]byte(rule.AccountIDs), &accounts); err == nil && len(accounts) > 0 {
			query = query.Where("account_id IN ?", accounts)
		}
	}
	if rule.AdIDs != "" {
		var adIDs []string
		if err := json.Unmarshal([]byte(rule.AdIDs), &adIDs); err == nil && len(adIDs) > 0 {
			query = query.Where("ad_id IN ?", adIDs)
		}
	}
	if rule.AdGroupIDs != "" {
		var groupIDs []string
		if err := json.Unmarshal([]byte(rule.AdGroupIDs), &groupIDs); err == nil && len(groupIDs) > 0 {
			query = query.Where("adgroup_id IN ?", groupIDs)
		}
	}
	if rule.CampaignIDs != "" {
		var campIDs []string
		if err := json.Unmarshal([]byte(rule.CampaignIDs), &campIDs); err == nil && len(campIDs) > 0 {
			query = query.Where("campaign_id IN ?", campIDs)
		}
	}

	var snapshots []model.AdPerformanceSnapshot
	if err := query.Order("snapshot_time DESC").Limit(1000).Find(&snapshots).Error; err != nil {
		return nil, err
	}

	return snapshots, nil
}

// ======== 执行记录相关 ========

// ListExecutions 获取执行记录列表
func (s *HostingExecutorService) ListExecutions(accountID string, status int, page, pageSize int) ([]model.HostingExecution, int64, error) {
	var executions []model.HostingExecution
	var total int64

	query := s.db.Model(&model.HostingExecution{})
	if accountID != "" {
		query = query.Where("account_id = ?", accountID)
	}
	if status > 0 {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&executions).Error; err != nil {
		return nil, 0, err
	}

	return executions, total, nil
}

// GetExecutionByID 获取执行记录详情
func (s *HostingExecutorService) GetExecutionByID(id uint) (*model.HostingExecution, error) {
	var execution model.HostingExecution
	if err := s.db.First(&execution, id).Error; err != nil {
		return nil, err
	}
	return &execution, nil
}

// RollbackExecution 回滚执行
func (s *HostingExecutorService) RollbackExecution(id uint) error {
	return s.executor.RollbackExecution(id)
}

// ======== 看板统计 ========

// DashboardStats 看板统计数据
type DashboardStats struct {
	ActiveRules    int64 `json:"active_rules"`
	TotalRules     int64 `json:"total_rules"`
	TodayExec      int64 `json:"today_exec"`
	TotalExec      int64 `json:"total_exec"`
	SuccessExec    int64 `json:"success_exec"`
	FailedExec     int64 `json:"failed_exec"`
	UnreadAlerts   int64 `json:"unread_alerts"`
	TotalAlerts    int64 `json:"total_alerts"`
}

// GetDashboardStats 获取看板统计
func (s *HostingExecutorService) GetDashboardStats() (*DashboardStats, error) {
	stats := &DashboardStats{}

	// 活跃规则数
	s.db.Model(&model.HostingRule{}).Where("status = ?", 1).Count(&stats.ActiveRules)
	// 总规则数
	s.db.Model(&model.HostingRule{}).Count(&stats.TotalRules)
	// 今日执行次数
	today := time.Now().Format("2006-01-02")
	s.db.Model(&model.HostingExecution{}).Where("DATE(created_at) = ?", today).Count(&stats.TodayExec)
	// 累计执行次数
	s.db.Model(&model.HostingExecution{}).Count(&stats.TotalExec)
	// 成功次数
	s.db.Model(&model.HostingExecution{}).Where("status = ?", model.ExecStatusSuccess).Count(&stats.SuccessExec)
	// 失败次数
	s.db.Model(&model.HostingExecution{}).Where("status = ?", model.ExecStatusFailed).Count(&stats.FailedExec)
	// 未读告警
	s.db.Model(&model.HostingAlert{}).Where("status = ?", model.AlertStatusUnread).Count(&stats.UnreadAlerts)
	// 总告警
	s.db.Model(&model.HostingAlert{}).Count(&stats.TotalAlerts)

	return stats, nil
}

// TrendData 趋势数据
type TrendData struct {
	Date  string `json:"date"`
	Count int64  `json:"count"`
}

// GetExecutionTrend 获取执行趋势(近7天)
func (s *HostingExecutorService) GetExecutionTrend() ([]TrendData, error) {
	var results []TrendData
	startDate := time.Now().AddDate(0, 0, -6).Format("2006-01-02")

	err := s.db.Raw(`SELECT DATE(created_at) as date, COUNT(*) as count
		FROM hosting_executions
		WHERE DATE(created_at) >= ?
		GROUP BY DATE(created_at)
		ORDER BY date`, startDate).Scan(&results).Error

	return results, err
}

// ======== 数据快照 ========

// SavePerformanceSnapshot 保存广告性能快照
func (s *HostingExecutorService) SavePerformanceSnapshot(snapshot *model.AdPerformanceSnapshot) error {
	return s.db.Create(snapshot).Error
}

// SavePerformanceSnapshots 批量保存广告性能快照
func (s *HostingExecutorService) SavePerformanceSnapshots(snapshots []model.AdPerformanceSnapshot) error {
	if len(snapshots) == 0 {
		return nil
	}
	return s.db.CreateInBatches(snapshots, 200).Error
}

// CleanOldSnapshots 清理旧的性能快照（保留最近7天）
func (s *HostingExecutorService) CleanOldSnapshots() error {
	cutoff := time.Now().AddDate(0, 0, -7)
	result := s.db.Where("snapshot_time < ?", cutoff).Delete(&model.AdPerformanceSnapshot{})
	logger.Logger.Info("清理旧性能快照完成",
		zap.Int64("deleted", result.RowsAffected))
	return result.Error
}

// ======== 数据采集辅助 ========

// CollectPerformanceSnapshots 从现有报表数据生成快照
// 从 daily_reports 和 ads 表中提取数据生成 AdPerformanceSnapshot
func (s *HostingExecutorService) CollectPerformanceSnapshots() error {
	logger.Logger.Info("开始采集广告性能快照")

	// 获取最近有消耗的广告
	type AdWithReport struct {
		AccountID         string
		AdID              string
		AdName            string
		AdGroupID         string
		AdGroupName       string
		CampaignID        string
		AdStatus          int8
		BidAmount         float64
		BidMode           string
		Impressions       int64
		Clicks            int64
		Conversions       int64
		Spend             float64
		CTR               float64
		CVR               float64
		CPC               float64
		CPM               float64
		CostPerConversion float64
		DailyBudget       float64
		ABTestGroup       string
	}

	var rows []AdWithReport
	err := s.db.Raw(`
		SELECT 
			a.account_id,
			a.ad_id,
			a.ad_name,
			a.adgroup_id,
			ag.adgroup_name,
			ag.campaign_id,
			a.ad_status,
			ag.bid_amount,
			'' as bid_mode,
			COALESCE(SUM(dr.view_count), 0) as impressions,
			COALESCE(SUM(dr.click_count), 0) as clicks,
			COALESCE(SUM(dr.conversion_count), 0) as conversions,
			COALESCE(SUM(dr.spend), 0) as spend,
			0 as ctr,
			0 as cvr,
			CASE WHEN SUM(dr.click_count) > 0 THEN SUM(dr.spend) / SUM(dr.click_count) ELSE 0 END as cpc,
			CASE WHEN SUM(dr.view_count) > 0 THEN SUM(dr.spend) / SUM(dr.view_count) * 1000 ELSE 0 END as cpm,
			CASE WHEN SUM(dr.conversion_count) > 0 THEN SUM(dr.spend) / SUM(dr.conversion_count) ELSE 0 END as cost_per_conversion,
			COALESCE(c.daily_budget, 0) as daily_budget,
			'' as ab_test_group
		FROM ads a
		LEFT JOIN ad_groups ag ON a.adgroup_id = ag.adgroup_id
		LEFT JOIN campaigns c ON ag.campaign_id = c.campaign_id
		LEFT JOIN daily_reports dr ON a.account_id = dr.account_id AND dr.date >= ?
		GROUP BY a.account_id, a.ad_id, a.ad_name, a.adgroup_id, ag.adgroup_name, ag.campaign_id, 
		         a.ad_status, ag.bid_amount, c.daily_budget
		HAVING spend > 0 OR impressions > 0
	`, time.Now().AddDate(0, 0, -1).Format("2006-01-02")).Scan(&rows).Error

	if err != nil {
		return fmt.Errorf("查询报表数据失败: %w", err)
	}

	now := time.Now()
	var snapshots []model.AdPerformanceSnapshot

	for _, row := range rows {
		snapshot := model.AdPerformanceSnapshot{
			SnapshotTime:      now,
			AccountID:         row.AccountID,
			AdID:              row.AdID,
			AdName:            row.AdName,
			AdGroupID:         row.AdGroupID,
			AdGroupName:       row.AdGroupName,
			CampaignID:        row.CampaignID,
			AdStatus:          row.AdStatus,
			Impressions:       row.Impressions,
			Clicks:            row.Clicks,
			Conversions:       row.Conversions,
			Spend:             row.Spend,
			CTR:               s.calcCTR(row.Clicks, row.Impressions),
			CVR:               s.calcCVR(row.Conversions, row.Clicks),
			CPC:               row.CPC,
			CPM:               row.CPM,
			CostPerConversion: row.CostPerConversion,
			DailyBudget:       row.DailyBudget,
			DailyBudgetUsed:   row.Spend,
			BudgetRatio:       s.calcBudgetRatio(row.Spend, row.DailyBudget),
			BidAmount:         row.BidAmount,
			BidMode:           row.BidMode,
			ABTestGroup:       row.ABTestGroup,
		}

		// 估算在线天数（简化，从创建时间到现在的天数）
		var ad model.Ad
		if err := s.db.Where("ad_id = ?", row.AdID).First(&ad).Error; err == nil {
			snapshot.OnlineDays = int(time.Since(ad.CreatedAt).Hours() / 24)
			snapshot.DeliveryHours = int(time.Since(ad.CreatedAt).Hours())
		}

		snapshots = append(snapshots, snapshot)
	}

	if err := s.SavePerformanceSnapshots(snapshots); err != nil {
		return fmt.Errorf("保存快照失败: %w", err)
	}

	logger.Logger.Info("广告性能快照采集完成", zap.Int("count", len(snapshots)))
	return nil
}

func (s *HostingExecutorService) calcCTR(clicks, impressions int64) float64 {
	if impressions == 0 {
		return 0
	}
	return float64(clicks) / float64(impressions)
}

func (s *HostingExecutorService) calcCVR(conversions, clicks int64) float64 {
	if clicks == 0 {
		return 0
	}
	return float64(conversions) / float64(clicks)
}

func (s *HostingExecutorService) calcBudgetRatio(spend, budget float64) float64 {
	if budget == 0 {
		return 0
	}
	return spend / budget
}

// CheckAdStatuses 检查广告状态（用于风险预警）
func (s *HostingExecutorService) CheckAdStatuses() error {
	logger.Logger.Info("开始检查广告状态")

	// 从快照中读取诊断状态异常的广告
	alerts := []model.HostingAlert{}

	var abnormalAds []model.AdPerformanceSnapshot
	s.db.Where("diagnosis_status != ?", 0).Find(&abnormalAds)

	for _, ad := range abnormalAds {
		alert := model.HostingAlert{
			AccountID:    ad.AccountID,
			AlertType:    model.AlertTypeAdAbnormal,
			AlertTitle:   fmt.Sprintf("广告异常: %s", ad.AdName),
			AlertContent: fmt.Sprintf("广告 %s (%s) 诊断状态异常(状态码: %d)", ad.AdName, ad.AdID, ad.DiagnosisStatus),
			Severity:     model.AlertSeverityHigh,
			Status:       model.AlertStatusUnread,
		}
		alerts = append(alerts, alert)
	}

	if len(alerts) > 0 {
		if err := s.db.Create(&alerts).Error; err != nil {
			return fmt.Errorf("保存告警记录失败: %w", err)
		}
		logger.Logger.Info("广告状态检查完成，发现异常", zap.Int("count", len(alerts)))
	} else {
		logger.Logger.Info("广告状态检查完成，未发现异常")
	}

	return nil
}

// CleanupOldRecords 清理旧记录
func (s *HostingExecutorService) CleanupOldRecords(days int) error {
	cutoff := time.Now().AddDate(0, 0, -days)

	// 清理旧执行记录
	if err := s.db.Where("created_at < ?", cutoff).Delete(&model.HostingExecution{}).Error; err != nil {
		return fmt.Errorf("清理旧执行记录失败: %w", err)
	}

	// 清理已处理的告警
	if err := s.db.Where("created_at < ? AND status IN ?", cutoff, []int8{model.AlertStatusHandled, model.AlertStatusIgnored}).
		Delete(&model.HostingAlert{}).Error; err != nil {
		return fmt.Errorf("清理旧告警失败: %w", err)
	}

	logger.Logger.Info("清理旧记录完成", zap.Int("days", days))
	return nil
}
