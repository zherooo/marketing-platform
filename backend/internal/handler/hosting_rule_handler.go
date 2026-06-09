package handler

import (
	"strconv"

	"marketing-platform/internal/middleware"
	"marketing-platform/internal/model"
	"marketing-platform/internal/service"

	"github.com/gin-gonic/gin"
)

// HostingRuleHandler 托管规则处理器
type HostingRuleHandler struct {
	ruleService *service.HostingRuleService
	execService *service.HostingExecutorService
}

// NewHostingRuleHandler 创建托管规则处理器
func NewHostingRuleHandler() *HostingRuleHandler {
	return &HostingRuleHandler{
		ruleService: service.NewHostingRuleService(),
		execService: service.NewHostingExecutorService(),
	}
}

// CreateRule 创建规则
// POST /api/v1/hosting/rules
func (h *HostingRuleHandler) CreateRule(c *gin.Context) {
	var rule model.HostingRule
	if err := c.ShouldBindJSON(&rule); err != nil {
		middleware.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	if rule.RuleName == "" {
		middleware.BadRequest(c, "规则名称不能为空")
		return
	}
	if rule.Category == "" {
		middleware.BadRequest(c, "规则分类不能为空")
		return
	}
	if rule.TriggerCondition == nil {
		middleware.BadRequest(c, "触发条件不能为空")
		return
	}
	if rule.ExecutionAction == nil {
		middleware.BadRequest(c, "执行动作不能为空")
		return
	}

	// 设置默认值
	if rule.Priority == 0 {
		rule.Priority = 5
	}
	if rule.CooldownMinutes == 0 {
		rule.CooldownMinutes = 60
	}
	if rule.MaxExecutionsPerDay == 0 {
		rule.MaxExecutionsPerDay = 10
	}
	rule.Status = 1

	if err := h.ruleService.CreateRule(&rule); err != nil {
		middleware.InternalError(c, "创建规则失败: "+err.Error())
		return
	}

	middleware.Success(c, rule)
}

// UpdateRule 更新规则
// PUT /api/v1/hosting/rules/:id
func (h *HostingRuleHandler) UpdateRule(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		middleware.BadRequest(c, "无效的规则ID")
		return
	}

	existing, err := h.ruleService.GetRuleByID(uint(id))
	if err != nil {
		middleware.NotFound(c, "规则不存在")
		return
	}

	var updates model.HostingRule
	if err := c.ShouldBindJSON(&updates); err != nil {
		middleware.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	// 仅更新允许修改的字段
	existing.RuleName = updates.RuleName
	existing.Category = updates.Category
	existing.Scene = updates.Scene
	existing.Description = updates.Description
	existing.Priority = updates.Priority
	existing.TriggerCondition = updates.TriggerCondition
	existing.ExecutionAction = updates.ExecutionAction
	existing.AccountIDs = updates.AccountIDs
	existing.CampaignIDs = updates.CampaignIDs
	existing.AdGroupIDs = updates.AdGroupIDs
	existing.AdIDs = updates.AdIDs
	existing.CooldownMinutes = updates.CooldownMinutes
	existing.MaxExecutionsPerDay = updates.MaxExecutionsPerDay

	if err := h.ruleService.UpdateRule(existing); err != nil {
		middleware.InternalError(c, "更新规则失败: "+err.Error())
		return
	}

	middleware.Success(c, existing)
}

// DeleteRule 删除规则
// DELETE /api/v1/hosting/rules/:id
func (h *HostingRuleHandler) DeleteRule(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		middleware.BadRequest(c, "无效的规则ID")
		return
	}

	if err := h.ruleService.DeleteRule(uint(id)); err != nil {
		middleware.InternalError(c, "删除规则失败: "+err.Error())
		return
	}

	middleware.Success(c, nil)
}

// GetRule 获取规则详情
// GET /api/v1/hosting/rules/:id
func (h *HostingRuleHandler) GetRule(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		middleware.BadRequest(c, "无效的规则ID")
		return
	}

	rule, err := h.ruleService.GetRuleByID(uint(id))
	if err != nil {
		middleware.NotFound(c, "规则不存在")
		return
	}

	middleware.Success(c, rule)
}

// ListRules 获取规则列表
// GET /api/v1/hosting/rules
func (h *HostingRuleHandler) ListRules(c *gin.Context) {
	category := c.Query("category")
	status := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	rules, total, err := h.ruleService.ListRules(category, status, page, pageSize)
	if err != nil {
		middleware.InternalError(c, "获取规则列表失败: "+err.Error())
		return
	}

	middleware.Success(c, gin.H{
		"list":      rules,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// ToggleRuleStatus 切换规则启用/禁用
// POST /api/v1/hosting/rules/:id/toggle
func (h *HostingRuleHandler) ToggleRuleStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		middleware.BadRequest(c, "无效的规则ID")
		return
	}

	rule, err := h.ruleService.ToggleRuleStatus(uint(id))
	if err != nil {
		middleware.InternalError(c, "切换规则状态失败: "+err.Error())
		return
	}

	middleware.Success(c, rule)
}

// TestRule 测试规则（手动评估不执行）
// POST /api/v1/hosting/rules/:id/test
func (h *HostingRuleHandler) TestRule(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		middleware.BadRequest(c, "无效的规则ID")
		return
	}

	results, err := h.execService.EvaluateSingleRule(uint(id))
	if err != nil {
		middleware.InternalError(c, "规则测试失败: "+err.Error())
		return
	}

	middleware.Success(c, gin.H{
		"matched_count": len(results),
		"results":       results,
	})
}
