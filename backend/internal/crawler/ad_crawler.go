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

// AdCrawler 广告抓取器
type AdCrawler struct {
	*BaseCrawler
}

// NewAdCrawler 创建广告抓取器
func NewAdCrawler(db *gorm.DB, apiClient *APIClient) *AdCrawler {
	return &AdCrawler{
		BaseCrawler: NewBaseCrawler(db, apiClient),
	}
}

// AdItem 广告数据项
type AdItem struct {
	AdID        string `json:"ad_id"`
	AdName      string `json:"ad_name"`
	AdGroupID   string `json:"adgroup_id"`
	CampaignID  string `json:"campaign_id"`
	Status      string `json:"status"`
	CreatedTime string `json:"created_time"`
}

// AdList 广告列表响应
type AdList struct {
	List     []AdItem `json:"list"`
	PageInfo PageInfo `json:"page_info"`
}

// CrawlAd 抓取广告
func (c *AdCrawler) Crawl(ctx context.Context, task model.CrawlTaskItem) error {
	log.Printf("[AdCrawler] 开始抓取账号 %s 的广告", task.AccountID)

	// 先获取所有广告组
	adgroups, err := c.getAdGroups(task.AccountID)
	if err != nil {
		return fmt.Errorf("获取广告组失败: %w", err)
	}

	if len(adgroups) == 0 {
		log.Printf("[AdCrawler] 账号 %s 没有广告组", task.AccountID)
		return nil
	}

	successCount := int64(0)
	failCount := int64(0)
	adIDs := make([]string, 0)
	var mu sync.Mutex
	var wg sync.WaitGroup

	// 为每个广告组创建 worker，最多5个并发
	semaphore := make(chan struct{}, 5)

	for _, adgroup := range adgroups {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		semaphore <- struct{}{}
		wg.Add(1)

		go func(groupID string) {
			defer wg.Done()
			defer func() { <-semaphore }()

			page := 1
			pageSize := 500

			for {
				data, err := c.apiClient.GetAds(ctx, task.AccountID, groupID, page, pageSize)
				if err != nil {
					log.Printf("[AdCrawler] 获取广告失败 (adgroup: %s): %v", groupID, err)
					atomic.AddInt64(&failCount, int64(pageSize))
					break
				}

				var adList AdList
				if err := json.Unmarshal(data, &adList); err != nil {
					log.Printf("[AdCrawler] 解析广告失败: %v", err)
					atomic.AddInt64(&failCount, int64(pageSize))
					break
				}

				if len(adList.List) == 0 {
					break
				}

				for _, item := range adList.List {
					createdTime, _ := time.Parse("2006-01-02 15:04:05", item.CreatedTime)
					ad := model.Ad{
						AdID:       item.AdID,
						GroupID:    item.AdGroupID,
						CampaignID: item.CampaignID,
						AccountID:  task.AccountID,
						AdName:     item.AdName,
						Status:     c.parseStatus(item.Status),
						CreatedAt:  createdTime,
					}

					if err := c.saveAd(&ad); err != nil {
						log.Printf("[AdCrawler] 保存广告失败: %v", err)
						atomic.AddInt64(&failCount, 1)
						continue
					}

					mu.Lock()
					adIDs = append(adIDs, item.AdID)
					mu.Unlock()
					atomic.AddInt64(&successCount, 1)
				}

				if page >= adList.PageInfo.TotalPage {
					break
				}
				page++
			}
		}(adgroup)
	}

	wg.Wait()

	// 删除不再存在的广告
	if err := c.deleteNonExistentAds(task.AccountID, adIDs); err != nil {
		log.Printf("[AdCrawler] 删除失效广告失败: %v", err)
	}

	log.Printf("[AdCrawler] 账号 %s 抓取完成: 成功 %d, 失败 %d",
		task.AccountID, successCount, failCount)
	return nil
}

// getAdGroups 获取账号下所有广告组
func (c *AdCrawler) getAdGroups(accountID string) ([]string, error) {
	var adgroups []model.AdGroup
	if err := c.db.Where("account_id = ?", accountID).Pluck("group_id", &adgroups).Error; err != nil {
		return nil, err
	}

	ids := make([]string, len(adgroups))
	for i, ag := range adgroups {
		ids[i] = ag.GroupID
	}
	return ids, nil
}

// parseStatus 解析状态
func (c *AdCrawler) parseStatus(status string) int {
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

// saveAd 保存广告
func (c *AdCrawler) saveAd(ad *model.Ad) error {
	return c.db.Where("account_id = ? AND ad_id = ?",
		ad.AccountID, ad.AdID).
		Assign(ad).
		FirstOrCreate(ad).Error
}

// deleteNonExistentAds 删除不再存在的广告
func (c *AdCrawler) deleteNonExistentAds(accountID string, existingIDs []string) error {
	if len(existingIDs) == 0 {
		return nil
	}

	return c.db.Where("account_id = ? AND ad_id NOT IN ?", accountID, existingIDs).
		Delete(&model.Ad{}).Error
}

// CrawlAdsForAdGroup 只抓取指定广告组下的广告，返回抓取到的广告ID列表
func (c *AdCrawler) CrawlAdsForAdGroup(ctx context.Context, accountID, adgroupID string) ([]string, error) {
	log.Printf("[AdCrawler] 开始抓取广告组 %s 下的广告", adgroupID)
	adIDs := make([]string, 0)
	var mu sync.Mutex
	var wg sync.WaitGroup

	page := 1
	pageSize := 500

	for {
		select {
		case <-ctx.Done():
			return adIDs, ctx.Err()
		default:
		}

		data, err := c.apiClient.GetAds(ctx, accountID, adgroupID, page, pageSize)
		if err != nil {
			return adIDs, fmt.Errorf("获取广告失败: %w", err)
		}

		var adList AdList
		if err := json.Unmarshal(data, &adList); err != nil {
			return adIDs, fmt.Errorf("解析广告失败: %w", err)
		}

		if len(adList.List) == 0 {
			break
		}

		batchSize := 100
		for i := 0; i < len(adList.List); i += batchSize {
			end := i + batchSize
			if end > len(adList.List) {
				end = len(adList.List)
			}
			batch := adList.List[i:end]

			wg.Add(1)
			go func(items []AdItem) {
				defer wg.Done()
				for _, item := range items {
					createdTime, _ := time.Parse("2006-01-02 15:04:05", item.CreatedTime)
					ad := model.Ad{
						AdID:       item.AdID,
						GroupID:    item.AdGroupID,
						CampaignID: item.CampaignID,
						AccountID:  accountID,
						AdName:     item.AdName,
						Status:     c.parseStatus(item.Status),
						CreatedAt:  createdTime,
					}
					if err := c.saveAd(&ad); err != nil {
						log.Printf("[AdCrawler] 保存广告失败: %v", err)
						continue
					}
					mu.Lock()
					adIDs = append(adIDs, item.AdID)
					mu.Unlock()
				}
			}(batch)
		}

		if page >= adList.PageInfo.TotalPage {
			break
		}
		page++
	}

	wg.Wait()
	log.Printf("[AdCrawler] 广告组 %s 抓取完成，共 %d 个广告", adgroupID, len(adIDs))
	return adIDs, nil
}

// CreativeCrawler 广告创意抓取器
type CreativeCrawler struct {
	*BaseCrawler
}

// NewCreativeCrawler 创建广告创意抓取器
func NewCreativeCrawler(db *gorm.DB, apiClient *APIClient) *CreativeCrawler {
	return &CreativeCrawler{
		BaseCrawler: NewBaseCrawler(db, apiClient),
	}
}

// CreativeItem 创意数据项
type CreativeItem struct {
	CreativeID      string                 `json:"creative_id"`
	AdID            string                 `json:"ad_id"`
	CreativeElements map[string]interface{} `json:"creative_elements"`
	PreviewURL      string                 `json:"preview_url"`
	Status          string                 `json:"status"`
}

// CreativeList 创意列表响应
type CreativeList struct {
	List     []CreativeItem `json:"list"`
	PageInfo PageInfo      `json:"page_info"`
}

// CrawlCreative 抓取广告创意
func (c *CreativeCrawler) Crawl(ctx context.Context, task model.CrawlTaskItem) error {
	log.Printf("[CreativeCrawler] 开始抓取账号 %s 的广告创意", task.AccountID)

	// 先获取所有广告
	ads, err := c.getAds(task.AccountID)
	if err != nil {
		return fmt.Errorf("获取广告失败: %w", err)
	}

	if len(ads) == 0 {
		log.Printf("[CreativeCrawler] 账号 %s 没有广告", task.AccountID)
		return nil
	}

	successCount := int64(0)
	failCount := int64(0)
	creativeIDs := make([]string, 0)
	var mu sync.Mutex
	var wg sync.WaitGroup

	// 最多5个并发
	semaphore := make(chan struct{}, 5)

	for _, ad := range ads {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		semaphore <- struct{}{}
		wg.Add(1)

		go func(adID string) {
			defer wg.Done()
			defer func() { <-semaphore }()

			page := 1
			pageSize := 500

			for {
				data, err := c.apiClient.GetCreatives(ctx, task.AccountID, adID, page, pageSize)
				if err != nil {
					log.Printf("[CreativeCrawler] 获取创意失败 (ad: %s): %v", adID, err)
					atomic.AddInt64(&failCount, int64(pageSize))
					break
				}

				var creativeList CreativeList
				if err := json.Unmarshal(data, &creativeList); err != nil {
					log.Printf("[CreativeCrawler] 解析创意失败: %v", err)
					atomic.AddInt64(&failCount, int64(pageSize))
					break
				}

				if len(creativeList.List) == 0 {
					break
				}

				for _, item := range creativeList.List {
					elementsJSON, _ := json.Marshal(item.CreativeElements)
					creative := model.AdCreative{
						CreativeID:   item.CreativeID,
						AccountID:   task.AccountID,
						CreativeName: adID, // 关联到广告ID
						ImageIDs:    string(elementsJSON),
						Title:       item.PreviewURL,
						Status:      c.parseStatus(item.Status),
						CreatedAt:  time.Now(),
					}

					if err := c.saveCreative(&creative); err != nil {
						log.Printf("[CreativeCrawler] 保存创意失败: %v", err)
						atomic.AddInt64(&failCount, 1)
						continue
					}

					mu.Lock()
					creativeIDs = append(creativeIDs, item.CreativeID)
					mu.Unlock()
					atomic.AddInt64(&successCount, 1)
				}

				if page >= creativeList.PageInfo.TotalPage {
					break
				}
				page++
			}
		}(ad)
	}

	wg.Wait()

	// 删除不再存在的创意
	if err := c.deleteNonExistentCreatives(task.AccountID, creativeIDs); err != nil {
		log.Printf("[CreativeCrawler] 删除失效创意失败: %v", err)
	}

	log.Printf("[CreativeCrawler] 账号 %s 抓取完成: 成功 %d, 失败 %d",
		task.AccountID, successCount, failCount)
	return nil
}

// getAds 获取账号下所有广告
func (c *CreativeCrawler) getAds(accountID string) ([]string, error) {
	var ads []model.Ad
	if err := c.db.Where("account_id = ?", accountID).Pluck("ad_id", &ads).Error; err != nil {
		return nil, err
	}

	ids := make([]string, len(ads))
	for i, ad := range ads {
		ids[i] = ad.AdID
	}
	return ids, nil
}

// parseStatus 解析状态
func (c *CreativeCrawler) parseStatus(status string) int {
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

// saveCreative 保存创意
func (c *CreativeCrawler) saveCreative(creative *model.AdCreative) error {
	return c.db.Where("account_id = ? AND creative_id = ?",
		creative.AccountID, creative.CreativeID).
		Assign(creative).
		FirstOrCreate(creative).Error
}

// deleteNonExistentCreatives 删除不再存在的创意
func (c *CreativeCrawler) deleteNonExistentCreatives(accountID string, existingIDs []string) error {
	if len(existingIDs) == 0 {
		return nil
	}

	return c.db.Where("account_id = ? AND creative_id NOT IN ?", accountID, existingIDs).
		Delete(&model.AdCreative{}).Error
}

// CrawlCreativesForAd 只抓取指定广告下的创意
func (c *CreativeCrawler) CrawlCreativesForAd(ctx context.Context, accountID, adID string) error {
	log.Printf("[CreativeCrawler] 开始抓取广告 %s 下的创意", adID)

	page := 1
	pageSize := 500

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		data, err := c.apiClient.GetCreatives(ctx, accountID, adID, page, pageSize)
		if err != nil {
			return fmt.Errorf("获取创意失败: %w", err)
		}

		var creativeList CreativeList
		if err := json.Unmarshal(data, &creativeList); err != nil {
			return fmt.Errorf("解析创意失败: %w", err)
		}

		if len(creativeList.List) == 0 {
			break
		}

		for _, item := range creativeList.List {
			elementsJSON, _ := json.Marshal(item.CreativeElements)
			creative := model.AdCreative{
				CreativeID:   item.CreativeID,
				AccountID:    accountID,
				CreativeName: adID,
				ImageIDs:     string(elementsJSON),
				Title:        item.PreviewURL,
				Status:       c.parseStatus(item.Status),
				CreatedAt:    time.Now(),
			}
			if err := c.saveCreative(&creative); err != nil {
				log.Printf("[CreativeCrawler] 保存创意失败: %v", err)
				continue
			}
		}

		if page >= creativeList.PageInfo.TotalPage {
			break
		}
		page++
	}

	log.Printf("[CreativeCrawler] 广告 %s 下创意抓取完成", adID)
	return nil
}
