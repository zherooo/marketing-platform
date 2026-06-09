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

// MaterialCrawler 广告素材抓取器
type MaterialCrawler struct {
	*BaseCrawler
}

// NewMaterialCrawler 创建广告素材抓取器
func NewMaterialCrawler(db *gorm.DB, apiClient *APIClient) *MaterialCrawler {
	return &MaterialCrawler{
		BaseCrawler: NewBaseCrawler(db, apiClient),
	}
}

// MaterialItem 素材数据项
type MaterialItem struct {
	MaterialID   string                 `json:"material_id"`
	MaterialType string                 `json:"material_type"`
	MaterialURL  string                 `json:"material_url"`
	Width        int                    `json:"width"`
	Height       int                    `json:"height"`
	FileSize     int64                  `json:"file_size"`
	Status       string                 `json:"status"`
	Labels       []string               `json:"labels"`
	Extra        map[string]interface{} `json:"extra"`
}

// MaterialList 素材列表响应
type MaterialList struct {
	List     []MaterialItem `json:"list"`
	PageInfo PageInfo       `json:"page_info"`
}

// MaterialType 素材类型常量
var MaterialTypes = []string{
	"IMAGE",      // 图片
	"VIDEO",      // 视频
	"CANVAS",     // 图文
	"MiniProgram", // 小程序
}

// CrawlMaterial 抓取广告素材
func (c *MaterialCrawler) Crawl(ctx context.Context, task model.CrawlTaskItem) error {
	log.Printf("[MaterialCrawler] 开始抓取账号 %s 的广告素材", task.AccountID)

	successCount := int64(0)
	failCount := int64(0)
	materialIDs := make([]string, 0)
	var mu sync.Mutex
	var wg sync.WaitGroup

	// 最多5个并发
	semaphore := make(chan struct{}, 5)

	for _, materialType := range MaterialTypes {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		semaphore <- struct{}{}
		wg.Add(1)

		go func(mType string) {
			defer wg.Done()
			defer func() { <-semaphore }()

			page := 1
			pageSize := 500

			for {
				data, err := c.apiClient.GetMaterials(ctx, task.AccountID, mType, page, pageSize)
				if err != nil {
					log.Printf("[MaterialCrawler] 获取素材失败 (type: %s): %v", mType, err)
					atomic.AddInt64(&failCount, int64(pageSize))
					break
				}

				var materialList MaterialList
				if err := json.Unmarshal(data, &materialList); err != nil {
					log.Printf("[MaterialCrawler] 解析素材失败: %v", err)
					atomic.AddInt64(&failCount, int64(pageSize))
					break
				}

				if len(materialList.List) == 0 {
					break
				}

				for _, item := range materialList.List {
					labelsJSON, _ := json.Marshal(item.Labels)
					
					material := model.AdMaterial{
						MaterialID:   item.MaterialID,
						AccountID:    task.AccountID,
						MaterialType: c.parseMaterialType(item.MaterialType),
						MaterialURL:  item.MaterialURL,
						Width:        item.Width,
						Height:       item.Height,
						FileSize:     item.FileSize,
						Signature:    string(labelsJSON),
						Status:       c.parseStatus(item.Status),
						CreatedAt:    time.Now(),
					}

					if err := c.saveMaterial(&material); err != nil {
						log.Printf("[MaterialCrawler] 保存素材失败: %v", err)
						atomic.AddInt64(&failCount, 1)
						continue
					}

					mu.Lock()
					materialIDs = append(materialIDs, item.MaterialID)
					mu.Unlock()
					atomic.AddInt64(&successCount, 1)
				}

				if page >= materialList.PageInfo.TotalPage {
					break
				}
				page++
			}
		}(materialType)
	}

	wg.Wait()

	// 删除不再存在的素材
	if err := c.deleteNonExistentMaterials(task.AccountID, materialIDs); err != nil {
		log.Printf("[MaterialCrawler] 删除失效素材失败: %v", err)
	}

	log.Printf("[MaterialCrawler] 账号 %s 抓取完成: 成功 %d, 失败 %d",
		task.AccountID, successCount, failCount)
	return nil
}

// parseMaterialType 解析素材类型
func (c *MaterialCrawler) parseMaterialType(materialType string) int {
	switch materialType {
	case "IMAGE":
		return 1
	case "VIDEO":
		return 2
	case "AUDIO":
		return 3
	default:
		return 0
	}
}

// parseStatus 解析状态
func (c *MaterialCrawler) parseStatus(status string) int {
	switch status {
	case "MATERIAL_STATUS_OK":
		return 1
	case "MATERIAL_STATUS_SUSPEND":
		return 2
	case "MATERIAL_STATUS_DELETE":
		return 3
	default:
		return 0
	}
}

// saveMaterial 保存素材
func (c *MaterialCrawler) saveMaterial(material *model.AdMaterial) error {
	return c.db.Where("account_id = ? AND material_id = ?",
		material.AccountID, material.MaterialID).
		Assign(material).
		FirstOrCreate(material).Error
}

// deleteNonExistentMaterials 删除不再存在的素材
func (c *MaterialCrawler) deleteNonExistentMaterials(accountID string, existingIDs []string) error {
	if len(existingIDs) == 0 {
		return nil
	}

	return c.db.Where("account_id = ? AND material_id NOT IN ?", accountID, existingIDs).
		Delete(&model.AdMaterial{}).Error
}
