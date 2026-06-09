package engine

import (
	"encoding/json"
	"fmt"
	"time"

	"marketing-platform/internal/database"
	"marketing-platform/internal/logger"
	"marketing-platform/internal/model"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// ActionExecutor 动作执行器 —— 执行规则的 THEN 动作
type ActionExecutor struct {
	db *gorm.DB
}

// NewActionExecutor 创建动作执行器
func NewActionExecutor() *ActionExecutor {
	return &ActionExecutor{
		db: database.GetDB(),
	}
}

// Execute 执行触发结果对应的动作
func (e *ActionExecutor) Execute(result TriggerResult) (*model.HostingExecution, error) {
	execution := &model.HostingExecution{
		RuleID:     result.Rule.ID,
		RuleName:   result.Rule.RuleName,
		AccountID:  result.AccountID,
		TargetID:   result.TargetID,
		TargetType: result.TargetType,
		ActionType: result.Rule.ExecutionAction.Type,
		Status:     model.ExecStatusPending,
	}

	// 保存触发快照
	snapshotJSON, _ := json.Marshal(result.Snapshot)
	execution.TriggerSnapshot = string(snapshotJSON)

	action := result.Rule.ExecutionAction
	actionParamsJSON, _ := json.Marshal(action)
	execution.ActionParams = string(actionParamsJSON)

	// 根据动作类型执行
	var err error
	switch action.Type {
	case "pause_ad":
		err = e.pauseAd(execution, result, action)
	case "adjust_bid":
		err = e.adjustBid(execution, result, action)
	case "raise_budget":
		err = e.raiseBudget(execution, result, action)
	case "notify":
		err = e.sendNotification(execution, result, action)
	case "resume_ad":
		err = e.resumeAd(execution, result, action)
	case "quick_start":
		err = e.quickStart(execution, result, action)
	default:
		err = fmt.Errorf("未知的动作类型: %s", action.Type)
	}

	now := time.Now()
	execution.ExecutedAt = &now

	if err != nil {
		execution.Status = model.ExecStatusFailed
		execution.ErrorMsg = err.Error()
		logger.Logger.Error("执行动作失败",
			zap.Uint("rule_id", result.Rule.ID),
			zap.String("action", action.Type),
			zap.String("target", result.TargetID),
			zap.Error(err))
	} else {
		execution.Status = model.ExecStatusSuccess
		logger.Logger.Info("执行动作成功",
			zap.Uint("rule_id", result.Rule.ID),
			zap.String("action", action.Type),
			zap.String("target", result.TargetID))
	}

	// 保存执行记录
	if saveErr := e.db.Create(execution).Error; saveErr != nil {
		logger.Logger.Error("保存执行记录失败", zap.Error(saveErr))
		return execution, saveErr
	}

	// 更新规则统计
	e.db.Model(&model.HostingRule{}).Where("id = ?", result.Rule.ID).
		Updates(map[string]interface{}{
			"today_exec_count": gorm.Expr("today_exec_count + 1"),
			"total_exec_count": gorm.Expr("total_exec_count + 1"),
		})

	return execution, nil
}

// pauseAd 暂停广告
func (e *ActionExecutor) pauseAd(execution *model.HostingExecution, result TriggerResult, action *model.ExecutionAction) error {
	beforeJSON, _ := json.Marshal(map[string]interface{}{
		"target_id": result.TargetID,
		"action":    "pause",
	})

	execution.BeforeValue = string(beforeJSON)

	// TODO: 调用腾讯广告 API 暂停广告
	// apiClient := NewAPIClient(result.AccountID)
	// resp, err := apiClient.UpdateAdStatus(result.TargetID, "PAUSED")
	// execution.APIRawResp = resp

	logger.Logger.Info("[模拟] 暂停广告",
		zap.String("account_id", result.AccountID),
		zap.String("ad_id", result.TargetID),
		zap.String("reason", result.Reason))

	execution.APIRawResp = fmt.Sprintf(`{"code":0,"message":"[模拟成功] 广告 %s 已暂停"}`, result.TargetID)
	return nil
}

// adjustBid 调整出价
func (e *ActionExecutor) adjustBid(execution *model.HostingExecution, result TriggerResult, action *model.ExecutionAction) error {
	originalBid := result.Snapshot.BidAmount
	newBid := originalBid
	if action.BidAdjustRatio > 0 {
		newBid = originalBid * action.BidAdjustRatio
	}

	beforeJSON, _ := json.Marshal(map[string]interface{}{
		"target_id":    result.TargetID,
		"original_bid": originalBid,
	})
	afterJSON, _ := json.Marshal(map[string]interface{}{
		"new_bid":    newBid,
		"adjust_pct": action.BidAdjustRatio,
	})

	execution.BeforeValue = string(beforeJSON)
	execution.AfterValue = string(afterJSON)

	// TODO: 调用腾讯广告 API 调整出价
	logger.Logger.Info("[模拟] 调整出价",
		zap.String("ad_id", result.TargetID),
		zap.Float64("original_bid", originalBid),
		zap.Float64("new_bid", newBid))

	execution.APIRawResp = fmt.Sprintf(`{"code":0,"message":"[模拟成功] 出价已从 %.2f 调整为 %.2f"}`, originalBid, newBid)
	return nil
}

// raiseBudget 提升日限额
func (e *ActionExecutor) raiseBudget(execution *model.HostingExecution, result TriggerResult, action *model.ExecutionAction) error {
	originalBudget := result.Snapshot.DailyBudget
	newBudget := originalBudget
	if action.BudgetRaiseRatio > 0 {
		newBudget = originalBudget * action.BudgetRaiseRatio
	} else if action.BudgetRaiseAmount > 0 {
		newBudget = originalBudget + action.BudgetRaiseAmount
	}

	beforeJSON, _ := json.Marshal(map[string]interface{}{
		"target_id":       result.TargetID,
		"original_budget": originalBudget,
	})
	afterJSON, _ := json.Marshal(map[string]interface{}{
		"new_budget": newBudget,
	})

	execution.BeforeValue = string(beforeJSON)
	execution.AfterValue = string(afterJSON)

	// TODO: 调用腾讯广告 API 提升日限额
	logger.Logger.Info("[模拟] 提升日限额",
		zap.String("target_id", result.TargetID),
		zap.Float64("original_budget", originalBudget),
		zap.Float64("new_budget", newBudget))

	execution.APIRawResp = fmt.Sprintf(`{"code":0,"message":"[模拟成功] 日限额已从 %.2f 提升至 %.2f"}`, originalBudget, newBudget)
	return nil
}

// sendNotification 发送通知
func (e *ActionExecutor) sendNotification(execution *model.HostingExecution, result TriggerResult, action *model.ExecutionAction) error {
	alert := &model.HostingAlert{
		RuleID:        result.Rule.ID,
		AccountID:     result.AccountID,
		AlertType:     result.Rule.Category,
		AlertTitle:    fmt.Sprintf("[智能托管] %s", result.Rule.RuleName),
		AlertContent:  result.Reason,
		Severity:      model.AlertSeverityMedium,
		Status:        model.AlertStatusUnread,
		NotifyChannel: "system",
	}

	if len(action.NotifyChannels) > 0 {
		alert.NotifyChannel = action.NotifyChannels[0]
	}
	if len(action.NotifyUsers) > 0 {
		alert.NotifyUser = action.NotifyUsers[0]
	}
	if action.Message != "" {
		alert.AlertContent = action.Message + " - " + result.Reason
	}

	// 危险告警提升级别
	if result.Rule.Category == "risk_alert" {
		alert.Severity = model.AlertSeverityHigh
	}

	if err := e.db.Create(alert).Error; err != nil {
		logger.Logger.Error("保存告警记录失败", zap.Error(err))
		return err
	}

	execution.APIRawResp = fmt.Sprintf(`{"code":0,"message":"通知已发送", "alert_id": %d}`, alert.ID)

	// TODO: 实际发送通知（邮件、短信、钉钉、飞书等）
	logger.Logger.Info("[模拟] 发送通知",
		zap.String("title", alert.AlertTitle),
		zap.String("content", alert.AlertContent))

	return nil
}

// resumeAd 恢复广告投放（定时启用）
func (e *ActionExecutor) resumeAd(execution *model.HostingExecution, result TriggerResult, action *model.ExecutionAction) error {
	beforeJSON, _ := json.Marshal(map[string]interface{}{
		"target_id": result.TargetID,
		"action":    "resume",
	})

	execution.BeforeValue = string(beforeJSON)

	// TODO: 调用腾讯广告 API 启用广告
	logger.Logger.Info("[模拟] 启用广告",
		zap.String("ad_id", result.TargetID))

	execution.APIRawResp = fmt.Sprintf(`{"code":0,"message":"[模拟成功] 广告 %s 已启用"}`, result.TargetID)
	return nil
}

// quickStart 一键起量
func (e *ActionExecutor) quickStart(execution *model.HostingExecution, result TriggerResult, action *model.ExecutionAction) error {
	beforeJSON, _ := json.Marshal(map[string]interface{}{
		"target_id": result.TargetID,
		"action":    "quick_start",
	})

	execution.BeforeValue = string(beforeJSON)

	// TODO: 调用腾讯广告 API 一键起量
	logger.Logger.Info("[模拟] 一键起量",
		zap.String("ad_id", result.TargetID),
		zap.String("reason", result.Reason))

	execution.APIRawResp = fmt.Sprintf(`{"code":0,"message":"[模拟成功] 广告 %s 一键起量已启动"}`, result.TargetID)
	return nil
}

// RollbackExecution 回滚执行
func (e *ActionExecutor) RollbackExecution(execID uint) error {
	var exec model.HostingExecution
	if err := e.db.First(&exec, execID).Error; err != nil {
		return fmt.Errorf("执行记录不存在: %w", err)
	}

	if exec.Status == model.ExecStatusRolled {
		return fmt.Errorf("执行记录已回滚")
	}

	// 根据 beforeValue 恢复
	var beforeData map[string]interface{}
	if err := json.Unmarshal([]byte(exec.BeforeValue), &beforeData); err != nil {
		return fmt.Errorf("解析回滚数据失败: %w", err)
	}

	logger.Logger.Info("回滚执行动作",
		zap.Uint("exec_id", execID),
		zap.String("action", exec.ActionType),
		zap.Any("before_data", beforeData))

	// 更新状态为已回滚
	now := time.Now()
	e.db.Model(&exec).Updates(map[string]interface{}{
		"status":     model.ExecStatusRolled,
		"rollback_at": now,
	})

	return nil
}
