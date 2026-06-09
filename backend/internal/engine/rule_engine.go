package engine

import (
	"fmt"
	"time"

	"marketing-platform/internal/database"
	"marketing-platform/internal/logger"
	"marketing-platform/internal/model"

	"go.uber.org/zap"
)

// RuleEngine 规则引擎 —— 负责评估规则触发条件
type RuleEngine struct{}

// NewRuleEngine 创建规则引擎
func NewRuleEngine() *RuleEngine {
	return &RuleEngine{}
}

// EvaluateRule 评估单条规则是否满足触发条件
// rule: 规则定义
// snapshots: 该规则作用范围内的广告性能快照列表
func (e *RuleEngine) EvaluateRule(rule *model.HostingRule, snapshots []model.AdPerformanceSnapshot) ([]TriggerResult, error) {
	if rule.Status == 0 {
		logger.Logger.Debug("规则已禁用，跳过评估", zap.Uint("rule_id", rule.ID))
		return nil, nil
	}

	// 检查冷却时间
	hasCooldown := e.checkCooldown(rule)
	if hasCooldown {
		logger.Logger.Debug("规则尚在冷却期", zap.Uint("rule_id", rule.ID))
		return nil, nil
	}

	// 检查今日执行次数上限
	reachedLimit := e.checkDailyLimit(rule)
	if reachedLimit {
		logger.Logger.Debug("规则已达到今日执行上限", zap.Uint("rule_id", rule.ID))
		return nil, nil
	}

	var results []TriggerResult
	condition := rule.TriggerCondition
	if condition == nil {
		return nil, fmt.Errorf("规则 %d 的触发条件为空", rule.ID)
	}

	switch condition.Type {
	case "cost_control":
		results = e.evaluateCostControl(rule, snapshots, condition)
	case "budget_manage":
		results = e.evaluateBudgetManage(rule, snapshots, condition)
	case "effect_optimize":
		results = e.evaluateEffectOptimize(rule, snapshots, condition)
	case "risk_alert":
		results = e.evaluateRiskAlert(snapshots, condition)
	default:
		logger.Logger.Warn("未知的规则类型", zap.String("type", condition.Type))
	}

	return results, nil
}

// TriggerResult 触发结果
type TriggerResult struct {
	AccountID  string                       `json:"account_id"`
	TargetID   string                       `json:"target_id"`
	TargetType string                       `json:"target_type"`
	Snapshot   *model.AdPerformanceSnapshot `json:"snapshot"`
	Rule       *model.HostingRule           `json:"rule"`
	Reason     string                       `json:"reason"`
	Matched    bool                         `json:"matched"`
}

// ---- 成本控制 ----

func (e *RuleEngine) evaluateCostControl(rule *model.HostingRule, snapshots []model.AdPerformanceSnapshot, condition *model.TriggerCondition) []TriggerResult {
	var results []TriggerResult

	for i := range snapshots {
		sn := &snapshots[i]

		// 场景1: 转化成本超出预期
		if condition.Metric == "conversion_cost" && sn.CostPerConversion > 0 {
			if e.compareFloat(sn.CostPerConversion, condition.Threshold, condition.Operator) {
				results = append(results, TriggerResult{
					AccountID:  sn.AccountID,
					TargetID:   sn.AdID,
					TargetType: "ad",
					Snapshot:   sn,
					Rule:       rule,
					Reason:     fmt.Sprintf("转化成本 %.2f > 目标 %.2f", sn.CostPerConversion, condition.Threshold),
					Matched:    true,
				})
			}
		}

		// 场景2: 凌晨时段成本失控
		if condition.Metric == "cpc" && sn.CPC > 0 {
			now := time.Now()
			hour := now.Hour()
			if condition.TimeRange != "" {
				// 简化处理: "0-6" 表示凌晨 0-6 点
				if hour >= 0 && hour < 6 {
					if e.compareFloat(sn.CPC, condition.Threshold, condition.Operator) {
						results = append(results, TriggerResult{
							AccountID:  sn.AccountID,
							TargetID:   sn.AdID,
							TargetType: "ad",
							Snapshot:   sn,
							Rule:       rule,
							Reason:     fmt.Sprintf("凌晨时段 CPC %.4f > 目标 %.4f", sn.CPC, condition.Threshold),
							Matched:    true,
						})
					}
				}
			}
		}
	}

	return results
}

// ---- 预算管理 ----

func (e *RuleEngine) evaluateBudgetManage(rule *model.HostingRule, snapshots []model.AdPerformanceSnapshot, condition *model.TriggerCondition) []TriggerResult {
	var results []TriggerResult

	for i := range snapshots {
		sn := &snapshots[i]

		// 场景3: 日消耗触顶
		if condition.Metric == "budget_ratio" && sn.BudgetRatio > 0 {
			if e.compareFloat(sn.BudgetRatio, condition.Threshold, condition.Operator) {
				results = append(results, TriggerResult{
					AccountID:  sn.AccountID,
					TargetID:   sn.AdGroupID,
					TargetType: "adgroup",
					Snapshot:   sn,
					Rule:       rule,
					Reason:     fmt.Sprintf("日预算消耗比例 %.2f%% >= %.2f%%", sn.BudgetRatio*100, condition.Threshold*100),
					Matched:    true,
				})
			}
		}

		// 场景4: 消耗数据异常(学习期内消耗未达标)
		if condition.Metric == "impressions" && sn.IsLearningPhase {
			if !e.compareFloat(float64(sn.Impressions), condition.Threshold, condition.Operator) {
				results = append(results, TriggerResult{
					AccountID:  sn.AccountID,
					TargetID:   sn.AdID,
					TargetType: "ad",
					Snapshot:   sn,
					Rule:       rule,
					Reason:     fmt.Sprintf("学习期内曝光 %d 未达标(阈值 %.0f)", sn.Impressions, condition.Threshold),
					Matched:    true,
				})
			}
		}
		// 消耗不达标
		if condition.Metric == "spend" && sn.IsLearningPhase {
			if !e.compareFloat(sn.Spend, condition.Threshold, condition.Operator) {
				results = append(results, TriggerResult{
					AccountID:  sn.AccountID,
					TargetID:   sn.AdID,
					TargetType: "ad",
					Snapshot:   sn,
					Rule:       rule,
					Reason:     fmt.Sprintf("学习期内消耗 %.2f 未达标(阈值 %.2f)", sn.Spend, condition.Threshold),
					Matched:    true,
				})
			}
		}
	}

	return results
}

// ---- 效果优化 ----

func (e *RuleEngine) evaluateEffectOptimize(rule *model.HostingRule, snapshots []model.AdPerformanceSnapshot, condition *model.TriggerCondition) []TriggerResult {
	var results []TriggerResult

	for i := range snapshots {
		sn := &snapshots[i]

		// 场景5: 低效广告自动关停
		if condition.Metric == "conversions" && sn.OnlineDays >= condition.AdMinDays {
			if condition.MaxConvCount > 0 && int(sn.Conversions) < condition.MaxConvCount {
				// 同时检查成本是否超标
				if condition.Threshold > 0 && sn.CostPerConversion > condition.Threshold {
					results = append(results, TriggerResult{
						AccountID:  sn.AccountID,
						TargetID:   sn.AdID,
						TargetType: "ad",
						Snapshot:   sn,
						Rule:       rule,
						Reason:     fmt.Sprintf("投放%d天，转化%d < %d，成本%.2f > %.2f", sn.OnlineDays, sn.Conversions, condition.MaxConvCount, sn.CostPerConversion, condition.Threshold),
						Matched:    true,
					})
				}
			}
		}

		// 场景7: 新广告起量期（24-48小时内）
		if condition.Metric == "delivery_hours" && sn.DeliveryHours >= 24 && sn.DeliveryHours <= 48 {
			results = append(results, TriggerResult{
				AccountID:  sn.AccountID,
				TargetID:   sn.AdID,
				TargetType: "ad",
				Snapshot:   sn,
				Rule:       rule,
				Reason:     fmt.Sprintf("新广告投放%d小时，处于起量期", sn.DeliveryHours),
				Matched:    true,
			})
		}
	}

	// 场景6: A/B 测试判定更优广告
	if len(snapshots) >= 2 {
		abResults := e.evaluateABTest(rule, snapshots)
		results = append(results, abResults...)
	}

	return results
}

// evaluateABTest 评估A/B测试中的优胜广告
func (e *RuleEngine) evaluateABTest(rule *model.HostingRule, snapshots []model.AdPerformanceSnapshot) []TriggerResult {
	var results []TriggerResult
	// 简单实现：比较同广告组的广告，找CVR最高且成本最低的
	groups := make(map[string][]int)
	for i, sn := range snapshots {
		if sn.ABTestGroup != "" {
			groups[sn.ABTestGroup] = append(groups[sn.ABTestGroup], i)
		}
	}

	for _, indices := range groups {
		if len(indices) < 2 {
			continue
		}
		var bestIdx int
		bestScore := 0.0
		for _, idx := range indices {
			sn := &snapshots[idx]
			score := sn.CVR - sn.CostPerConversion/10000 // 简单加权
			if score > bestScore {
				bestScore = score
				bestIdx = idx
			}
		}
		sn := &snapshots[bestIdx]
		results = append(results, TriggerResult{
			AccountID:  sn.AccountID,
			TargetID:   sn.AdID,
			TargetType: "ad",
			Snapshot:   sn,
			Rule:       rule,
			Reason:     fmt.Sprintf("A/B测试中表现最优: CVR=%.4f, CPA=%.2f", sn.CVR, sn.CostPerConversion),
			Matched:    true,
		})
	}
	return results
}

// ---- 风险预警 ----

func (e *RuleEngine) evaluateRiskAlert(snapshots []model.AdPerformanceSnapshot, condition *model.TriggerCondition) []TriggerResult {
	var results []TriggerResult
	// 风险预警直接在快照层面判断
	for i := range snapshots {
		sn := &snapshots[i]
		// 诊断状态异常（拒审）
		if sn.DiagnosisStatus != 0 {
			reason := ""
			switch sn.DiagnosisStatus {
			case 1:
				reason = fmt.Sprintf("广告 %s 被拒审", sn.AdName)
			default:
				reason = fmt.Sprintf("广告 %s 状态异常(诊断状态=%d)", sn.AdName, sn.DiagnosisStatus)
			}
			results = append(results, TriggerResult{
				AccountID:  sn.AccountID,
				TargetID:   sn.AdID,
				TargetType: "ad",
				Snapshot:   sn,
				Reason:     reason,
				Matched:    true,
			})
		}
	}
	return results
}

// ---- 辅助函数 ----

// compareFloat 比较浮点数
func (e *RuleEngine) compareFloat(actual, expected float64, operator string) bool {
	switch operator {
	case "gt":
		return actual > expected
	case "gte":
		return actual >= expected
	case "lt":
		return actual < expected
	case "lte":
		return actual <= expected
	case "eq":
		return actual == expected
	default:
		return false
	}
}

// checkCooldown 检查规则是否在冷却期
func (e *RuleEngine) checkCooldown(rule *model.HostingRule) bool {
	if rule.CooldownMinutes <= 0 {
		return false
	}

	var lastExec model.HostingExecution
	err := database.GetDB().
		Where("rule_id = ?", rule.ID).
		Order("created_at DESC").
		First(&lastExec).Error

	if err != nil {
		return false // 没有执行记录，不需要冷却
	}

	coolDuration := time.Duration(rule.CooldownMinutes) * time.Minute
	return time.Since(lastExec.CreatedAt) < coolDuration
}

// checkDailyLimit 检查今日执行次数是否达到上限
func (e *RuleEngine) checkDailyLimit(rule *model.HostingRule) bool {
	if rule.MaxExecutionsPerDay <= 0 {
		return false
	}
	return rule.TodayExecCount >= rule.MaxExecutionsPerDay
}
