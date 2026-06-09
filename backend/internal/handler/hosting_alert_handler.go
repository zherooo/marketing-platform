package handler

import (
	"strconv"

	"marketing-platform/internal/middleware"
	"marketing-platform/internal/service"

	"github.com/gin-gonic/gin"
)

// HostingAlertHandler 托管告警处理器
type HostingAlertHandler struct {
	alertService *service.HostingAlertService
}

// NewHostingAlertHandler 创建托管告警处理器
func NewHostingAlertHandler() *HostingAlertHandler {
	return &HostingAlertHandler{
		alertService: service.NewHostingAlertService(),
	}
}

// ListAlerts 获取告警列表
// GET /api/v1/hosting/alerts
func (h *HostingAlertHandler) ListAlerts(c *gin.Context) {
	accountID := c.Query("account_id")
	alertType := c.Query("alert_type")
	severityStr := c.Query("severity")
	statusStr := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	severity, _ := strconv.Atoi(severityStr)
	status, _ := strconv.Atoi(statusStr)

	alerts, total, err := h.alertService.ListAlerts(accountID, alertType, severity, status, page, pageSize)
	if err != nil {
		middleware.InternalError(c, "获取告警列表失败: "+err.Error())
		return
	}

	middleware.Success(c, gin.H{
		"list":      alerts,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetAlert 获取告警详情
// GET /api/v1/hosting/alerts/:id
func (h *HostingAlertHandler) GetAlert(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		middleware.BadRequest(c, "无效的告警ID")
		return
	}

	alert, err := h.alertService.GetAlertByID(uint(id))
	if err != nil {
		middleware.NotFound(c, "告警不存在")
		return
	}

	middleware.Success(c, alert)
}

// MarkAlertRead 标记告警为已读
// POST /api/v1/hosting/alerts/:id/read
func (h *HostingAlertHandler) MarkAlertRead(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		middleware.BadRequest(c, "无效的告警ID")
		return
	}

	if err := h.alertService.MarkAsRead(uint(id)); err != nil {
		middleware.InternalError(c, "标记已读失败: "+err.Error())
		return
	}

	middleware.Success(c, nil)
}

// HandleAlert 处理告警
// POST /api/v1/hosting/alerts/:id/handle
func (h *HostingAlertHandler) HandleAlert(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		middleware.BadRequest(c, "无效的告警ID")
		return
	}

	var req struct {
		Handler string `json:"handler"`
		Result  string `json:"result"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	if err := h.alertService.HandleAlert(uint(id), req.Handler, req.Result); err != nil {
		middleware.InternalError(c, "处理告警失败: "+err.Error())
		return
	}

	middleware.Success(c, gin.H{"message": "处理成功"})
}

// IgnoreAlert 忽略告警
// POST /api/v1/hosting/alerts/:id/ignore
func (h *HostingAlertHandler) IgnoreAlert(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		middleware.BadRequest(c, "无效的告警ID")
		return
	}

	if err := h.alertService.IgnoreAlert(uint(id)); err != nil {
		middleware.InternalError(c, "忽略告警失败: "+err.Error())
		return
	}

	middleware.Success(c, gin.H{"message": "已忽略"})
}
