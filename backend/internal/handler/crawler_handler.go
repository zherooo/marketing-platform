package handler

import (
	"log"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

"marketing-platform/internal/model"
"marketing-platform/internal/scheduler"
)

// CrawlerHandler 抓取控制器
type CrawlerHandler struct {
	scheduler *scheduler.Scheduler
}

// NewCrawlerHandler 创建抓取控制器
func NewCrawlerHandler(scheduler *scheduler.Scheduler) *CrawlerHandler {
	return &CrawlerHandler{
		scheduler: scheduler,
	}
}

// StartCrawlRequest 启动抓取请求
type StartCrawlRequest struct {
	AccountIDs []string `json:"account_ids" binding:"required"` // 账号ID列表
	DataTypes  []string `json:"data_types" binding:"required"`  // 数据类型列表
	StartDate  string   `json:"start_date"`                    // 开始日期 YYYY-MM-DD
	EndDate    string   `json:"end_date"`                      // 结束日期 YYYY-MM-DD
}

// StartCrawl 手动启动抓取
// POST /api/crawler/start
func (h *CrawlerHandler) StartCrawl(c *gin.Context) {
	var req StartCrawlRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
			"data":    nil,
		})
		return
	}

	// 解析日期
	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		c.JSON(400, gin.H{
			"code":    400,
			"message": "开始日期格式错误，应为 YYYY-MM-DD",
			"data":    nil,
		})
		return
	}

	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		c.JSON(400, gin.H{
			"code":    400,
			"message": "结束日期格式错误，应为 YYYY-MM-DD",
			"data":    nil,
		})
		return
	}

	crawlReq := &model.CrawlRequest{
		AccountIDs: req.AccountIDs,
		DataTypes:  req.DataTypes,
		StartDate:  startDate,
		EndDate:    endDate,
		Manual:     true,
	}

	stats, err := h.scheduler.TriggerManualCrawl(crawlReq)
	if err != nil {
		c.JSON(500, gin.H{
			"code":    500,
			"message": "启动抓取失败: " + err.Error(),
			"data":    nil,
		})
		return
	}

	c.JSON(200, gin.H{
		"code":    200,
		"message": "抓取任务已启动",
		"data":    stats,
	})
}

// GetStatistics 获取抓取统计
// GET /api/crawler/statistics
func (h *CrawlerHandler) GetStatistics(c *gin.Context) {
	stats, err := h.scheduler.GetStatistics()
	if err != nil {
		c.JSON(500, gin.H{
			"code":    500,
			"message": "获取统计失败: " + err.Error(),
			"data":    nil,
		})
		return
	}

	c.JSON(200, gin.H{
		"code":    200,
		"message": "success",
		"data":    stats,
	})
}

// GetRunningTasks 获取正在运行的任务
// GET /api/crawler/tasks/running
func (h *CrawlerHandler) GetRunningTasks(c *gin.Context) {
	tasks, err := h.scheduler.GetRunningTasks()
	if err != nil {
		c.JSON(500, gin.H{
			"code":    500,
			"message": "获取任务失败: " + err.Error(),
			"data":    nil,
		})
		return
	}

	c.JSON(200, gin.H{
		"code":    200,
		"message": "success",
		"data":    tasks,
	})
}

// ListTasks 查询任务列表
// GET /api/crawler/tasks
func (h *CrawlerHandler) ListTasks(c *gin.Context) {
	accountID := c.Query("account_id")
	statusStr := c.Query("status")
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "20")

	status := -1
	if statusStr != "" {
		var err error
		status, err = strconv.Atoi(statusStr)
		if err != nil {
			c.JSON(400, gin.H{
				"code":    400,
				"message": "status 参数错误",
				"data":    nil,
			})
			return
		}
	}

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	tasks, total, err := h.scheduler.ListTasks(accountID, status, page, pageSize)
	if err != nil {
		c.JSON(500, gin.H{
			"code":    500,
			"message": "获取任务列表失败: " + err.Error(),
			"data":    nil,
		})
		return
	}

	c.JSON(200, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"list":       tasks,
			"total":      total,
			"page":       page,
			"page_size":  pageSize,
			"total_page": (total + int64(pageSize) - 1) / int64(pageSize),
		},
	})
}

// CancelTask 取消任务
// POST /api/crawler/tasks/:task_id/cancel
func (h *CrawlerHandler) CancelTask(c *gin.Context) {
	taskID := c.Param("task_id")
	if taskID == "" {
		c.JSON(400, gin.H{
			"code":    400,
			"message": "任务ID不能为空",
			"data":    nil,
		})
		return
	}

	if err := h.scheduler.CancelTask(taskID); err != nil {
		c.JSON(500, gin.H{
			"code":    500,
			"message": "取消任务失败: " + err.Error(),
			"data":    nil,
		})
		return
	}

	c.JSON(200, gin.H{
		"code":    200,
		"message": "任务已取消",
		"data":    nil,
	})
}

// RetryTask 重试任务
// POST /api/crawler/tasks/:task_id/retry
func (h *CrawlerHandler) RetryTask(c *gin.Context) {
	taskID := c.Param("task_id")
	if taskID == "" {
		c.JSON(400, gin.H{
			"code":    400,
			"message": "任务ID不能为空",
			"data":    nil,
		})
		return
	}

	if err := h.scheduler.RetryTask(taskID); err != nil {
		c.JSON(500, gin.H{
			"code":    500,
			"message": "重试任务失败: " + err.Error(),
			"data":    nil,
		})
		return
	}

	c.JSON(200, gin.H{
		"code":    200,
		"message": "任务已重新加入队列",
		"data":    nil,
	})
}

// TriggerHourlyReport 手动触发小时报表抓取
// POST /api/crawler/trigger/hourly
func (h *CrawlerHandler) TriggerHourlyReport(c *gin.Context) {
	yesterday := time.Now().AddDate(0, 0, -1)
	req := &model.CrawlRequest{
		AccountIDs: []string{}, // 空数组表示所有账号
		DataTypes:  []string{model.DataTypeHourlyReport},
		StartDate:  yesterday,
		EndDate:    yesterday,
		Manual:     true,
	}

	stats, err := h.scheduler.TriggerManualCrawl(req)
	if err != nil {
		c.JSON(500, gin.H{
			"code":    500,
			"message": "触发失败: " + err.Error(),
			"data":    nil,
		})
		return
	}

	c.JSON(200, gin.H{
		"code":    200,
		"message": "小时报表抓取任务已启动",
		"data":    stats,
	})
}

// TriggerDailyReport 手动触发动日报表抓取
// POST /api/crawler/trigger/daily
func (h *CrawlerHandler) TriggerDailyReport(c *gin.Context) {
	yesterday := time.Now().AddDate(0, 0, -1)
	req := &model.CrawlRequest{
		AccountIDs: []string{},
		DataTypes:  []string{model.DataTypeDailyReport},
		StartDate:  yesterday,
		EndDate:    yesterday,
		Manual:     true,
	}

	stats, err := h.scheduler.TriggerManualCrawl(req)
	if err != nil {
		c.JSON(500, gin.H{
			"code":    500,
			"message": "触发失败: " + err.Error(),
			"data":    nil,
		})
		return
	}

	c.JSON(200, gin.H{
		"code":    200,
		"message": "日报表抓取任务已启动",
		"data":    stats,
	})
}

// TriggerAllStruct 手动触发所有广告结构抓取
// POST /api/crawler/trigger/struct
func (h *CrawlerHandler) TriggerAllStruct(c *gin.Context) {
	req := &model.CrawlRequest{
		AccountIDs: []string{},
		DataTypes: []string{
			model.DataTypeCampaign,
			model.DataTypeAdGroup,
			model.DataTypeAd,
			model.DataTypeCreative,
			model.DataTypeMaterial,
		},
		StartDate: time.Time{},
		EndDate:   time.Time{},
		Manual:    true,
	}

	stats, err := h.scheduler.TriggerManualCrawl(req)
	if err != nil {
		c.JSON(500, gin.H{
			"code":    500,
			"message": "触发失败: " + err.Error(),
			"data":    nil,
		})
		return
	}

	c.JSON(200, gin.H{
		"code":    200,
		"message": "广告结构抓取任务已启动",
		"data":    stats,
	})
}

// TriggerCampaign 手动触发广告系列抓取
// POST /api/crawler/trigger/campaign
func (h *CrawlerHandler) TriggerCampaign(c *gin.Context) {
	req := &model.CrawlRequest{
		AccountIDs: []string{},
		DataTypes: []string{
			model.DataTypeCampaign,
		},
		StartDate: time.Time{},
		EndDate:   time.Time{},
		Manual:    true,
	}

	stats, err := h.scheduler.TriggerManualCrawl(req)
	if err != nil {
		c.JSON(500, gin.H{
			"code":    500,
			"message": "触发失败: " + err.Error(),
			"data":    nil,
		})
		return
	}

	c.JSON(200, gin.H{
		"code":    200,
		"message": "广告系列抓取任务已启动",
		"data":    stats,
	})
}

// TriggerCampaignCascade 手动触发广告系列级联抓取
// POST /api/crawler/trigger/campaign/cascade
func (h *CrawlerHandler) TriggerCampaignCascade(c *gin.Context) {
	campaignID := c.Param("campaign_id")
	accountID := c.Query("account_id")

	if campaignID == "" || accountID == "" {
		c.JSON(400, gin.H{
			"code":    400,
			"message": "campaign_id 和 account_id 不能为空",
			"data":    nil,
		})
		return
	}

	// 异步执行级联抓取
	go func() {
		if err := h.scheduler.CrawlCampaignCascade(accountID, campaignID); err != nil {
			log.Printf("级联抓取广告系列失败: %v", err)
		}
	}()

	c.JSON(200, gin.H{
		"code":    200,
		"message": "广告系列级联抓取已触发，正在抓取该系列下的广告组、广告、创意、素材",
		"data":    nil,
	})
}

// TriggerAdGroupCascade 手动触发广告组级联抓取
// POST /api/crawler/trigger/adgroup/:adgroup_id/cascade
func (h *CrawlerHandler) TriggerAdGroupCascade(c *gin.Context) {
	adgroupID := c.Param("adgroup_id")
	accountID := c.Query("account_id")

	if adgroupID == "" || accountID == "" {
		c.JSON(400, gin.H{
			"code":    400,
			"message": "adgroup_id 和 account_id 不能为空",
			"data":    nil,
		})
		return
	}

	// 异步执行级联抓取
	go func() {
		if err := h.scheduler.CrawlAdGroupCascade(accountID, adgroupID); err != nil {
			log.Printf("级联抓取广告组失败: %v", err)
		}
	}()

	c.JSON(200, gin.H{
		"code":    200,
		"message": "广告组级联抓取已触发，正在抓取该组下的广告、创意、素材",
		"data":    nil,
	})
}

// TriggerAdCascade 手动触发广告级联抓取
// POST /api/crawler/trigger/ad/:ad_id/cascade
func (h *CrawlerHandler) TriggerAdCascade(c *gin.Context) {
	adID := c.Param("ad_id")
	accountID := c.Query("account_id")

	if adID == "" || accountID == "" {
		c.JSON(400, gin.H{
			"code":    400,
			"message": "ad_id 和 account_id 不能为空",
			"data":    nil,
		})
		return
	}

	// 异步执行级联抓取
	go func() {
		if err := h.scheduler.CrawlAdCascade(accountID, adID); err != nil {
			log.Printf("级联抓取广告失败: %v", err)
		}
	}()

	c.JSON(200, gin.H{
		"code":    200,
		"message": "广告级联抓取已触发，正在抓取该广告下的创意、素材",
		"data":    nil,
	})
}
