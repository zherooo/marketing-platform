package crawler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"gorm.io/gorm"

	"marketing-platform/internal/model"
)

// CampaignCrawler 广告系列抓取器
type CampaignCrawler struct {
	*BaseCrawler
}

// NewCampaignCrawler 创建广告系列抓取器
func NewCampaignCrawler(db *gorm.DB, apiClient *APIClient) *CampaignCrawler {
	return &CampaignCrawler{
		BaseCrawler: NewBaseCrawler(db, apiClient),
	}
}

// CampaignItem 广告系列数据项
type CampaignItem struct {
	CampaignID   string `json:"campaign_id"`
	CampaignName string `json:"campaign_name"`
	CampaignType string `json:"campaign_type"`
	Status       string `json:"status"`
	DailyBudget  int64  `json:"daily_budget"`
	CreatedTime  string `json:"created_time"`
}

// CampaignList 广告系列列表响应
type CampaignList struct {
	List     []CampaignItem `json:"list"`
	PageInfo PageInfo       `json:"page_info"`
}

// CrawlCampaign 抓取广告系列
func (c *CampaignCrawler) Crawl(ctx context.Context, task model.CrawlTaskItem) error {
	log.Printf("[CampaignCrawler] 开始抓取账号 %s 的广告系列", task.AccountID)

	page := 1
	pageSize := 500
	successCount := int64(0)
	failCount := int64(0)
	campaignIDs := make([]string, 0)

	var mu sync.Mutex
	var wg sync.WaitGroup

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		data, err := c.apiClient.GetCampaigns(ctx, task.AccountID, page, pageSize)
		if err != nil {
			log.Printf("[CampaignCrawler] 获取广告系列失败: %v", err)
			atomic.AddInt64(&failCount, int64(pageSize))
			break
		}

		var campaignList CampaignList
		if err := json.Unmarshal(data, &campaignList); err != nil {
			log.Printf("[CampaignCrawler] 解析广告系列失败: %v", err)
			atomic.AddInt64(&failCount, int64(pageSize))
			break
		}

		if len(campaignList.List) == 0 {
			break
		}

		// 批量处理
		batchSize := 100
		for i := 0; i < len(campaignList.List); i += batchSize {
			end := i + batchSize
			if end > len(campaignList.List) {
				end = len(campaignList.List)
			}
			batch := campaignList.List[i:end]

			wg.Add(1)
			go func(items []CampaignItem) {
				defer wg.Done()

				for _, item := range items {
					createdTime, _ := time.Parse("2006-01-02 15:04:05", item.CreatedTime)
					campaign := model.Campaign{
						CampaignID:   item.CampaignID,
						AccountID:    task.AccountID,
						CampaignName: item.CampaignName,
						CampaignType: c.parseCampaignType(item.CampaignType),
						Status:       c.parseStatus(item.Status),
						CreatedAt:    createdTime,
					}

					if err := c.saveCampaign(&campaign); err != nil {
						log.Printf("[CampaignCrawler] 保存广告系列失败: %v", err)
						atomic.AddInt64(&failCount, 1)
						continue
					}

					mu.Lock()
					campaignIDs = append(campaignIDs, item.CampaignID)
					mu.Unlock()
					atomic.AddInt64(&successCount, 1)
				}
			}(batch)
		}

		if page >= campaignList.PageInfo.TotalPage {
			break
		}
		page++
	}

	wg.Wait()

	// 删除不再存在的广告系列
	if err := c.deleteNonExistentCampaigns(task.AccountID, campaignIDs); err != nil {
		log.Printf("[CampaignCrawler] 删除失效广告系列失败: %v", err)
	}

	log.Printf("[CampaignCrawler] 账号 %s 抓取完成: 成功 %d, 失败 %d",
		task.AccountID, successCount, failCount)
	return nil
}

// parseCampaignType 解析推广计划类型
func (c *CampaignCrawler) parseCampaignType(campaignType string) int {
	switch campaignType {
	case "CAMPAIGN_TYPE_NORMAL":
		return 1
	case "CAMPAIGN_TYPE_WEBinar":
		return 2
	case "CAMPAIGN_TYPE_ECOMMERCE":
		return 3
	case "CAMPAIGN_TYPE_APP":
		return 4
	case "CAMPAIGN_TYPE_BRAND_CONTAINER":
		return 5
	default:
		return 0
	}
}

// parseStatus 解析状态
func (c *CampaignCrawler) parseStatus(status string) int {
	switch status {
	case "AD_STATUS_NORMAL":
		return 1
	case "AD_STATUS_SUSPEND":
		return 2
	case "AD_STATUS_DELETE":
		return 3
	default:
		return 0
	}
}

// saveCampaign 保存广告系列
func (c *CampaignCrawler) saveCampaign(campaign *model.Campaign) error {
	return c.db.Where("account_id = ? AND campaign_id = ?",
		campaign.AccountID, campaign.CampaignID).
		Assign(campaign).
		FirstOrCreate(campaign).Error
}

// deleteNonExistentCampaigns 删除不再存在的广告系列
func (c *CampaignCrawler) deleteNonExistentCampaigns(accountID string, existingIDs []string) error {
	if len(existingIDs) == 0 {
		return nil
	}

	return c.db.Where("account_id = ? AND campaign_id NOT IN ?", accountID, existingIDs).
		Delete(&model.Campaign{}).Error
}

// CrawlAdGroupsForCampaign 只抓取指定广告系列下的广告组，返回抓取到的广告组ID列表
func (c *CampaignCrawler) CrawlAdGroupsForCampaign(ctx context.Context, accountID, campaignID string) ([]string, error) {
	log.Printf("[CampaignCrawler] 开始抓取广告系列 %s 下的广告组", campaignID)
	adgroupIDs := make([]string, 0)
	var mu sync.Mutex
	var wg sync.WaitGroup

	page := 1
	pageSize := 500

	for {
		select {
		case <-ctx.Done():
			return adgroupIDs, ctx.Err()
		default:
		}

		data, err := c.apiClient.GetAdGroups(ctx, accountID, campaignID, page, pageSize)
		if err != nil {
			return adgroupIDs, fmt.Errorf("获取广告组失败: %w", err)
		}

		var adgroupList AdGroupList
		if err := json.Unmarshal(data, &adgroupList); err != nil {
			return adgroupIDs, fmt.Errorf("解析广告组失败: %w", err)
		}

		if len(adgroupList.List) == 0 {
			break
		}

		batchSize := 100
		for i := 0; i < len(adgroupList.List); i += batchSize {
			end := i + batchSize
			if end > len(adgroupList.List) {
				end = len(adgroupList.List)
			}
			batch := adgroupList.List[i:end]

			wg.Add(1)
			go func(items []AdGroupItem) {
				defer wg.Done()
				for _, item := range items {
					createdTime, _ := time.Parse("2006-01-02 15:04:05", item.CreatedTime)
					adgroup := model.AdGroup{
						GroupID:    item.AdGroupID,
						CampaignID: item.CampaignID,
						AccountID:  accountID,
						GroupName:  item.AdGroupName,
						Status:     c.parseStatus(item.Status),
						CreatedAt:  createdTime,
					}
					if err := c.saveAdGroup(&adgroup); err != nil {
						log.Printf("[CampaignCrawler] 保存广告组失败: %v", err)
						continue
					}
					mu.Lock()
					adgroupIDs = append(adgroupIDs, item.AdGroupID)
					mu.Unlock()
				}
			}(batch)
		}

		if page >= adgroupList.PageInfo.TotalPage {
			break
		}
		page++
	}

	wg.Wait()
	log.Printf("[CampaignCrawler] 广告系列 %s 抓取完成，共 %d 个广告组", campaignID, len(adgroupIDs))
	return adgroupIDs, nil
}

// saveAdGroup 保存广告组（复用保存逻辑）
func (c *CampaignCrawler) saveAdGroup(adgroup *model.AdGroup) error {
	return c.db.Where("account_id = ? AND group_id = ?",
		adgroup.AccountID, adgroup.GroupID).
		Assign(adgroup).
		FirstOrCreate(adgroup).Error
}

// AdGroupCrawler 广告组抓取器
type AdGroupCrawler struct {
	*BaseCrawler
	campaignCrawler *CampaignCrawler
}

// NewAdGroupCrawler 创建广告组抓取器
func NewAdGroupCrawler(db *gorm.DB, apiClient *APIClient) *AdGroupCrawler {
	return &AdGroupCrawler{
		BaseCrawler:     NewBaseCrawler(db, apiClient),
		campaignCrawler: NewCampaignCrawler(db, apiClient),
	}
}

// AdGroupItem 广告组数据项
type AdGroupItem struct {
	AdGroupID   string `json:"adgroup_id"`
	AdGroupName string `json:"adgroup_name"`
	CampaignID  string `json:"campaign_id"`
	Status      string `json:"status"`
	BidAmount   int64  `json:"bid_amount"`
	CreatedTime string `json:"created_time"`
}

// AdGroupList 广告组列表响应
type AdGroupList struct {
	List     []AdGroupItem `json:"list"`
	PageInfo PageInfo      `json:"page_info"`
}

// CrawlAdGroup 抓取广告组
func (c *AdGroupCrawler) Crawl(ctx context.Context, task model.CrawlTaskItem) error {
	log.Printf("[AdGroupCrawler] 开始抓取账号 %s 的广告组", task.AccountID)

	// 先获取所有广告系列
	campaigns, err := c.getCampaigns(task.AccountID)
	if err != nil {
		return fmt.Errorf("获取广告系列失败: %w", err)
	}

	if len(campaigns) == 0 {
		log.Printf("[AdGroupCrawler] 账号 %s 没有广告系列", task.AccountID)
		return nil
	}

	successCount := int64(0)
	failCount := int64(0)
	adgroupIDs := make([]string, 0)
	var mu sync.Mutex
	var wg sync.WaitGroup

	// 为每个广告系列创建 worker
	semaphore := make(chan struct{}, 5) // 最多5个并发

	for _, campaign := range campaigns {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		semaphore <- struct{}{}
		wg.Add(1)

		go func(campaignID string) {
			defer wg.Done()
			defer func() { <-semaphore }()

			page := 1
			pageSize := 500

			for {
				data, err := c.apiClient.GetAdGroups(ctx, task.AccountID, campaignID, page, pageSize)
				if err != nil {
					log.Printf("[AdGroupCrawler] 获取广告组失败 (campaign: %s): %v", campaignID, err)
					atomic.AddInt64(&failCount, int64(pageSize))
					break
				}

				var adgroupList AdGroupList
				if err := json.Unmarshal(data, &adgroupList); err != nil {
					log.Printf("[AdGroupCrawler] 解析广告组失败: %v", err)
					atomic.AddInt64(&failCount, int64(pageSize))
					break
				}

				if len(adgroupList.List) == 0 {
					break
				}

				for _, item := range adgroupList.List {
					createdTime, _ := time.Parse("2006-01-02 15:04:05", item.CreatedTime)
					adgroup := model.AdGroup{
						GroupID:     item.AdGroupID,
						CampaignID:  item.CampaignID,
						AccountID:   task.AccountID,
						GroupName:   item.AdGroupName,
						Status:      c.parseStatus(item.Status),
						CreatedAt:   createdTime,
					}

					if err := c.saveAdGroup(&adgroup); err != nil {
						log.Printf("[AdGroupCrawler] 保存广告组失败: %v", err)
						atomic.AddInt64(&failCount, 1)
						continue
					}

					mu.Lock()
					adgroupIDs = append(adgroupIDs, item.AdGroupID)
					mu.Unlock()
					atomic.AddInt64(&successCount, 1)
				}

				if page >= adgroupList.PageInfo.TotalPage {
					break
				}
				page++
			}
		}(campaign)
	}

	wg.Wait()

	// 删除不再存在的广告组
	if err := c.deleteNonExistentAdGroups(task.AccountID, adgroupIDs); err != nil {
		log.Printf("[AdGroupCrawler] 删除失效广告组失败: %v", err)
	}

	log.Printf("[AdGroupCrawler] 账号 %s 抓取完成: 成功 %d, 失败 %d",
		task.AccountID, successCount, failCount)
	return nil
}

// getCampaigns 获取账号下所有广告系列
func (c *AdGroupCrawler) getCampaigns(accountID string) ([]string, error) {
	var campaigns []model.Campaign
	if err := c.db.Where("account_id = ?", accountID).Pluck("campaign_id", &campaigns).Error; err != nil {
		return nil, err
	}

	ids := make([]string, len(campaigns))
	for i, c := range campaigns {
		ids[i] = c.CampaignID
	}
	return ids, nil
}

// parseStatus 解析状态
func (c *AdGroupCrawler) parseStatus(status string) int {
	switch status {
	case "AD_STATUS_NORMAL":
		return 1
	case "AD_STATUS_SUSPEND":
		return 2
	case "AD_STATUS_DELETE":
		return 3
	default:
		return 0
	}
}

// saveAdGroup 保存广告组
func (c *AdGroupCrawler) saveAdGroup(adgroup *model.AdGroup) error {
	return c.db.Where("account_id = ? AND group_id = ?",
		adgroup.AccountID, adgroup.GroupID).
		Assign(adgroup).
		FirstOrCreate(adgroup).Error
}

// deleteNonExistentAdGroups 删除不再存在的广告组
func (c *AdGroupCrawler) deleteNonExistentAdGroups(accountID string, existingIDs []string) error {
	if len(existingIDs) == 0 {
		return nil
	}

	return c.db.Where("account_id = ? AND group_id NOT IN ?", accountID, existingIDs).
		Delete(&model.AdGroup{}).Error
}
