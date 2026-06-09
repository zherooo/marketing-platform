package handler

import (
	"strconv"

	"marketing-platform/internal/middleware"
	"marketing-platform/internal/service"

	"github.com/gin-gonic/gin"
)

// HostingExecutionHandler 托管执行记录处理器
type HostingExecutionHandler struct {
	execService *service.HostingExecutorService
}

// NewHostingExecutionHandler 创建托管执行记录处理器
func NewHostingExecutionHandler() *HostingExecutionHandler {
	return &HostingExecutionHandler{
		execService: service.NewHostingExecutorService(),
	}
}

// ListExecutions 获取执行记录列表
// GET /api/v1/hosting/executions
func (h *HostingExecutionHandler) ListExecutions(c *gin.Context) {
	accountID := c.Query("account_id")
	statusStr := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	status, _ := strconv.Atoi(statusStr)

	executions, total, err := h.execService.ListExecutions(accountID, status, page, pageSize)
	if err != nil {
		middleware.InternalError(c, "获取执行记录失败: "+err.Error())
		return
	}

	middleware.Success(c, gin.H{
		"list":      executions,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetExecution 获取执行记录详情
// GET /api/v1/hosting/executions/:id
func (h *HostingExecutionHandler) GetExecution(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		middleware.BadRequest(c, "无效的执行记录ID")
		return
	}

	execution, err := h.execService.GetExecutionByID(uint(id))
	if err != nil {
		middleware.NotFound(c, "执行记录不存在")
		return
	}

	middleware.Success(c, execution)
}

// RollbackExecution 回滚执行
// POST /api/v1/hosting/executions/:id/rollback
func (h *HostingExecutionHandler) RollbackExecution(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		middleware.BadRequest(c, "无效的执行记录ID")
		return
	}

	if err := h.execService.RollbackExecution(uint(id)); err != nil {
		middleware.InternalError(c, "回滚失败: "+err.Error())
		return
	}

	middleware.Success(c, gin.H{"message": "回滚成功"})
}

// GetDashboard 获取看板数据
// GET /api/v1/hosting/dashboard
func (h *HostingExecutionHandler) GetDashboard(c *gin.Context) {
	stats, err := h.execService.GetDashboardStats()
	if err != nil {
		middleware.InternalError(c, "获取看板数据失败: "+err.Error())
		return
	}

	trend, err := h.execService.GetExecutionTrend()
	if err != nil {
		middleware.InternalError(c, "获取趋势数据失败: "+err.Error())
		return
	}

	middleware.Success(c, gin.H{
		"stats": stats,
		"trend": trend,
	})
}

// TriggerEvaluate 手动触发评估
// POST /api/v1/hosting/trigger/evaluate
func (h *HostingExecutionHandler) TriggerEvaluate(c *gin.Context) {
	if err := h.execService.ExecuteAllActiveRules(); err != nil {
		middleware.InternalError(c, "触发评估失败: "+err.Error())
		return
	}

	middleware.Success(c, gin.H{"message": "评估触发成功"})
}

// TriggerCollect 手动触发性能快照采集
// POST /api/v1/hosting/trigger/collect
func (h *HostingExecutionHandler) TriggerCollect(c *gin.Context) {
	if err := h.execService.CollectPerformanceSnapshots(); err != nil {
		middleware.InternalError(c, "采集快照失败: "+err.Error())
		return
	}

	middleware.Success(c, gin.H{"message": "快照采集成功"})
}
