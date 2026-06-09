package handler

import (
	"strconv"
	"time"

	"marketing-platform/internal/middleware"
	"marketing-platform/internal/service"

	"github.com/gin-gonic/gin"
)

// ReportHandler 报表处理器
type ReportHandler struct {
	reportService *service.ReportService
}

// NewReportHandler 创建报表处理器
func NewReportHandler() *ReportHandler {
	return &ReportHandler{
		reportService: service.NewReportService(),
	}
}

// GetDailyReports 获取日报表
// @Summary 获取日报表
// @Tags report
// @Param account_id query string true "账号ID"
// @Param start_date query string true "开始日期 格式: 2006-01-02"
// @Param end_date query string true "结束日期 格式: 2006-01-02"
// @Param page query int false "页码 默认1"
// @Param page_size query int false "每页大小 默认20"
// @Router /api/v1/reports/daily [get]
func (h *ReportHandler) GetDailyReports(c *gin.Context) {
	accountID := c.Query("account_id")
	if accountID == "" {
		middleware.BadRequest(c, "account_id is required")
		return
	}

	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")
	if startDateStr == "" || endDateStr == "" {
		middleware.BadRequest(c, "start_date and end_date are required")
		return
	}

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		middleware.BadRequest(c, "Invalid start_date format, use 2006-01-02")
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		middleware.BadRequest(c, "Invalid end_date format, use 2006-01-02")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	reports, total, err := h.reportService.GetDailyReports(accountID, startDate, endDate, page, pageSize)
	if err != nil {
		middleware.InternalError(c, "Failed to get reports")
		return
	}

	middleware.Success(c, gin.H{
		"list":      reports,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetHourlyReports 获取小时报表
func (h *ReportHandler) GetHourlyReports(c *gin.Context) {
	accountID := c.Query("account_id")
	if accountID == "" {
		middleware.BadRequest(c, "account_id is required")
		return
	}

	dateStr := c.Query("date")
	if dateStr == "" {
		middleware.BadRequest(c, "date is required")
		return
	}

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		middleware.BadRequest(c, "Invalid date format, use 2006-01-02")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	reports, total, err := h.reportService.GetHourlyReports(accountID, date, page, pageSize)
	if err != nil {
		middleware.InternalError(c, "Failed to get reports")
		return
	}

	middleware.Success(c, gin.H{
		"list":      reports,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetTargetReports 获取定向报表
func (h *ReportHandler) GetTargetReports(c *gin.Context) {
	accountID := c.Query("account_id")
	if accountID == "" {
		middleware.BadRequest(c, "account_id is required")
		return
	}

	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")
	if startDateStr == "" || endDateStr == "" {
		middleware.BadRequest(c, "start_date and end_date are required")
		return
	}

	startDate, _ := time.Parse("2006-01-02", startDateStr)
	endDate, _ := time.Parse("2006-01-02", endDateStr)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	reports, total, err := h.reportService.GetTargetReports(accountID, startDate, endDate, page, pageSize)
	if err != nil {
		middleware.InternalError(c, "Failed to get reports")
		return
	}

	middleware.Success(c, gin.H{
		"list":      reports,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetDailyTrend 获取日报趋势
func (h *ReportHandler) GetDailyTrend(c *gin.Context) {
	accountID := c.Query("account_id")
	if accountID == "" {
		middleware.BadRequest(c, "account_id is required")
		return
	}

	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")
	if startDateStr == "" || endDateStr == "" {
		middleware.BadRequest(c, "start_date and end_date are required")
		return
	}

	startDate, _ := time.Parse("2006-01-02", startDateStr)
	endDate, _ := time.Parse("2006-01-02", endDateStr)

	trend, err := h.reportService.GetDailyTrend(accountID, startDate, endDate)
	if err != nil {
		middleware.InternalError(c, "Failed to get trend data")
		return
	}

	middleware.Success(c, gin.H{
		"list": trend,
	})
}
