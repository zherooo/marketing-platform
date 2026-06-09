package crawler

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"gorm.io/gorm"

	"marketing-platform/internal/model"
)

// CrawlerInterface 抓取器接口
type CrawlerInterface interface {
	Crawl(ctx context.Context, task model.CrawlTaskItem) error
}

// BaseCrawler 抓取器基类
type BaseCrawler struct {
	db       *gorm.DB
	apiClient *APIClient
	batchSize int
}

// NewBaseCrawler 创建基础抓取器
func NewBaseCrawler(db *gorm.DB, apiClient *APIClient) *BaseCrawler {
	return &BaseCrawler{
		db:        db,
		apiClient: apiClient,
		batchSize: 100,
	}
}

// HourlyReportCrawler 小时报表抓取器
type HourlyReportCrawler struct {
	*BaseCrawler
}

// NewHourlyReportCrawler 创建小时报表抓取器
func NewHourlyReportCrawler(db *gorm.DB, apiClient *APIClient) *HourlyReportCrawler {
	return &HourlyReportCrawler{
		BaseCrawler: NewBaseCrawler(db, apiClient),
	}
}

// HourlyReportItem 小时报表数据项
type HourlyReportItem struct {
	Date         string  `json:"date"`
	Hour         int     `json:"hour"`
	CampaignID   string  `json:"campaign_id"`
	CampaignName string  `json:"campaign_name"`
	AdgroupID   string  `json:"adgroup_id"`
	AdgroupName string  `json:"adgroup_name"`
	AdID        string  `json:"ad_id"`
	AdName      string  `json:"ad_name"`
	ViewCount   int64   `json:"view_count"`
	ClickCount  int64   `json:"click_count"`
	ConvertCount int64  `json:"convert_count"`
	Spend       float64 `json:"spend"`
}

// HourlyReportList 报表列表响应
type HourlyReportList struct {
	List     []HourlyReportItem `json:"list"`
	PageInfo PageInfo           `json:"page_info"`
}

// PageInfo 分页信息
type PageInfo struct {
	Page       int `json:"page"`
	PageSize   int `json:"page_size"`
	TotalCount int `json:"total_count"`
	TotalPage  int `json:"total_page"`
}

// CrawlHourlyReport 抓取小时报表
func (c *HourlyReportCrawler) Crawl(ctx context.Context, task model.CrawlTaskItem) error {
	log.Printf("[HourlyReportCrawler] 开始抓取账号 %s, 日期范围: %s ~ %s",
		task.AccountID, task.StartDate.Format("2006-01-02"), task.EndDate.Format("2006-01-02"))

	date := task.StartDate.Format("2006-01-02")
	page := 1
	pageSize := 500
	successCount := int64(0)
	failCount := int64(0)
	
	var mu sync.Mutex
	var wg sync.WaitGroup

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		data, err := c.apiClient.GetHourlyReport(ctx, task.AccountID, date, page, pageSize)
		if err != nil {
			log.Printf("[HourlyReportCrawler] 获取报表失败: %v", err)
			atomic.AddInt64(&failCount, int64(pageSize))
			break
		}

		var reportList HourlyReportList
		if err := json.Unmarshal(data, &reportList); err != nil {
			log.Printf("[HourlyReportCrawler] 解析报表失败: %v", err)
			atomic.AddInt64(&failCount, int64(pageSize))
			break
		}

		if len(reportList.List) == 0 {
			break
		}

		// 批量处理
		batchSize := 100
		for i := 0; i < len(reportList.List); i += batchSize {
			end := i + batchSize
			if end > len(reportList.List) {
				end = len(reportList.List)
			}
			batch := reportList.List[i:end]

			wg.Add(1)
			go func(items []HourlyReportItem) {
				defer wg.Done()

				reports := make([]model.HourlyReport, 0, len(items))
				for _, item := range items {
					reportDate, _ := time.Parse("2006-01-02", item.Date)
					datetime := time.Date(reportDate.Year(), reportDate.Month(), reportDate.Day(), item.Hour, 0, 0, 0, reportDate.Location())
					report := model.HourlyReport{
						AccountID:     task.AccountID,
						Level:         "REPORT_LEVEL_AD",
						Date:          reportDate,
						Hour:          item.Hour,
						Datetime:      datetime,
						CampaignID:    item.CampaignID,
						CampaignName:  item.CampaignName,
						AdgroupID:    item.AdgroupID,
						AdgroupName:  item.AdgroupName,
						AdID:         item.AdID,
						AdName:       item.AdName,
						ViewCount:    item.ViewCount,
						ClickCount:   item.ClickCount,
						ConvertCount: item.ConvertCount,
						Spend:        item.Spend,
					}
					reports = append(reports, report)
				}

				// 批量保存
				if err := c.saveBatch(reports); err != nil {
					log.Printf("[HourlyReportCrawler] 保存报表失败: %v", err)
					atomic.AddInt64(&failCount, int64(len(items)))
					return
				}

				mu.Lock()
				successCount += int64(len(items))
				mu.Unlock()
			}(batch)
		}

		// 检查是否还有下一页
		if page >= reportList.PageInfo.TotalPage {
			break
		}
		page++
	}

	wg.Wait()

	// 更新任务进度
	log.Printf("[HourlyReportCrawler] 账号 %s 抓取完成: 成功 %d, 失败 %d",
		task.AccountID, successCount, failCount)
	return nil
}

// saveBatch 批量保存报表数据
func (c *HourlyReportCrawler) saveBatch(reports []model.HourlyReport) error {
	if len(reports) == 0 {
		return nil
	}

	// 使用批量插入
	return c.db.CreateInBatches(reports, 100).Error
}

// DailyReportCrawler 日报表抓取器
type DailyReportCrawler struct {
	*BaseCrawler
}

// NewDailyReportCrawler 创建日报表抓取器
func NewDailyReportCrawler(db *gorm.DB, apiClient *APIClient) *DailyReportCrawler {
	return &DailyReportCrawler{
		BaseCrawler: NewBaseCrawler(db, apiClient),
	}
}

// CrawlDailyReport 抓取日报表
func (c *DailyReportCrawler) Crawl(ctx context.Context, task model.CrawlTaskItem) error {
	log.Printf("[DailyReportCrawler] 开始抓取账号 %s, 日期范围: %s ~ %s",
		task.AccountID, task.StartDate.Format("2006-01-02"), task.EndDate.Format("2006-01-02"))

	startDate := task.StartDate.Format("2006-01-02")
	endDate := task.EndDate.Format("2006-01-02")
	page := 1
	pageSize := 500
	successCount := int64(0)
	failCount := int64(0)

	var mu sync.Mutex
	var wg sync.WaitGroup

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		data, err := c.apiClient.GetDailyReport(ctx, task.AccountID, startDate, endDate, page, pageSize)
		if err != nil {
			log.Printf("[DailyReportCrawler] 获取报表失败: %v", err)
			atomic.AddInt64(&failCount, int64(pageSize))
			break
		}

		var reportList HourlyReportList
		if err := json.Unmarshal(data, &reportList); err != nil {
			log.Printf("[DailyReportCrawler] 解析报表失败: %v", err)
			atomic.AddInt64(&failCount, int64(pageSize))
			break
		}

		if len(reportList.List) == 0 {
			break
		}

		// 批量处理
		batchSize := 100
		for i := 0; i < len(reportList.List); i += batchSize {
			end := i + batchSize
			if end > len(reportList.List) {
				end = len(reportList.List)
			}
			batch := reportList.List[i:end]

			wg.Add(1)
			go func(items []HourlyReportItem) {
				defer wg.Done()

				reports := make([]model.DailyReport, 0, len(items))
				for _, item := range items {
					reportDate, _ := time.Parse("2006-01-02", item.Date)
					report := model.DailyReport{
						AccountID:     task.AccountID,
						Level:         "REPORT_LEVEL_AD",
						Date:          reportDate,
						CampaignID:    item.CampaignID,
						CampaignName:  item.CampaignName,
						AdgroupID:    item.AdgroupID,
						AdgroupName:  item.AdgroupName,
						AdID:         item.AdID,
						AdName:       item.AdName,
						ViewCount:    item.ViewCount,
						ClickCount:   item.ClickCount,
						ConvertCount: item.ConvertCount,
						Spend:        item.Spend,
					}
					reports = append(reports, report)
				}

				if err := c.saveDailyBatch(reports); err != nil {
					log.Printf("[DailyReportCrawler] 保存报表失败: %v", err)
					atomic.AddInt64(&failCount, int64(len(items)))
					return
				}

				mu.Lock()
				successCount += int64(len(items))
				mu.Unlock()
			}(batch)
		}

		if page >= reportList.PageInfo.TotalPage {
			break
		}
		page++
	}

	wg.Wait()

	log.Printf("[DailyReportCrawler] 账号 %s 抓取完成: 成功 %d, 失败 %d",
		task.AccountID, successCount, failCount)
	return nil
}

// saveDailyBatch 批量保存日报表数据
func (c *DailyReportCrawler) saveDailyBatch(reports []model.DailyReport) error {
	if len(reports) == 0 {
		return nil
	}
	return c.db.CreateInBatches(reports, 100).Error
}
